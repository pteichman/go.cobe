package cobe

import (
	"log"
	"regexp"
	"strings"
)

type tokenizer interface {
	Split(string) []string
	Join([]string) string
}

type whitespaceTokenizer struct{}

func (t *whitespaceTokenizer) Split(str string) []string {
	return strings.Fields(str)
}

func (t *whitespaceTokenizer) Join(strs []string) string {
	return strings.Join(strs, " ")
}

// CobeTokenizer considers these to be tokens:
//
//  * one or more consecutive Unicode word characters (plus apostrophe
//    and dash)
//  * one or more consecutive Unicode non-word characters, possibly with
//    internal whitespace
//  * the whitespace between word or non-word tokens
//  * an HTTP url, [word]: followed by any run of non-space characters.
//
// This tokenizer collapses multiple spaces in a whitespace token into
// a single space character.
//
// It preserves differences in case. foo, Foo, and FOO are different
// tokens.
type cobeTokenizer struct {
	re *regexp.Regexp
}

func newCobeTokenizer() *cobeTokenizer {
	re := regexp.MustCompile(`(` +
		`\w+:\S+` +
		`|[\w'-]+` +
		`|[^\w\s][^\w]*[^\w\s]` +
		`|[^\w\s]` +
		`|\s+` +
		`)`)

	return &cobeTokenizer{re}
}

func (t *cobeTokenizer) Split(str string) []string {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return nil
	}

	tokens := t.re.FindAllString(str, -1)

	// Collapse runs of whitespace into a single space.
	for i, token := range tokens {
		if strings.TrimSpace(token) == "" {
			tokens[i] = " "
		}
	}

	return tokens
}

func (t *cobeTokenizer) Join(strs []string) string {
	return strings.Join(strs, "")
}

// A MegaHAL compatible tokenizer. Any of these are tokens:
//
//   * one or more consecutive alpha characters (plus apostrophe)
//   * one or more consecutive numeric characters
//   * one or more consecutive punctuation/space characters (not apostrophe)
//
// This tokenizer ignored differences in capitalization.
type megaHALTokenizer struct {
	re *regexp.Regexp
}

func newMegaHALTokenizer() *megaHALTokenizer {
	re := regexp.MustCompile(`([A-Z']+|[0-9]+|[^A-Z'0-9]+)`)

	return &megaHALTokenizer{re}
}

func (t *megaHALTokenizer) Split(str string) []string {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return nil
	}

	// Add ending punctuation if it is missing.
	if strings.IndexAny(str[len(str)-1:], ".!?") == -1 {
		str = str + "."
	}

	return t.re.FindAllString(strings.ToUpper(str), -1)
}

// Capitalize the first alpha character in the reply, along with the
// first alpha character that follows any of [.?!] and a space.
func (t *megaHALTokenizer) Join(strs []string) string {
	log.Fatal("Not implemented")
	return ""
}
