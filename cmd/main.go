package main

import (
	"fmt"

	"github.com/timnewsham/tre"
)

func main() {
	s := "(a|b)*\\\\\\[ x[cr-y]\\n"
	n, err := tre.Parse(s)
	if err != nil {
		fmt.Printf("error %v\n", err)
	} else {
		fmt.Printf("parsed %s\n", s)
		n.Print(1)
	}
}
