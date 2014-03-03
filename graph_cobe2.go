package cobe

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

import (
	"bitbucket.org/tebeka/snowball"
	_ "github.com/mattn/go-sqlite3"
	"github.com/phf/go-queue/queue"
)

// This is a straight port of the Python cobe brain.

type tokenID int64
type nodeID int64
type edgeID int64

type Direction int

const (
	Forward Direction = iota
	Reverse
)

type graph struct {
	db *sql.DB
	q  *stmts

	stemmer stemmer

	endTokenID   tokenID
	endContextID nodeID
}

type stmts struct {
	selectInfo *sql.Stmt
	insertInfo *sql.Stmt
	updateInfo *sql.Stmt
	deleteInfo *sql.Stmt

	selectToken *sql.Stmt
	insertToken *sql.Stmt

	selectNode *sql.Stmt
	insertNode *sql.Stmt

	incrEdge   *sql.Stmt
	insertEdge *sql.Stmt

	selectEdgeText    *sql.Stmt
	selectEdgeCounts  *sql.Stmt
	selectRandomToken *sql.Stmt
	selectRandomNode  *sql.Stmt

	insertStem       *sql.Stmt
	selectStemTokens *sql.Stmt
}

func openGraph(path string) (*graph, error) {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		err = InitGraph(path, defaultGraphOptions)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("file:%s?cache=shared&mode=rwc", path)

	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, err
	}

	err = pragmas(db)
	if err != nil {
		return nil, err
	}

	stmts := new(stmts)
	err = prepareInfoSql(db, stmts)
	if err != nil {
		return nil, err
	}

	g := &graph{db: db, q: stmts}

	err = prepareSql(db, stmts, g.getOrder())
	if err != nil {
		return nil, err
	}

	lang, err := g.GetInfoString("stemmer")
	if lang != "" {
		s, err := snowball.New(lang)
		if err != nil {
			log.Printf("Error initializing stemmer: %s", err)
		} else {
			g.stemmer = newCobeStemmer(s)
		}
	}

	g.endTokenID = g.GetOrCreateToken("")
	g.endContextID = g.GetOrCreateNode(g.endContext())

	return g, nil
}

func (g *graph) Close() {
	if g.db != nil {
		g.db.Close()
		g.db = nil
	}
}

func (g *graph) getOrder() int {
	str, err := g.GetInfoString("order")
	if err != nil {
		log.Println(err)
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		log.Println(err)
	}

	return val
}

func (g *graph) endContext() []tokenID {
	return []tokenID(repeat(g.getOrder(), g.endTokenID))
}

func repeat(n int, id tokenID) []tokenID {
	ret := make([]tokenID, n)
	for i := 0; i < n; i++ {
		ret[i] = id
	}

	return ret
}

func pragmas(db *sql.DB) error {
	// Disable the SQLite cache. Its pages tend to get swapped
	// out, even if the database file is in buffer cache.
	err := exec0(db, "PRAGMA cache_size=0")
	if err != nil {
		return err
	}

	err = exec0(db, "PRAGMA page_size=4096")
	if err != nil {
		return err
	}

	// Make speed-for-reliability tradeoffs that improve bulk
	// learning.
	err = exec0(db, "PRAGMA journal_mode=truncate")
	if err != nil {
		return err
	}

	err = exec0(db, "PRAGMA temp_store=memory")
	if err != nil {
		return err
	}

	err = exec0(db, "PRAGMA synchronous=OFF")
	if err != nil {
		return err
	}

	return nil
}

func prepareInfoSql(db *sql.DB, stmts *stmts) error {
	var err error

	stmts.selectInfo, err = db.Prepare(
		"SELECT text FROM info WHERE attribute = ?")
	if err != nil {
		return err
	}

	stmts.insertInfo, err = db.Prepare(
		"INSERT INTO info (attribute, text) VALUES (?, ?)")
	if err != nil {
		return err
	}

	stmts.updateInfo, err = db.Prepare(
		"UPDATE info SET text = ? WHERE attribute = ?")
	if err != nil {
		return err
	}

	stmts.deleteInfo, err = db.Prepare(
		"DELETE FROM info WHERE attribute = ?")
	if err != nil {
		return err
	}

	return nil
}

