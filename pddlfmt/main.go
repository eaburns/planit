package main

import (
	"os"
	"log"
	"planit/pddl"
)

func main() {
	for _, path := range os.Args[1:] {
		file, err := os.Open(path)
		if err != nil {
			log.Println(err)
			continue
		}
		switch d, p, err := pddl.Parse(path, file); {
		case err != nil:
			log.Println(err)
		case d != nil:
			pddl.PrintDomain(os.Stdout, d)
		case p != nil:
			pddl.PrintProblem(os.Stdout, p)
		}
		file.Close()
	}
}