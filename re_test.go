package tre

import (
	"fmt"
	"slices"
	"testing"

	"github.com/alecthomas/assert"
)

type matchFunc func(t *testing.T, pat, s string, wantMatch bool, wantGroups ...string)

func matchNfa(t *testing.T, pat, s string, wantMatch bool, wantGroups ...string) {
	//t.Helper()
	p, err := Parse(pat)
	assert.NoError(t, err)
	nfa := MakeNfa(p)
	groups, m := MatchNfa(nfa, s)

	if m != wantMatch || !slices.Equal(groups, wantGroups) {
		fmt.Printf("match %v with %v was %v %v wanted %v %v\n", s, pat, m, groups, wantMatch, wantGroups)
		fmt.Printf("parsed (see test-nfa.dot):\n")
		p.Print(1)
		nfa.Dot("test-nfa.dot", pat)
	}
	assert.Equal(t, m, wantMatch)
	assert.True(t, slices.Equal(groups, wantGroups))
}

func matchNfaBounded(t *testing.T, pat, s string, wantMatch bool, wantGroups ...string) {
	//t.Helper()
	p, err := ParseBounded(pat)
	assert.NoError(t, err)
	nfa := MakeNfa(p)
	groups, m := MatchNfa(nfa, s)
	if m != wantMatch || !slices.Equal(groups, wantGroups) {
		fmt.Printf("match %v with %v was %v %v wanted %v %v\n", s, pat, m, groups, wantMatch, wantGroups)
		fmt.Printf("parsed (see test-nfa-bounded.dot):\n")
		p.Print(1)
		nfa.Dot("test-nfa-bounded.dot", pat)
	}
	assert.Equal(t, m, wantMatch)
	assert.True(t, slices.Equal(groups, wantGroups))
}

func matchDfa(t *testing.T, pat, s string, wantMatch bool, wantGroups ...string) {
	//t.Helper()
	p, err := Parse(pat)
	assert.NoError(t, err)
	nfa := MakeNfa(p)
	dfa := MakeDfa(nfa)
	groups, m := MatchDfa(dfa, s)
	if m != wantMatch || !slices.Equal(groups, wantGroups) {
		fmt.Printf("match %v with %v was %v %v wanted %v %v\n", s, pat, m, groups, wantMatch, wantGroups)
		fmt.Printf("parsed (see test-dfa.dot):\n")
		p.Print(1)
		dfa.Dot("test-dfa.dot", pat)
	}
	assert.Equal(t, m, wantMatch)
	assert.True(t, slices.Equal(groups, wantGroups))
}

func expectMatch(t *testing.T, match matchFunc, pat, s string, wantGroups ...string) {
	//t.Helper()
	match(t, pat, s, true, wantGroups...)
}

func expectNoMatch(t *testing.T, match matchFunc, pat, s string, wantGroups ...string) {
	//t.Helper()
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

			expectNoMatch(t, m, "a*aa*", "")
			expectMatch(t, m, "a*aa*", "a")
			expectMatch(t, m, "a*aa*", "aaaaaa")

			expectMatch(t, m, ".*", "hello")
			expectMatch(t, m, "a.*", "ahello")
			expectMatch(t, m, ".*a", "helloa")
			expectNoMatch(t, m, ".*a", "hellob")

			expectMatch(t, m, "a?", "")
			expectMatch(t, m, "a?", "a")
			expectMatch(t, m, "b*a?b*", "a")
			expectMatch(t, m, "b*a?b*", "bbbb")
			expectMatch(t, m, "b*a?b*", "bbabb")
			expectMatch(t, m, "a*a?a*", "")
			expectMatch(t, m, "a*a?a*", "aaaaa")

			// no infinite loops on ** please even though the NFA has epsilon cycles.
			expectMatch(t, m, "a**", "")
			expectMatch(t, m, "a**", "a")
			expectMatch(t, m, "a**", "aaaaa")
			expectNoMatch(t, m, "a**", "aaaaab")

			// overlapping
			expectMatch(t, m, "hello|help", "hello")
			expectMatch(t, m, "hello|help", "help")
			expectNoMatch(t, m, "hello|help", "hellop")

			expectMatch(t, m, "he(?ll)o(?a*)", "helloaaa", "ll", "aaa")

			// greedy matching should capture all the a's in the group.
			expectMatch(t, m, "a(?a*)a?b", "aaaab", "aaa")

			// greedy matching should make this match fail because all the a's are in the group.
			expectNoMatch(t, m, "a(?a*)ab", "aaaab")
		})
	}
}

func TestReBounded(t *testing.T) {
	m := matchNfaBounded
	expectMatch(t, m, "/a*/", "")
	expectMatch(t, m, "/a*/", "aaaaa")
	expectMatch(t, m, "#a*#", "aaaaa")

	_, err := ParseBounded("/foo/")
	assert.NoError(t, err)

	_, err = ParseBounded("/foo/ ")
	assert.Error(t, err)

	_, err = ParseBounded("/foo")
	assert.Error(t, err)

	_, err = ParseBounded("#/foo/bar/baz#")
	assert.NoError(t, err)

	_, err = ParseBounded("/\\/foo\\/bar\\/baz/")
	assert.NoError(t, err)

	_, err = ParseBounded("/foo/bar/baz/")
	assert.Error(t, err)
}
