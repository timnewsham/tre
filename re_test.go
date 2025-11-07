package tre

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
)

type matchFunc func(t *testing.T, pat, s string, wantMatch bool)

func matchNfa(t *testing.T, pat, s string, wantMatch bool) {
	t.Helper()
	p, err := Parse(pat)
	assert.NoError(t, err)
	nfa := MakeNfa(p)
	m := MatchNfa(nfa, s)
	if m != wantMatch {
		fmt.Printf("match %v with %v failed\n", s, pat)
		fmt.Printf("parsed:\n")
		p.Print(1)
		nfa.Dot("test-nfa.dot", pat)
	}
	assert.Equal(t, m, wantMatch)
}

func matchDfa(t *testing.T, pat, s string, wantMatch bool) {
	t.Helper()
	p, err := Parse(pat)
	assert.NoError(t, err)
	nfa := MakeNfa(p)
	dfa := MakeDfa(nfa)
	m := MatchDfa(dfa, s)
	if m != wantMatch {
		fmt.Printf("match %v with %v failed\n", s, pat)
		fmt.Printf("parsed:\n")
		p.Print(1)
		dfa.Dot("test-dfa.dot", pat)
	}
	assert.Equal(t, m, wantMatch)
}

func expectMatch(t *testing.T, match matchFunc, pat, s string) {
	t.Helper()
	match(t, pat, s, true)
}

func expectNoMatch(t *testing.T, match matchFunc, pat, s string) {
	t.Helper()
	match(t, pat, s, false)
}

func TestRe(t *testing.T) {
	matchers := []struct {
		name    string
		matcher matchFunc
	}{
		{"nfa-match", matchNfa},
		{"dfa-match", matchDfa},
	}

	for _, test := range matchers {
		m := test.matcher
		t.Run(test.name, func(t *testing.T) {
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
			expectMatch(t, m, "(hello|goodbye)", "hello")
			expectMatch(t, m, "(hello|goodbye)", "goodbye")
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
			expectMatch(t, m, "(hello|goodbye)*", "")
			expectMatch(t, m, "(hello|goodbye)*", "goodbye")
			expectMatch(t, m, "(hello|goodbye)*", "goodbyehello")

			// no infinite loops on ** please even though the NFA has epsilon cycles.
			expectMatch(t, m, "a**", "")
			expectMatch(t, m, "a**", "a")
			expectMatch(t, m, "a**", "aaaaa")
			expectNoMatch(t, m, "a**", "aaaaab")

			// overlapping
			expectMatch(t, m, "hello|help", "hello")
			expectMatch(t, m, "hello|help", "help")
			expectNoMatch(t, m, "hello|help", "hellop")
		})
	}
}
