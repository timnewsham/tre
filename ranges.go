package tre

import (
	"fmt"
	"strings"
)

type Range struct {
	rmin rune
	rmax rune
}

func printable(ch rune) string {
	if ch == '-' {
		return "\\-"
	}
	v := fmt.Sprintf("%q", ch)
	return v[1 : len(v)-1]
}

func (r Range) String() string {
	if r.rmax < r.rmin {
		return fmt.Sprintf("EMPTY")
	}
	if r.rmin == r.rmax {
		return fmt.Sprintf("%s", printable(r.rmin))
	}
	return fmt.Sprintf("%s-%s", printable(r.rmin), printable(r.rmax))
}

func (r Range) Contains(v rune) bool {
	return r.rmin <= v && v <= r.rmax
}

func (r Range) overlaps(other Range) bool {
	return r.rmin <= other.rmax && other.rmin <= r.rmax
}

func (r *Range) union(other Range) bool {
	// Note: the plus ones allow merging adjacent ranges that dont overlap.
	if r.rmin <= other.rmax+1 && other.rmin <= r.rmax+1 {
		if other.rmin < r.rmin {
			r.rmin = other.rmin
		}
		if r.rmax < other.rmax {
			r.rmax = other.rmax
		}
		return true
	}
	return false
}

// ranges contains pairs of min, max, in sorted order.
type Ranges []Range

func newRange(rmin, rmax rune) Ranges {
	var rs Ranges
	rs.Add(rmin, rmax)
	return rs
}

func newRange1(ch rune) Ranges {
	return newRange(ch, ch)
}

func (rs Ranges) String() string {
	var strs []string
	for _, r := range rs {
		strs = append(strs, r.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ""))
}

func (rs Ranges) Contains(v rune) bool {
	// TODO: binary search for efficiency
	for _, r := range rs {
		if r.Contains(v) {
			return true
		}
	}
	return false
}

func (rs *Ranges) Add1(ch rune) {
	rs.Add(ch, ch)
}

func (rs *Ranges) Add(rmin, rmax rune) {
	if rmin > rmax {
		return
	}

	var before, after Ranges
	for _, r := range *rs {
		switch {
		case r.rmax+1 < rmin:
			// r entirely before (rmin..rmax).
			before = append(before, r)
		case rmax+1 < r.rmin:
			// r entirely after (rmin..rmax).
			after = append(after, r)
		default:
			// r overlaps with (rmin..rmax).
			// update (rmin..rmax) to include all of r.
			rmin = min(rmin, r.rmin)
			rmax = max(rmax, r.rmax)
		}
	}

	*rs = append(append(before, Range{rmin, rmax}), after...)
}

func (rs *Ranges) AddRanges(rs2 Ranges) {
	for _, r := range rs2 {
		rs.Add(r.rmin, r.rmax)
	}
}

const maxRune rune = 0x7ffffffe // XXX hack, adding one doesnt roll over.

func FullRanges() Ranges {
	var rs Ranges
	rs.Add(0, maxRune)
	return rs
}

func (rs Ranges) Invert() Ranges {
	var outrs Ranges

	var pos rune
	maxpos := maxRune
	for _, r := range rs {
		if pos < r.rmin {
			outrs.Add(pos, r.rmin-1)
		}
		pos = r.rmax + 1
	}

	if pos < maxpos {
		outrs.Add(pos, maxpos)
	}
	return outrs
}

// Diff returns ranges only in as, ranges in both, and ranges only in bs.
func Diff(as, bs Ranges) (Ranges, Ranges, Ranges) {
	trimmed := func(pos rune, r Range) Range {
		// r trimmed to start at pos or later.
		return Range{max(pos, r.rmin), r.rmax}
	}

	var (
		both, onlyA, onlyB Ranges
		aIdx, bIdx         int
		pos                rune
	)
	for aIdx < len(as) && bIdx < len(bs) {
		// get the unconsumed part of a and b.
		a := trimmed(pos, as[aIdx])
		b := trimmed(pos, bs[bIdx])

		// consume the smaller part of a or b, and advance pos.
		// XXX TODO fix this: we assume adding 1 to pos/max doesnt overflow!
		switch {
		case a.rmin < b.rmin:
			pos = min(a.rmax+1, b.rmin)
			onlyA.Add(a.rmin, pos-1)

		case b.rmin < a.rmin:
			pos = min(b.rmax+1, a.rmin)
			onlyB.Add(b.rmin, pos-1)

		case a.rmin == b.rmin:
			pos = min(a.rmax+1, b.rmax+1)
			both.Add(a.rmin, pos-1)
		}

		// move to next element if pos has completely consumed the current element.
		if pos > a.rmax {
			aIdx += 1
		}
		if pos > b.rmax {
			bIdx += 1
		}
	}

	for aIdx < len(as) {
		// consume unconsumed parts of a.
		a := trimmed(pos, as[aIdx])
		onlyA.Add(a.rmin, a.rmax)
		aIdx++
		pos = a.rmax + 1
	}

	for bIdx < len(bs) {
		// consume unconsumed parts of b.
		b := trimmed(pos, bs[bIdx])
		onlyB.Add(b.rmin, b.rmax)
		bIdx++
		pos = b.rmax + 1
	}

	return onlyA, both, onlyB
}
