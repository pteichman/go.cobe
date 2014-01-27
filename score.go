package cobe

import "math"

type scorer interface {
	Score(reply *reply) float64
}

type cobeScorer struct{}

func (s *cobeScorer) Score(reply *reply) float64 {
	var info float64
	g := reply.graph

	// Calculate the information content of the edges in this reply.
	for _, edge := range reply.edges {
		info -= g.getEdgeLogprob(edge)
	}

	// Apply MegaHAL's fudge factor to discourage overly long
	// replies.

	// First, we have (graph.order - 1) extra edges on either end
	// of the reply, cobe 2.0 learns from (endToken, endToken,
	// ...).
	nWords := len(reply.edges) - (g.getOrder()-1)*2

	if nWords > 16 {
		info /= math.Sqrt(float64(nWords - 1))
	} else if nWords >= 32 {
		info /= float64(nWords)
	}

	return info
}
