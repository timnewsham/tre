package tre

import (
	"testing"

	"github.com/alecthomas/assert"
)

type matchFunc func(t *testing.T, pat, s string) bool

func matchNfa(t *testing.T, pat, s string) bool {
	t.Helper()
	p, err := Parse(pat)
	assert.NoError(t, err)
	nfa := MakeNfa(p)
	return MatchNfa(nfa, s)
}

func expectMatch(t *testing.T, match matchFunc, pat, s string) {
	t.Helper()
	m := match(t, pat, s)
	assert.True(t, m)
}

func expectNoMatch(t *testing.T, match matchFunc, pat, s string) {
	t.Helper()
	m := match(t, pat, s)
	assert.False(t, m)
}

func TestRe(t *testing.T) {
	m := matchNfa
	expectMatch(t, m, "a", "a")
	expectMatch(t, m, "[a-z]", "a")
	expectNoMatch(t, m, "[a-z]", "X")
	expectMatch(t, m, "[^a-z]", "X")
	expectNoMatch(t, m, "[^a-z]", "a")
	expectMatch(t, m, "[axz]", "a")
	expectMatch(t, m, "[axz]", "x")
	expectMatch(t, m, "[axz]", "z")
	expectNoMatch(t, m, "[axz]", "y")
	expectMatch(t, m, "hello", "hello")
	expectNoMatch(t, m, "hello", "hell")
	expectNoMatch(t, m, "hello", "helloo")
	expectMatch(t, m, "(hello)|(goodbye)", "hello")
	expectMatch(t, m, "(hello)|(goodbye)", "goodbye")
	expectMatch(t, m, "(hello|goodbye)", "hello")
	// XXX
	//expectMatch(t, m, "(hello|goodbye)", "goodbye")
	expectMatch(t, m, "x(a|b)*y", "xy")
	expectNoMatch(t, m, "x(a|b)*y", "xab")
	expectMatch(t, m, "x(a|b)*y", "xay")
	expectMatch(t, m, "x(a|b)*y", "xbay")
	expectMatch(t, m, "x(a|b)*y", "xabay")
	expectMatch(t, m, "x(a|b)*y", "xababy")
	expectNoMatch(t, m, "x(a|b)+y", "xy")
	expectNoMatch(t, m, "x(a|b)+y", "xab")
	expectMatch(t, m, "x(a|b)+y", "xay")
	expectMatch(t, m, "x(a|b)+y", "xbay")
	expectMatch(t, m, "x(a|b)+y", "xabay")
	expectMatch(t, m, "x(a|b)+y", "xababy")
	// XXX
	//expectMatch(t, m, "((hello)|(goodbye))*", "")
	//expectMatch(t, m, "((hello)|(goodbye))*", "goodbye")
	//expectMatch(t, m, "((hello)|(goodbye))*", "goodbyehello")
}
