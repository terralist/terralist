package matcher

import (
	"strconv"
	"testing"
)

var testData = []struct {
	text     string
	pattern  string
	expected bool
}{
	// Algorithm validation
	{"baaabab", "ba*****ab", true},
	{"baaabab", "ba*****ab", true},
	{"bab", "ba*ab", false},
	{"aab", "a*ab", true},
	{"aab", "a*****ab", true},
	{"", "*a*****ab", false},
	{"babab", "ba*ab****", true},
	{"anything", "****", true},
	{"anything", "*", true},
	{"", "****", true},
	{"", "?*", false},
	{"aacab", "aa?ab", true},
	{"bb", "b*b", true},
	{"bb", "a*a", false},
	{"baaabab", "baaabab", true},
	{"baaabaa", "baaabab", false},
	{"zbaaabab", "?baaabab", true},
	{"baaabab", "?baaabab", false},
	{"", "*baaaba*", false},
	{"baaaba", "*baaaba*", true},
	{"abcbc", "a*bc", true},
	{"baaaba:bbaa", "baaaba:bb*", true},
	{"baaaba:bbaa", "baaaba:*", true},
	{"b:", "b:*", true},
	{"b:", "b:?", false},
	{"b:", "b:?*", false},
	{"b:a", "b:?*", true},
	{"b:aa", "b:?*", true},
	{"aaa-ccc-ddd-bbb", "aaa-*-bbb", true},
	{"aaa--bbb", "aaa-*-bbb", true},
	{"aaa-ccc-ddd", "aaa-*-bbb", false},

	// Real cases
	{"modules:get", "modules:*", true},
	{"modules:post", "modules:p*", true},
	{"modules:put", "modules:p*", true},
	{"modules:get", "modules:p*", false},
	{"modules:delete", "modules:p*", false},
	{"modules:delete", "*", true},
	{"provider:aws", "provider:aws", true},
	{"provider:aws", "provider:google", false},
	{"provider:google", "provider:*", true},
}

func TestMatch(t *testing.T) {
	for _, td := range testData {
		result := Match(td.text, td.pattern)

		if result != td.expected {
			t.Errorf(
				"Match(%s, %s) = %s, expected %s",
				td.text,
				td.pattern,
				strconv.FormatBool(result),
				strconv.FormatBool(td.expected),
			)
		}
	}
}
