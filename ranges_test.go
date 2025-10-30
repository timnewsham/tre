package tre

import (
	"testing"

	"github.com/alecthomas/assert"
)

func buildRanges(t *testing.T, ss ...string) Ranges {
	var rs Ranges
	for _, s := range ss {
		runes := []rune(s)
		if len(s) == 1 {
			rs.Add1(runes[0])
		} else if len(s) == 2 {
			rs.Add(runes[0], runes[1])
		} else {
			assert.True(t, false, "bad: " + s)
		}
	}
	return rs
}

func expectRanges(t *testing.T, expect string, ss ...string) {
	t.Helper()
	rs := buildRanges(t, ss...)
	rstr := rs.String()
	assert.Equal(t, rstr, expect)
}

func TestRanges(t *testing.T) {
	expectRanges(t, "[]")
	expectRanges(t, "[ac]", "a", "c")

	// these should merge, no matter which one is added first
	expectRanges(t, "[a-b]", "a", "b")
	expectRanges(t, "[a-b]", "b", "a")

	// overlaps merged
	expectRanges(t, "[a-e]", "ac", "be")
	expectRanges(t, "[a-e]", "be", "ac")

	// adjacents merged
	expectRanges(t, "[a-e]", "ac", "de")
	expectRanges(t, "[a-e]", "de", "ac")

	// overlaps all of the earlier parts
	expectRanges(t, "[a-e]", "b", "c", "d", "ae")
	expectRanges(t, "[a-e]", "b", "d", "ae")

	// fills gaps between existing regions
	expectRanges(t, "[a-e]", "e", "a", "bd")
	expectRanges(t, "[a-e]", "e", "c", "a", "bd")

	expectRanges(t, "[a-z]", "xz", "ab", "fj", "ay")
}
