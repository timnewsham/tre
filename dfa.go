package tre

import (
	"fmt"
	"slices"
	"unsafe"
)

type Edge struct {
	class Ranges
	next  *Dfa
}

type Dfa struct {
	accept bool
	edges  []Edge
}

func (p *Dfa) Dot() {
	nextId := 0
	ids := make(map[*Dfa]int)
	ids[nil] = 999999 // in case something goes wrong
	getId := func(p *Dfa) (int, bool) {
		id, ok := ids[p]
		if ok {
			return id, ok
		}
		nextId++
		ids[p] = nextId
		return nextId, false
	}

	walk := func(p *Dfa) {}
	walk = func(p *Dfa) {
		id, ok := getId(p)
		if ok {
			return
		}

		if p.accept {
			fmt.Printf("  node_%d [label = \"accept\"]\n", id)
		} else {
			fmt.Printf("  node_%d [label = \"%d\"]\n", id, id)
		}

		for _, edge := range p.edges {
			walk(edge.next)
		}
		for _, edge := range p.edges {
			fmt.Printf("  node_%d -> node_%d [label = \"%v\"]\n", id, ids[edge.next], edge.class)
		}
	}
	fmt.Printf("digraph G {\n")
	walk(p)
	fmt.Printf("}\n")
}

type NfaSet struct {
	set []*Nfa
	dfa *Dfa
}

func cmpNfa(a, b *Nfa) int {
	pa := uintptr(unsafe.Pointer(a))
	pb := uintptr(unsafe.Pointer(b))
	return int(pa - pb)
}

func sortNfas(ns []*Nfa) {
	slices.SortFunc(ns, cmpNfa)
}

func eqNfas(a, b []*Nfa) bool {
	return slices.Equal(a, b)
}

// addNfaSet finds set in l, or adds a new entry for set to l.
// It returns the updated list, the entry for s, and true if the item already existed, and false if it was newly created.
func addNfaSet(l []NfaSet, set []*Nfa) ([]NfaSet, *Dfa, bool) {
	sortNfas(set)

	for _, state := range l {
		if eqNfas(state.set, set) {
			//fmt.Printf("nfa set %v exists\n", set)
			return l, state.dfa, true
		}
	}

	//fmt.Printf("creating nfa set %v\n", set)
	dfa := &Dfa{accept: accepts(set)}
	l = append(l, NfaSet{set, dfa})
	return l, dfa, false
}

func classInClasses(cs []Ranges, c Ranges) bool {
	for _, c2 := range cs {
		if slices.Equal(c2, c) {
			return true
		}
	}
	return false
}

// disjointClasses returns a list of non-overlapping character classes
// accepted by the NFA states in ns.
func disjointClasses(ns []*Nfa) []Ranges {
	var classes []Ranges
	for _, n := range ns {
		if n.accept || n.split {
			continue
		}

		//fmt.Printf("class %v\n", n.class)
		if len(classes) == 0 {
			classes = append(classes, n.class)
			continue
		}

		var newClasses []Ranges
		for _, class := range classes {
			onlyA, both, onlyB := Diff(class, n.class)
			for _, c := range []Ranges{onlyA, both, onlyB} {
				if len(c) > 0 {
					if !classInClasses(newClasses, c) {
						newClasses = append(newClasses, c)
					}
				}
			}
		}
		classes = newClasses
	}
	//fmt.Printf("  disjoint clases %v\n", classes)
	return classes
}

func MakeDfa(n *Nfa) *Dfa {
	var states []NfaSet
	ns := advanceEpsilon(n) // follow epsilon edges from start

	states, dstart, _ := addNfaSet(states, ns)

	addEdge := func(d *Dfa, class Ranges, targ *Dfa) {
		for n := range d.edges {
			// if we already have an edge to targ
			// just augment its class with the new class.
			if d.edges[n].next == targ {
				d.edges[n].class.AddRanges(class)
				return
			}
		}
		d.edges = append(d.edges, Edge{class: class, next: targ})
	}

	explore := func(d *Dfa, ns []*Nfa) {}
	explore = func(d *Dfa, ns []*Nfa) {
		// NOTE: some of the disjointed classes might still go to the same location.
		// These get merged in addEdge.
		for _, class := range disjointClasses(ns) {
			ch := class[0].rmin // exemplary char. the rest should flow the same way.
			ns2 := advance(ns, ch)
			if len(ns2) == 0 {
				continue
			}

			var dtarg *Dfa
			var visited bool
			states, dtarg, visited = addNfaSet(states, ns2)
			addEdge(d, class, dtarg)
			if !visited {
				explore(dtarg, ns2)
			}
		}
	}

	explore(dstart, ns)
	return dstart
}

func MatchDfa(d *Dfa, s string) bool {
nextChar:
	for _, ch := range []rune(s) {
		for _, edge := range d.edges {
			if edge.class.Contains(ch) {
				d = edge.next
				continue nextChar
			}
		}
		return false
	}
	return d.accept
}
