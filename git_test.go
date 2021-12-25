package main

import "testing"

func TestCleanupURL(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{"git@github.com:muesli/gitty.git", "https://github.com/muesli/gitty"},
		{"git://github.com/muesli/gitty.git", "https://github.com/muesli/gitty"},
		{"http://github.com/muesli/gitty.git", "https://github.com/muesli/gitty"},
		{"https://github.com/muesli/gitty", "https://github.com/muesli/gitty"},
		{"ssh://git@git.domain.tld:2222/muesli/gitty.git", "https://git.domain.tld/muesli/gitty"},
	}

	for _, test := range tests {
		r, err := cleanupURL(test.input)
		if err != nil {
			t.Errorf("Error: %s", err)
		}
		if r != test.expected {
			t.Errorf("CleanupURL(%s) %s != %s", test.input, r, test.expected)
		}
	}
}
