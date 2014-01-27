package cobe

import "testing"
import "bitbucket.org/tebeka/snowball"

func TestCobeStemmer(t *testing.T) {
	snow, _ := snowball.New("english")
	s := newCobeStemmer(snow)

	// Straight port of the Python cobe stemmer.
	var tests = []struct {
		token    string
		expected string
	}{
		{"foo", "foo"},
		{"jumping", "jump"},
		{"running", "run"},

		{"Foo", "foo"},
		{"FOO", "foo"},
		{"FOO'S'", "foo"},
		{"FOOING", "foo"},
		{"Fooing", "foo"},

		{":)", ":)"},
		{":-)", ":)"},
		{":    )", ":)"},

		{":()", ":("},
		{":-(", ":("},
		{":    (", ":("},
		{":'    (", ":("},
	}

	for ti, tt := range tests {
		stem := s.Stem(tt.token)
		if tt.expected != stem {
			t.Errorf("[%d] %s\n%s !=\n%s", ti, tt.token, stem, tt.expected)
		}
	}
}