func prepareSql(db *sql.DB, stmts *stmts, order int) error {
	var err error

	stmts.selectToken, err = db.Prepare(
		"SELECT id FROM tokens WHERE text = ?")
	if err != nil {
		return err
	}

	stmts.insertToken, err = db.Prepare(
		"INSERT INTO tokens (text, is_word) VALUES (?, ?)")
	if err != nil {
		return err
	}

	args := nStrings(order, func(i int) string {
		return fmt.Sprintf("token%d_id = ?", i)
	})

	query := fmt.Sprintf("SELECT id FROM nodes WHERE %s",
		strings.Join(args, " AND "))

	stmts.selectNode, err = db.Prepare(query)
	if err != nil {
		return err
	}

	allTokens := nStrings(order, func(i int) string {
		return fmt.Sprintf("token%d_id", i)
	})

	allQ := nStrings(order, func(i int) string { return "?" })

	query = fmt.Sprintf(
		"INSERT INTO nodes (count, %s) VALUES (0, %s)",
		strings.Join(allTokens, ", "), strings.Join(allQ, ", "))

	stmts.insertNode, err = db.Prepare(query)
	if err != nil {
		return err
	}

	stmts.incrEdge, err = db.Prepare("UPDATE edges SET count = count + 1 " +
		"WHERE prev_node = ? AND next_node = ? AND has_space = ?")
	if err != nil {
		return err
	}

	stmts.insertEdge, err = db.Prepare(
		"INSERT INTO EDGES (prev_node, next_node, has_space, count) " +
			"VALUES (?, ?, ?, 1)")
	if err != nil {
		return err
	}

	query = fmt.Sprintf("SELECT tokens.text, edges.has_space "+
		"FROM nodes, edges, tokens "+
		"WHERE edges.id = ? AND edges.prev_node = nodes.id "+
		"AND nodes.token%d_id = tokens.id", order-1)

	stmts.selectEdgeText, err = db.Prepare(query)
	if err != nil {
		return err
	}

	stmts.selectEdgeCounts, err = db.Prepare("SELECT edges.count, nodes.count " +
		"FROM edges, nodes " +
		"WHERE edges.id = ? AND edges.prev_node = nodes.id")
	if err != nil {
		return err
	}

	// Generate a random known token from 2..max(id)
	// inclusive. Token 1 is endTokenId, so we skip it.
	stmts.selectRandomToken, err = db.Prepare(
		"SELECT (abs(random()) % (MAX(id)-1)) + 2 FROM tokens")
	if err != nil {
		return err
	}

	stmts.selectRandomNode, err = db.Prepare("SELECT id " +
		"FROM nodes WHERE token0_id = ? " +
		"LIMIT 1 OFFSET abs(random())%(SELECT count(*) FROM nodes " +
		"                              WHERE token0_id = ?)")
	if err != nil {
		return err
	}

	stmts.insertStem, err = db.Prepare(
		"INSERT INTO token_stems (token_id, stem) " +
			"VALUES (?, ?)")
	if err != nil {
		return err
	}

	stmts.selectStemTokens, err = db.Prepare("SELECT token_id " +
		"FROM token_stems WHERE token_stems.stem = ?")
	if err != nil {
		return err
	}

	return nil
}

func nStrings(n int, f func(int) string) []string {
	var ret = make([]string, n)
	for i := 0; i < n; i++ {
		ret[i] = f(i)
	}

	return ret
}

func exec0(db *sql.DB, query string) error {
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return nil
}

func (g *graph) GetInfoString(key string) (string, error) {
	var value string

	err := g.q.selectInfo.QueryRow(key).Scan(&value)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (g *graph) DelInfoString(key string) error {
	_, err := g.q.deleteInfo.Exec(key)
	return err
}

func (g *graph) SetInfoString(key, value string) error {
	res, err := g.q.updateInfo.Exec(key, value)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows > 0 {
		return nil
	}

	res, err = g.q.insertInfo.Exec(key, value)
	if err != nil {
		return err
	}

	return nil
}

func (g *graph) GetTokenID(text string) (tokenID, error) {
	var value int64

	err := g.q.selectToken.QueryRow(text).Scan(&value)
	if err != nil {
		return -1, err
	}

	return tokenID(value), nil
}

func (g *graph) filterPivots(tokens []string) []tokenID {
	known := g.getKnownTokenIds(tokens)

	words := g.filterWordTokenIds(known)
	if len(words) > 0 {
		return known
	}

	return known
}

func (g *graph) getKnownTokenIds(tokens []string) []tokenID {
	var ret []tokenID

	for _, token := range tokens {
		id, err := g.GetTokenID(token)
		if err == nil {
			ret = append(ret, tokenID(id))
		}
	}

	return ret
}

func (g *graph) filterWordTokenIds(tokenIds []tokenID) []tokenID {
	query := fmt.Sprintf(
		"SELECT id FROM tokens WHERE id IN (%s) AND is_word = 1",
		seqQ(len(tokenIds)))

	return g.filterTokens(query, tokenIds)
}

func (g *graph) filterTokens(query string, tokenIds []tokenID) []tokenID {
	rows, err := g.db.Query(query, toQueryArgs(tokenIds)...)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	var ret []tokenID
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ret = append(ret, tokenID(id))
	}

	return ret
}

func seqQ(n int) string {
	return strings.Join(
		nStrings(n, func(n int) string { return "?" }), ", ")
}

