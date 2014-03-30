package cobe

import "math"

type scorer interface {
	Score(reply *Reply) float64
}

type cobeScorer struct{}

func (s *cobeScorer) Score(reply *Reply) float64 {
	var info float64
	g := reply.graph

	// Calculate the information content of the edges in this reply.
	nodes := reply.nodes
	for i := 0; i < len(nodes)-1; i++ {
		info -= g.getEdgeLogprob(nodes[i], nodes[i+1])
	}

	// Apply MegaHAL's fudge factor to discourage overly long
	// replies.

	// First, we have (graph.order - 1) extra edges on either end
	// of the reply, cobe 2.0 learns from (endToken, endToken,
	// ...).
	nWords := len(reply.nodes) - (g.order-1)*2

	if nWords > 16 {
		info /= math.Sqrt(float64(nWords - 1))
	} else if nWords >= 32 {
		info /= float64(nWords)
	}

	return info
}
