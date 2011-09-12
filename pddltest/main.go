package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"goplan/pddl"
)

const dump = true

func main() {
	s, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic("Error reading standard input")
	}

	p := pddl.Parse(pddl.Lex("stdin", string(s)))
	res := p.ParseDomain()
//	res := p.ParseProblem()
	err = res.UniquifyVars()
	if err != nil {
		panic(err)
	}
	if (dump) {
		fmt.Printf("%+v\n", res)
	}
}
