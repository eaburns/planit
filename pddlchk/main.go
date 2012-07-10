package main

import (
	"errors"
	"fmt"
	"os"
	"planit/pddl"
)

func main() {
	var dom *pddl.Domain
	var prob *pddl.Problem

	dom, prob, err := parseFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	if len(os.Args) > 2 {
		switch d, p, err := parseFile(os.Args[2]); {
		case err != nil:
			errorExit(err)
		case d != nil:
			if dom != nil {
				errorExit(errors.New("two domains specified"))
			}
			dom = d
		case p != nil:
			if prob != nil {
				errorExit(errors.New("two problems specified"))
			}
			prob = p
		}
	}
	if err := pddl.Check(dom, prob); err != nil {
		errorExit(err)
	}
}

func errorExit(e error) {
	fmt.Println(e.Error())
	os.Exit(1)
}

func parseFile(path string) (*pddl.Domain, *pddl.Problem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	return pddl.Parse(path, file)
}
