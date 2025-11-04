package tre

import (
	"fmt"
)

type Nfa struct {
	class  Ranges // unless split is true
	next1  *Nfa
	next2  *Nfa // if split is true
	split  bool
	accept bool
}

func (p *Nfa) Dot(label string) {
	nextId := 0
	ids := make(map[*Nfa]int)
	ids[nil] = 999999 // in case something goes wrong
	getId := func(p *Nfa) (int, bool) {
		id, ok := ids[p]
		if ok {
			return id, ok
		}
		nextId++
		ids[p] = nextId
		return nextId, false
	}

	walk := func(p *Nfa) {}
	walk = func(p *Nfa) {
		id, ok := getId(p)
		if ok {
			return
		}

		if p.accept {
			fmt.Printf("  node_%d [label = \"accept\"]\n", id)
		} else {
			fmt.Printf("  node_%d [label = \"%d\"]\n", id, id)
		}

		if p.next1 != nil {
			walk(p.next1)
		}
		if p.next2 != nil {
			walk(p.next2)
		}
		if p.split {
			fmt.Printf("  node_%d -> node_%d\n", id, ids[p.next1])
			fmt.Printf("  node_%d -> node_%d\n", id, ids[p.next2])
		} else if !p.accept {
			fmt.Printf("  node_%d -> node_%d [label = \"%v\"]\n", id, ids[p.next1], p.class)
		}
	}

	fmt.Printf("digraph G {\n")
	fmt.Printf("  graph [rankdir = LR, label=%q]\n", label)
	walk(p)
	fmt.Printf("}\n")
}

type Frag struct {
	start *Nfa
	ends  []**Nfa
}

func frag(start *Nfa, ends ...**Nfa) *Frag {
	return &Frag{start: start, ends: ends}
}

func (p *Frag) outTo(start *Nfa) {
	for _, pEnd := range p.ends {
		*pEnd = start
	}
}

func nfaFrag(p *Parsed) *Frag {
	switch p.typ {
	case ParseClass:
		// -->[class]-->
		n := &Nfa{class: p.class}
		return frag(n, &n.next1)
	case ParseStar:
		//      V------------\
		// -->[alt]-->[left]-+
		//      \------------->
		left := nfaFrag(p.left)
		alt := &Nfa{split: true, next2: left.start}
		left.outTo(alt)
		return frag(alt, &alt.next1)
	case ParsePlus:
		// -->[left]-->[alt]-->
		//      ^-------/
		left := nfaFrag(p.left)
		alt := &Nfa{split: true, next2: left.start}
		left.outTo(alt)
		return frag(left.start, &alt.next1)
	case ParseConcat:
		// -->[left]-->[right]-->
		left := nfaFrag(p.left)
		right := nfaFrag(p.right)
		left.outTo(right.start)
		return frag(left.start, right.ends...)
	case ParseAlt:
		// --[alt]-->[left]-->
		//     \---->[right]-->
		left := nfaFrag(p.left)
		right := nfaFrag(p.right)
		alt := &Nfa{split: true, next1: left.start, next2: right.start}
		ends := append(left.ends, right.ends...)
		return frag(alt, ends...)
	default:
		panic(fmt.Errorf("unexpected %v", p))
		return nil
	}
}

func MakeNfa(p *Parsed) *Nfa {
	frag := nfaFrag(p)
	accept := &Nfa{accept: true}
	frag.outTo(accept)
	return frag.start
}

// addTargs adds targets that accept characters or are final states
// while following epsilon edges and avoiding duplicates.
func addTargs(n *Nfa, visited map[*Nfa]struct{}, l []*Nfa) []*Nfa {
	_, ok := visited[n]
	if !ok {
		visited[n] = struct{}{}
		switch {
		case n.split:
			l = addTargs(n.next1, visited, l)
			l = addTargs(n.next2, visited, l)
		default: // accepting states, and character consuming states.
			l = append(l, n)
		}
	}
	return l
}

func advanceEpsilon(n *Nfa) []*Nfa {
	visited := make(map[*Nfa]struct{})
	return addTargs(n, visited, nil)
}

func advance(ns []*Nfa, ch rune) []*Nfa {
	visited := make(map[*Nfa]struct{})
	var l []*Nfa
	for _, n := range ns {
		if !n.split && !n.accept && n.class.Contains(ch) {
			l = addTargs(n.next1, visited, l)
		}
	}
	return l
}

func accepts(ns []*Nfa) bool {
	for _, n := range ns {
		if n.accept {
			return true
		}
	}
	return false
}

func MatchNfa(n *Nfa, s string) bool {
	ns := advanceEpsilon(n) // follow epsilon edges from start
	for pos, ch := range []rune(s) {
		_ = pos
		ns = advance(ns, ch)
		if len(ns) == 0 {
			return false
		}
	}
	return accepts(ns)
}
