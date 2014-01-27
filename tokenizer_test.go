package cobe

import "testing"

func eq(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestCobeTokenizer(t *testing.T) {
	tok := newCobeTokenizer()

	// Straight port of the Python cobe tokenizer.
	var tests = []struct {
		str      string
		expected []string
	}{
		{"", []string{}},
		{" ", []string{}},
		{"hi.", []string{"hi", "."}},
		{"hi, cobe", []string{"hi", ",", " ", "cobe"}},
		{"hi - cobe", []string{"hi", " ", "-", " ", "cobe"}},
		{"hi  -  cobe", []string{"hi", " ", "-", " ", "cobe"}},
		{"-foo", []string{"-foo"}},

		{"foo", []string{"foo"}},
		{" foo", []string{"foo"}},
		{"  foo", []string{"foo"}},
		{"foo ", []string{"foo"}},
		{"foo  ", []string{"foo"}},

		{":)", []string{":)"}},
		{";)", []string{";)"}},
		{":(", []string{":("}},
		{";(", []string{";("}},

		{"http://www.google.com/", []string{"http://www.google.com/"}},
		{"https://www.google.com/", []string{"https://www.google.com/"}},
		{"cobe://www.google.com/", []string{"cobe://www.google.com/"}},
		{"cobe:www.google.com/", []string{"cobe:www.google.com/"}},
		{":foo", []string{":", "foo"}},

		{"testing :    (", []string{"testing", " ", ":    ("}},
		{"testing          :    (", []string{"testing", " ", ":    ("}},
		{"testing          :    (  foo", []string{"testing", " ", ":    (", " ", "foo"}},

		{"test-ing", []string{"test-ing"}},
		{":-)", []string{":-)"}},
		{"test-ing :-) 1-2-3", []string{"test-ing", " ", ":-)", " ", "1-2-3"}},

		{"don't :'(", []string{"don't", " ", ":'("}},
	}

	for ti, tt := range tests {
		tokens := tok.Split(tt.str)
		if !eq(tokens, tt.expected) {
			t.Errorf("[%d] %s\n%s !=\n%s", ti, tt.str, tokens, tt.expected)
		}
	}
}

func TestMegaHALTokenizer(t *testing.T) {
	tok := newMegaHALTokenizer()

	// Straight port of the Python MegaHAL tokenizer.
	var tests = []struct {
		str      string
		expected []string
	}{
		{"", []string{}},
		{" ", []string{}},
		{"hi.", []string{"HI", "."}},
		{"hi, cobe.", []string{"HI", ", ", "COBE", "."}},
		{"hi", []string{"HI", "."}},

		{"http://www.google.com/", []string{"HTTP", "://", "WWW", ".", "GOOGLE", ".", "COM", "/."}},

		{"hal's brain", []string{"HAL'S", " ", "BRAIN", "."}},
		{"',','", []string{"'", ",", "'", ",", "'", "."}},

		{"hal9000, test blah 12312", []string{"HAL", "9000", ", ", "TEST", " ", "BLAH", " ", "12312", "."}},
		{"hal9000's test", []string{"HAL", "9000", "'S", " ", "TEST", "."}},
	}

	for ti, tt := range tests {
		tokens := tok.Split(tt.str)
		if !eq(tokens, tt.expected) {
			t.Errorf("[%d] %s\n%s !=\n%s", ti, tt.str, tokens, tt.expected)
		}
	}
}
