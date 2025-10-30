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
	v := fmt.Sprintf("%q", ch)
	return v[1:len(v)-1]
}

func (r Range) String() string {
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
	newr := Range{rmin, rmax}
	fmt.Printf("add %v to %v\n", newr, *rs)

	for n, r := range *rs {
		// union will merge with adjacent regions.
		if (*rs)[n].union(newr) {
			fmt.Printf("- XXX merged %v with %v to get %v\n", r, newr, (*rs)[n])
			// We merged newr with rs[n].
			// now merge as many rs[m] after rs[n] as possible with rs[n].
			m := n
			for m < len(*rs) && (*rs)[n].union((*rs)[m]) {
				fmt.Printf("  also merged %v to get %v\n", (*rs)[m], (*rs)[n])
				m ++
			}

			// discard merged ranges rs[n+1:m]
			fmt.Printf("  and discard %v\n", (*rs)[n+1 : m])
			*rs = append((*rs)[:n+1], (*rs)[m:]...)
			fmt.Printf("  now %v\n", *rs)
			return
		}

		if rmin < r.rmax {
			// insert newr before rs[n].
			fmt.Printf("- inserting %v between %v and %v\n", newr, (*rs)[:n], (*rs)[n:])
			newrs := make(Ranges, len(*rs) + 1)
			copy(newrs[:n], (*rs)[:n])
			newrs[n] = newr
			copy(newrs[n+1:], (*rs)[n:])
			*rs = newrs
			fmt.Printf("  now %v\n", *rs)
			return
		}
	}

	// append newr
	fmt.Printf("- appending %v to %v\n", newr, *rs)
	*rs = append(*rs, newr)
	fmt.Printf("  now %v\n", *rs)
}