func (g *graph) GetOrCreateToken(text string) tokenID {
	token, err := g.GetTokenID(text)
	if err == nil {
		return token
	}

	var isWordRegexp = regexp.MustCompile(`\w`)
	isWord := isWordRegexp.FindStringIndex(text) != nil

	res, err := g.q.insertToken.Exec(text, isWord)
	if err != nil {
		return -1
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1
	}

	tokenID := tokenID(id)

	if g.stemmer != nil {
		stem := g.stemmer.Stem(text)
		if stem != "" {
			g.q.insertStem.Exec(tokenID, stem)
		}
	}

	return tokenID
}

func toQueryArgs(tokenIds []tokenID) []interface{} {
	ret := make([]interface{}, 0, len(tokenIds))
	for _, tokenID := range tokenIds {
		ret = append(ret, int64(tokenID))
	}

	return ret
}

func (g *graph) GetOrCreateNode(tokens []tokenID) nodeID {
	var node int64

	tokenIds := toQueryArgs(tokens)

	err := g.q.selectNode.QueryRow(tokenIds...).Scan(&node)
	if err == nil {
		return nodeID(node)
	}

	res, err := g.q.insertNode.Exec(tokenIds...)
	if err != nil {
		log.Println(err)
	}

	node, err = res.LastInsertId()
	if err != nil {
		log.Println(err)
	}

	return nodeID(node)
}

func (g *graph) addEdge(prev nodeID, next nodeID, hasSpace bool) {
	res, err := g.q.incrEdge.Exec(prev, next, hasSpace)
	if err != nil {
		log.Println(err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	if n == 0 {
		_, err := g.q.insertEdge.Exec(prev, next, hasSpace)
		if err != nil {
			log.Println(err)
		}
	}

	// The count on the next_node in the nodes table is
	// incremented here with database triggers. This registers
	// that the node has been seen an additional time (used by
	// scoring).
}

func (g *graph) getTextByEdge(edgeID edgeID) (string, bool, error) {
	var text string
	var hasSpace bool

	err := g.q.selectEdgeText.QueryRow(edgeID).Scan(&text, &hasSpace)
	if err != nil {
		return "", false, err
	}

	return text, hasSpace, nil
}

func (g *graph) getRandomToken() tokenID {
	var token int64
	g.q.selectRandomToken.QueryRow().Scan(&token)

	return tokenID(token)
}

func (g *graph) getRandomNodeWithToken(t tokenID) nodeID {
	var node int64
	g.q.selectRandomNode.QueryRow(t, t).Scan(&node)

	return nodeID(node)
}

func (g *graph) getTokensByStem(stem string) []tokenID {
	var ret []tokenID

	if g.stemmer == nil {
		return ret
	}

	rows, err := g.q.selectStemTokens.Query(g.stemmer.Stem(stem))
	if err != nil {
		log.Printf("ERROR: %s", err)
		return ret
	}

	var t int64
	for rows.Next() {
		rows.Scan(&t)
		ret = append(ret, tokenID(t))
	}

	return ret
}

func (g *graph) getEdgeLogprob(edgeID edgeID) float64 {
	// Each edges goes from an n-gram node (word1, word2, word3)
	// to another (word2, word3, word4).
	//
	// P(word4|word1, word2, word3) = count(edgeId) / count(prevNodeId)
	//
	var edgeCount, prevNodeCount int64
	g.q.selectEdgeCounts.QueryRow(edgeID).Scan(&edgeCount, &prevNodeCount)
	return math.Log2(float64(edgeCount)) - math.Log2(float64(prevNodeCount))
}

type node struct {
	node nodeID
	path []edgeID
}

type search struct {
	follow func(node nodeID) (*sql.Rows, error)
	end    nodeID
	left   *queue.Queue
	result []edgeID
}

func (s *search) Next() bool {
	for s.left.Len() > 0 {
		cur := s.left.PopFront().(*node)
		if cur.node == s.end {
			s.result = cur.path
			return true
		}

		rows, _ := s.follow(cur.node)
		for rows.Next() {
			var e, n int64
			rows.Scan(&e, &n)
			path := append(cur.path[:], edgeID(e))
			s.left.PushBack(&node{nodeID(n), path})
		}
	}

	s.result = nil
	return false
}

func (s *search) Result() []edgeID {
	return s.result
}

func (g *graph) search(start nodeID, end nodeID, dir Direction) *search {
	var q string
	if dir == Forward {
		q = "SELECT id, next_node FROM edges WHERE prev_node = ?"
	} else {
		q = "SELECT id, prev_node FROM edges WHERE next_node = ?"
	}

	follow := func(node nodeID) (*sql.Rows, error) {
		return g.db.Query(q, node)
	}

	left := queue.New()
	left.PushBack(&node{start, nil})

	return &search{follow, end, left, nil}
}
