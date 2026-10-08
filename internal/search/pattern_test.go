package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsPattern(t *testing.T) {
	for name, tc := range map[string]struct {
		text string
		want string
	}{
		"empty is no filter":                 {text: "", want: ""},
		"whitespace is no filter":            {text: "   ", want: ""},
		"partial name":                       {text: "ad", want: "%ad%"},
		"padding is trimmed":                 {text: "  ad  ", want: "%ad%"},
		"percent is escaped, not a wildcard": {text: "50%", want: `%50\%%`},
		"underscore is escaped":              {text: "a_b", want: `%a\_b%`},
		"backslash is escaped first":         {text: `a\b`, want: `%a\\b%`},
		"persian passes through":             {text: "سارا", want: "%سارا%"},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, ContainsPattern(tc.text))
		})
	}
}
