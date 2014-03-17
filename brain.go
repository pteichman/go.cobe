package cobe

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Brain struct {
	graph  *graph
	tok    tokenizer
	scorer scorer
}

const spaceTokenID tokenID = -1

func OpenBrain(path string) (*Brain, error) {
	graph, err := openGraph(path)
	if err != nil {
		return nil, err
	}

	version, err := graph.getInfoString("version")
	if err != nil {
		return nil, err
	}

	if version != "2" {
		return nil, fmt.Errorf("cannot read version %s brain", version)
	}

	tokenizer, err := graph.getInfoString("tokenizer")
	if err != nil {
		return nil, err
	}

	return &Brain{graph, getTokenizer(tokenizer), &cobeScorer{}}, nil
}

func (b *Brain) Close() {
	if b.graph != nil {
		b.graph.close()
		b.graph = nil
	}
}

func getTokenizer(name string) tokenizer {
	switch strings.ToLower(name) {
	case "cobe":
		return newCobeTokenizer()
	case "megahal":
		return newMegaHALTokenizer()
	}

	return nil
}

func (b *Brain) Learn(text string) {
	now := time.Now()

	tokens := b.tok.Split(text)

	// skip learning if too few tokens (but don't count spaces)
	if countGoodTokens(tokens) <= b.graph.order {
		stats.Inc("learn.skip", 1, 1.0)
		return
	}

	stats.Inc("learn", 1, 1.0)

	var tokenIds []tokenID
	for _, text := range tokens {
		var tokenID tokenID
		if text == " " {
			tokenID = spaceTokenID
		} else {
			tokenID = b.graph.getOrCreateToken(text)
		}

		tokenIds = append(tokenIds, tokenID)
	}

	var prevNode nodeID
	b.forEdges(tokenIds, func(prev, next []tokenID, hasSpace bool) {
		if prevNode == 0 {
			prevNode = b.graph.getOrCreateNode(prev)
		}
		nextNode := b.graph.getOrCreateNode(next)

		b.graph.addEdge(prevNode, nextNode, hasSpace)
		prevNode = nextNode
	})

	stats.Timing("learn.response_time", int64(time.Since(now)/time.Millisecond), 1.0)
}

func countGoodTokens(tokens []string) int {
	var count int
	for _, token := range tokens {
		if token != " " {
			count++
		}
	}

	return count
}

func (b *Brain) forEdges(tokenIds []tokenID, f func([]tokenID, []tokenID, bool)) {
	// Call f() on every N-gram (N = brain order) in tokenIds.
	order := b.graph.order

	chain := b.toChain(order, tokenIds)
	edges := toEdges(order, chain)

	for _, e := range edges {
		f(e.prev, e.next, e.hasSpace)
	}
}

func (b *Brain) toChain(order int, tokenIds []tokenID) []tokenID {
	var chain []tokenID
	for i := 0; i < order; i++ {
		chain = append(chain, b.graph.endTokenID)
	}

	chain = append(chain, tokenIds...)

	for i := 0; i < order; i++ {
		chain = append(chain, b.graph.endTokenID)
	}

	return chain
}

type edge struct {
	prev     []tokenID
	next     []tokenID
	hasSpace bool
}

func toEdges(order int, tokenIds []tokenID) []edge {
	var tokens []tokenID
	var spaces []int

	// Turn tokenIds (containing some SPACE_TOKEN_ID) into a list
	// of tokens and a list of positions in the tokens slice after
	// which spaces were found.

	for i := 0; i < len(tokenIds); i++ {
		tokens = append(tokens, tokenIds[i])

		if i < len(tokenIds)-1 && tokenIds[i+1] == spaceTokenID {
			spaces = append(spaces, len(tokens))
			i++
		}
	}

	var ret []edge

	prev := tokens[0:order]
	for i := 1; i < len(tokens)-order+1; i++ {
		next := tokens[i : i+order]

		var hasSpace bool
		if len(spaces) > 0 && spaces[0] == i+order-1 {
			hasSpace = true
			spaces = spaces[1:]
		}

		ret = append(ret, edge{prev, next, hasSpace})
		prev = next
	}

	return ret
}

