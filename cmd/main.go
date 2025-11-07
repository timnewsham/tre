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

	n, err := tre.Parse(s)
	if err != nil {
		fmt.Printf("error %v\n", err)
		return
	}

	fmt.Printf("parsed %s\n", s)
	n.Print(1)
	fmt.Printf("\n")
	nfa := tre.MakeNfa(n)
	nfa.Dot("nfa.dot", s)

	match := tre.MatchNfa(nfa, "helloaabbaaworld")
	fmt.Printf("match is %v\n", match)

	dfa := tre.MakeDfa(nfa)
	fmt.Printf("got dfa %v\n", dfa)
	dfa.Dot("dfa.dot", s)
}
