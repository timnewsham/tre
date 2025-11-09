package main

import (
	"fmt"
	"os"

	"github.com/timnewsham/tre"
)

func main() {
	// s := "(a|b)*\\\\\\[ x[cr-y]\\n"
	s := "(hello|help)(a|b)*world"
	//s := "hello|help|howdy|helx"
	//s := "(a|b|c)d"
	if len(os.Args) > 1 {
		s = os.Args[1]
	}

	targ := "helloaabbaaworld"
	if len(os.Args) > 2 {
		targ = os.Args[2]
	}

	fmt.Printf("expr %q targ %q\n", s, targ)

	n, err := tre.Parse(s)
	if err != nil {
		fmt.Printf("error %v\n", err)
		return
	}

	fmt.Printf("parsed %s\n", s)
	n.Print(1)
	fmt.Printf("\n")
	nfa := tre.MakeNfa(n)
	nfa.Dot("main-nfa.dot", s)

	groups, match := tre.MatchNfa(nfa, targ)
	fmt.Printf("match is %v %v\n", match, groups)

	dfa := tre.MakeDfa(nfa)
	fmt.Printf("got dfa %v\n", dfa)
	dfa.Dot("main-dfa.dot", s)

	groups, match = tre.MatchDfa(dfa, targ)
	fmt.Printf("match is %v %v\n", match, groups)
}
