package main

import (
	"fmt"
	"os"
	"flag"
	"io/ioutil"
	"goplan/pddl"
	"goplan/lifted"
)

var dpath *string = flag.String("d", "", "The PDDL domain file")
var ppath *string = flag.String("p", "", "The PDDL problem file")
var dump *bool  = flag.Bool("dump", false, "Dump ground planning problem")

func main() {
	flag.Parse()

	dom, err := domain()
	if err != nil {
		panic(err)
	}
	syms := lifted.NewSymtab()
	err = dom.AssignNums(syms)
	if err != nil {
		panic(err)
	}

	prob, err := problem()
	if err != nil {
		panic(err)
	}
	err = prob.AssignNums(syms)
	if err != nil {
		panic(err)
	}

	nacts := len(dom.Actions)
//	dom.ExpandQuants(append(dom.Constants, prob.Objects...))
	if (*dump) {
		fmt.Printf("%+v\n\n%+v", dom, prob)
	}
	fmt.Printf("%d actions\n", nacts)
	fmt.Printf("%d grounded actions\n", len(dom.Actions))
}

func domain() (*lifted.Domain, os.Error) {
	file, err := os.Open(*dpath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open domain file %s: %s", *dpath, err)
	}
	s, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read domain file %s: %s", *dpath, err)
	}
	p := pddl.Parse(pddl.Lex(*dpath, string(s)))
	return p.ParseDomain(), nil
}

func problem() (*lifted.Problem, os.Error) {
	file, err := os.Open(*ppath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open problem file %s: %s", *dpath, err)
	}
	s, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read problem file %s: %s", *dpath, err)
	}
	p := pddl.Parse(pddl.Lex(*ppath, string(s)))
	return p.ParseProblem(), nil
}