func (b *Brain) Reply(text string) string {
	now := time.Now()
	stats.Inc("reply", 1, 1.0)

	tokens := b.tok.Split(text)
	tokenIds := b.graph.filterPivots(unique(tokens))

	stemTokenIds := b.conflateStems(tokens)
	tokenIds = uniqueIds(append(tokenIds, stemTokenIds...))

	if len(tokenIds) == 0 {
		tokenIds = b.babble()
	}

	if len(tokenIds) == 0 {
		stats.Inc("error", 1, 1.0)
		return "I don't know enough to answer you yet!"
	}

	var count int

	var bestReply *reply
	var bestScore float64 = -1

	stop := make(chan bool)
	replies := b.replySearch(tokenIds, stop)

	timeout := time.After(500 * time.Millisecond)
loop:
	for {
		select {
		case edges := <-replies:
			if edges == nil {
				// Channel was closed: run another search
				replies = b.replySearch(tokenIds, stop)
				continue loop
			}

			reply := newReply(b.graph, edges)
			score := b.scorer.Score(reply)

			if score > bestScore {
				bestReply = reply
				bestScore = score
			}

			count++
		case <-timeout:
			if bestReply != nil {
				break loop
			} else {
				timeout = time.After(500 * time.Millisecond)
			}
		}
	}

	// Tell replies to stop and block until we're sure it has closed.
	close(stop)
	if _, ok := <-replies; ok {
		// Replies got unexpected results after search stop.
		stats.Inc("error", 1, 1.0)
	}

	stats.Inc("reply.candidate", int64(count), 1.0)

	clog.Info("Got %d total replies\n", count)
	if bestReply == nil {
		return "I don't know enough to answer you yet!"
	}

	ret := bestReply.ToString()
	stats.Timing("reply.response_time", int64(time.Since(now)/time.Millisecond), 1.0)
	return ret
}

func (b *Brain) conflateStems(tokens []string) []tokenID {
	var ret []tokenID

	for _, token := range tokens {
		tokenIds := b.graph.getTokensByStem(token)
		ret = append(ret, tokenIds...)
	}

	return ret
}

func (b *Brain) babble() []tokenID {
	var tokenIds []tokenID

	for i := 0; i < 5; i++ {
		t := b.graph.getRandomToken()
		if t > 0 {
			tokenIds = append(tokenIds, tokenID(t))
		}
	}

	return tokenIds
}

// replySearch combines a forward and a reverse search over the graph
// into a series of replies.
func (b *Brain) replySearch(tokenIds []tokenID, stop <-chan bool) <-chan []edgeID {
	pivotID := b.pickPivot(tokenIds)
	pivotNode := b.graph.getRandomNodeWithToken(pivotID)

	endNode := b.graph.endContextID

	revIter := &history{b.graph.search(pivotNode, endNode, reverse, stop), nil}
	fwdIter := &history{b.graph.search(pivotNode, endNode, forward, stop), nil}

	replies := make(chan []edgeID)

	go func() {
	loop:
		for {
			rev := revIter.next()
			if rev {
				// combine new rev with all fwds
				result := revIter.result()
				for _, f := range fwdIter.h {
					select {
					case replies <- join(result, f):
						// nothing
					case <-stop:
						break loop
					}
				}
			}

			fwd := fwdIter.next()
			if fwd {
				// combine new fwd with all revs
				result := fwdIter.result()
				for _, r := range revIter.h {
					select {
					case replies <- join(r, result):
						// nothing
					case <-stop:
						break loop
					}
				}
			}

			if !rev && !fwd {
				break
			}
		}

		close(replies)
	}()

	return replies
}

type history struct {
	s *search
	h [][]edgeID
}

func (h *history) next() bool {
	ret := h.s.next()
	if ret {
		h.h = append(h.h, h.s.result)
	}

	return ret
}

func (h *history) result() []edgeID {
	return h.s.result
}

func join(rev []edgeID, fwd []edgeID) []edgeID {
	edges := make([]edgeID, 0, len(rev)+len(fwd))

	// rev is a path from the pivot node to the beginning of a
	// reply: join its edges in reverse order.
	for i := len(rev) - 1; i >= 0; i-- {
		edges = append(edges, rev[i])
	}

	return append(edges, fwd...)
}

func (b *Brain) pickPivot(tokenIds []tokenID) tokenID {
	return tokenIds[rand.Intn(len(tokenIds))]
}

func unique(tokens []string) []string {
	// Reduce tokens to a unique set by sending them through a map.
	m := make(map[string]int)
	for _, token := range tokens {
		m[token]++
	}

	ret := make([]string, 0, len(m))
	for token := range m {
		ret = append(ret, token)
	}

	return ret
}

func uniqueIds(ids []tokenID) []tokenID {
	// Reduce token ids to a unique set by sending them through a map.
	m := make(map[tokenID]int)
	for _, id := range ids {
		m[id]++
	}

	ret := make([]tokenID, 0, len(m))
	for id := range m {
		ret = append(ret, id)
	}

	return ret
}

type reply struct {
	graph   *graph
	edges   []edgeID
	hasText bool
	text    string
}

func newReply(graph *graph, edges []edgeID) *reply {
	return &reply{graph, edges, false, ""}
}

func (r *reply) ToString() string {
	if !r.hasText {
		var parts []string

		// Skip any edges that don't contain word nodes.
		wordEdges := r.edges[1 : len(r.edges)-r.graph.order+1]

		for _, edge := range wordEdges {
			word, hasSpace, err := r.graph.getTextByEdge(edge)
			if err != nil {
				stats.Inc("error", 1, 1.0)
				clog.Error("can't get text", err)
			}

			if word == "" {
				stats.Inc("error", 1, 1.0)
				clog.Error("empty node text! %s", r.edges)
			}

			parts = append(parts, word)
			if hasSpace {
				parts = append(parts, " ")
			}
		}

		r.hasText = true
		r.text = strings.Join(parts, "")
	}

	return r.text
}
