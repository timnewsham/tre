package tre

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
)

type matchFunc func(t *testing.T, pat, s string, wantMatch bool, wantGroups ...string)

func debugParse(t *testing.T, pat string) *Parsed {
	p, err := Parse(pat)
	if err != nil {
		fmt.Printf("Parse error %v on %v\n", err, pat)
		p.Print(1)
	}
	assert.NoError(t, err)

	fmt.Printf("parsed %v: \n", pat)
	p.Print(1)
	return p
}

func debugNfa(t *testing.T, fn, pat string) *Nfa {
	nfa := MakeNfa(debugParse(t, pat))
	if fn != "" {
		fmt.Printf("writing NFA to %v\n", fn)
		nfa.Dot(fn, pat)
	}
	return nfa
}

func debugDfa(t *testing.T, fn, pat string) *Dfa {
	nfa := debugNfa(t, "", pat)
	dfa := MakeDfa(nfa)
	if fn != "" {
		fmt.Printf("writing DFA to %v\n", fn)
		dfa.Dot(fn, pat)
	}
	return dfa
}

func match(t *testing.T, mach Matcher, s string, wantMatch bool, wantGroups ...string) {
	//t.Helper()
	groups, match := mach.Match(s)
	assert.Equal(t, match, wantMatch)
	assert.Equal(t, groups, wantGroups)
}

func matchNfa(t *testing.T, pat, s string, wantMatch bool, wantGroups ...string) {
	//t.Helper()
	mach, err := NewNfa(pat)
	assert.NoError(t, err)
	match(t, mach, s, wantMatch, wantGroups...)
}

func matchNfaBounded(t *testing.T, pat, s string, wantMatch bool, wantGroups ...string) {
	//t.Helper()
	p, err := ParseBounded(pat)
	assert.NoError(t, err)
	match(t, MakeNfa(p), s, wantMatch, wantGroups...)
}

func matchDfa(t *testing.T, pat, s string, wantMatch bool, wantGroups ...string) {
	//t.Helper()
	mach, err := NewDfa(pat)
	assert.NoError(t, err)
	match(t, mach, s, wantMatch, wantGroups...)
}

func expectMatch(t *testing.T, match matchFunc, pat, s string, wantGroups ...string) {
	//t.Helper()
	ok := false
	defer func() {
		fmt.Printf("expectMatch %v %v %v -> %v\n", pat, s, wantGroups, ok)
	}()

	match(t, pat, s, true, wantGroups...)
	ok = true
}

func expectNoMatch(t *testing.T, match matchFunc, pat, s string, wantGroups ...string) {
	//t.Helper()
	ok := false
	defer func() {
		defer fmt.Printf("expectNoMatch %v %v %v -> %v\n", pat, s, wantGroups, ok)
	}()

	match(t, pat, s, false)
	ok = true
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

		// ugh, t.Run makes assert lib not give proper stack traces.. why?
		//t.Run(test.name, func(t *testing.T) {

		fmt.Printf("name: %v\n", test.name)
		if true {
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
		}
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
