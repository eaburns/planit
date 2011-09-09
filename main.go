package main

import (
	"fmt"
	"os"
	"io/ioutil"
)

const dump = false

func main() {
	s, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic("Error reading standard input")
	}

	l := lex("stdin", string(s))
	p := parse(l)
	d := p.parseDomain()
//	d := p.parseProblem()
	if (dump) {
		fmt.Printf("%+v\n", d)
	}
}
