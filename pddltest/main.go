package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"goplan/pddl"
)

const dump = false

func main() {
	s, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic("Error reading standard input")
	}

	p := pddl.Parse(pddl.Lex("stdin", string(s)))
	d := p.ParseDomain()
//	d := p.ParseProblem()
	if (dump) {
		fmt.Printf("%+v\n", d)
	}
}
