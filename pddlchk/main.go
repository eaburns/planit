package main

import (
	"log"
	"os"
	"planit/pddl"
)

func main() {
	log.SetFlags(0)
	var dom *pddl.Domain
	var prob *pddl.Problem

	ast, err := parseFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	switch r := ast.(type) {
	case *pddl.Domain:
		dom = r
	case *pddl.Problem:
		prob = r
	default:
		panic("Impossible")
	}

	if len(os.Args) > 2 {
		ast, err := parseFile(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		switch r := ast.(type) {
		case *pddl.Domain:
			if dom != nil {
				panic("two domains specified")
			}
			dom = r
		case *pddl.Problem:
			if dom == nil {
				panic("no domain specified")
			}
			prob = r
		default:
			panic("Impossible")
		}
	}
	if err := pddl.Check(dom, prob); err != nil {
		log.Fatal(err)
	}
}

func parseFile(path string) (interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return pddl.Parse(path, file)
}
