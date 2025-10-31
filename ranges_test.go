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
			assert.True(t, false, "bad: "+s)
		}
	}
	return rs
}

func expectRanges(t *testing.T, expect string, rs Ranges) {
	t.Helper()
	assert.Equal(t, rs.String(), expect)
}

func expectDiff(t *testing.T, expectA, expectBoth, expectB string, rs1, rs2 Ranges) {
	t.Helper()
	onlyA, both, onlyB := Diff(rs1, rs2)
	assert.Equal(t, onlyA.String(), expectA)
	assert.Equal(t, both.String(), expectBoth)
	assert.Equal(t, onlyB.String(), expectB)
}

func TestRanges(t *testing.T) {
	expectRanges(t, "[]", buildRanges(t))
	expectRanges(t, "[ac]", buildRanges(t, "a", "c"))

	// these should merge, no matter which one is added first
	expectRanges(t, "[a-b]", buildRanges(t, "a", "b"))
	expectRanges(t, "[a-b]", buildRanges(t, "b", "a"))

	// overlaps merged
	expectRanges(t, "[a-e]", buildRanges(t, "ac", "be"))
	expectRanges(t, "[a-e]", buildRanges(t, "be", "ac"))

	// adjacents merged
	expectRanges(t, "[a-e]", buildRanges(t, "ac", "de"))
	expectRanges(t, "[a-e]", buildRanges(t, "de", "ac"))

	// overlaps all of the earlier parts
	expectRanges(t, "[a-e]", buildRanges(t, "b", "c", "d", "ae"))
	expectRanges(t, "[a-e]", buildRanges(t, "b", "d", "ae"))

	// fills gaps between existing regions
	expectRanges(t, "[a-e]", buildRanges(t, "e", "a", "bd"))
	expectRanges(t, "[a-e]", buildRanges(t, "e", "c", "a", "bd"))

	expectRanges(t, "[a-z]", buildRanges(t, "xz", "ab", "fj", "ay"))

	// diffs...
	expectDiff(t, "[a-n]", "[]", "[]",
		buildRanges(t, "an"),
		buildRanges(t))

	expectDiff(t, "[]", "[]", "[c-p]",
		buildRanges(t),
		buildRanges(t, "cp"))

	expectDiff(t, "[a-b]", "[c-n]", "[o-p]",
		buildRanges(t, "an"),
		buildRanges(t, "cp"))

	expectDiff(t, "[a-bq-z]", "[c-p]", "[]",
		buildRanges(t, "az"),
		buildRanges(t, "cp"))

	expectDiff(t, "[]", "[c-p]", "[a-bq-z]",
		buildRanges(t, "cp"),
		buildRanges(t, "az"))

	expectDiff(t, "[arx]", "[]", "[c]",
		buildRanges(t, "a", "r", "x"),
		buildRanges(t, "c"))

	expectDiff(t, "[c]", "[]", "[arx]",
		buildRanges(t, "c"),
		buildRanges(t, "a", "r", "x"))

}
