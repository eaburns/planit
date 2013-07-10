// © 2013 the PlanIt Authors under the MIT license. See AUTHORS for the list of authors.

package main

import (
	"log"
	"os"
	"planit/pddl"
)

func main() {
	log.SetFlags(0)
	for _, path := range os.Args[1:] {
		file, err := os.Open(path)
		if err != nil {
			log.Println(err)
			continue
		}
		ast, err := pddl.Parse(path, file)
		if err != nil {
			log.Println(err)
			continue
		}
		switch r := ast.(type) {
		case *pddl.Domain:
			pddl.PrintDomain(os.Stdout, r)
		case *pddl.Problem:
			pddl.PrintProblem(os.Stdout, r)
		default:
			panic("impossible")
		}
		file.Close()
	}
}
