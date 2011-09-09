package main

import (
	"fmt"
	"os"
	"io/ioutil"
)

func main() {
	s, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic("Error reading standard input")
	}

	l := lex("stdin", string(s))
	p := parse(l)
//	d := p.parseDomain()
	d := p.parseProblem()
	fmt.Printf("%+v\n", d)
}
