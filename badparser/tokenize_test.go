package badparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizeLine(t *testing.T) {
	testCases := []struct {
		line     string
		expected []string
	}{
		{
			line:     "a b",
			expected: []string{"a", "b"},
		},
		{
			line:     "a  b",
			expected: []string{"a", "b"},
		},
		{
			line:     "a [b]",
			expected: []string{"a", "[", "b", "]"},
		},
		{
			line:     `a [foo="bar"]`,
			expected: []string{"a", "[", "foo", "=", "bar", "]"},
		},
		{
			line:     `foo"bar"baz`,
			expected: []string{"foo", "bar", "baz"},
		},
		{
			line:     `bindsym $mod+f fullscreen toggle`,
			expected: []string{"bindsym", "$mod+f", "fullscreen", "toggle"},
		},
		{
			line:     `a;b`,
			expected: []string{"a", ";", "b"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.line, func(t *testing.T) {
			result := tokenizeLine(tc.line)
			assert.Equal(t, tc.expected, result)
		})
	}
}
