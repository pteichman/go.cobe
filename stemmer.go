package cobe

import (
	"regexp"
	"strings"
	"unicode"

	"bitbucket.org/tebeka/snowball"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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
	cs.frowny = regexp.MustCompile(`:-?[' \(]*\(|☹|😦`)

	return &cs
}

func (s *cobeStemmer) Stem(token string) string {
	// Tokens with a word character go through the snowball stemmer.
	if s.words.FindString(token) != "" {
		return s.sub.Stem(stripAccents(strings.ToLower(token)))
	}

	if s.smiley.FindString(token) != "" {
		return ":)"
	}

	if s.frowny.FindString(token) != "" {
		return ":("
	}

	return ""
}

// stripAccents attempts to replace accented characters with an ASCII
// equivalent. This is an extreme oversimplication, but since cobe
// only uses this to create token equivalence (these strings are never
// displayed) it gets a pass.
func stripAccents(s string) string {
	s2, _, err := transform.String(stripT, s)
	if err != nil {
		return s
	}

	return s2
}

var stripT transform.Transformer

func init() {
	stripT = transform.Chain(
		norm.NFD,
		transform.RemoveFunc(isMn),
		norm.NFC)
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
