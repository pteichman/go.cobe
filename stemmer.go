package cobe

import (
	"regexp"
	"strings"

	"bitbucket.org/tebeka/snowball"
)

type stemmer interface {
	Stem(word string) string
}

// Wrap a snowball stemmer in one that also stems smileys.
type cobeStemmer struct {
	sub    stemmer
	words  *regexp.Regexp
	smiley *regexp.Regexp
	frowny *regexp.Regexp
}

func newCobeStemmer(s *snowball.Stemmer) *cobeStemmer {
	cs := cobeStemmer{sub: s}
	cs.words = regexp.MustCompile(`\w`)
	cs.smiley = regexp.MustCompile(`:-?[ \)]*\)|☺|☺️`)
	cs.frowny = regexp.MustCompile(`:-?[' \(]*\(`)

	return &cs
}

func (s *cobeStemmer) Stem(token string) string {
	// Tokens with a word character go through the snowball stemmer.
	if s.words.FindString(token) != "" {
		return s.sub.Stem(strings.ToLower(token))
	}

	if s.smiley.FindString(token) != "" {
		return ":)"
	}

	if s.frowny.FindString(token) != "" {
		return ":("
	}

	return ""
}
