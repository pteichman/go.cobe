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

func TestStripAccents(t *testing.T) {
	var tests = []struct {
		text     string
		expected string
	}{
		{"Queensrÿche", "Queensryche"},
		{"Blue Öyster Cult", "Blue Oyster Cult"},
		{"Motörhead", "Motorhead"},
		{"The Accüsed", "The Accused"},
		{"Mötley Crüe", "Motley Crue"},
		{"François", "Francois"},
		{"ą/ę/ś/ć", "a/e/s/c"},
	}

	for ti, tt := range tests {
		strip := stripAccents(tt.text)
		if tt.expected != strip {
			t.Errorf("[%d] %s expected %s; was %s", ti, tt.text,
				tt.expected, strip)
		}
	}
}
