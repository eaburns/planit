// © 2013 the PlanIt Authors under the MIT license. See AUTHORS for the list of authors.

// inertia prints the inertial status of each predicate.
// The table that is printed, for each domain, mimics
// that of Figure 5 in:
// On the Instantiation of ADL Operators Involving
// Arbitrary First-Order Formulas, by Koehler and
// Hoffmann, 2000.
package main

import (
	"fmt"
	"log"
	"os"
	"planit/pddl"
	"text/tabwriter"
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
		}
		switch d := ast.(type) {
		case *pddl.Problem:
			log.Println(path, "is a problem file, skipping")
		case *pddl.Domain:
			if err := pddl.Check(d, nil); err != nil {
				log.Println(err)
				continue
			}
			fmt.Println(d)
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
			fmt.Fprintln(w, "predicate\tpos. effect\tneg. effect\tstatus")
			for _, pred := range d.Predicates {
				status := "inertia"
				if pred.PosEffect && pred.NegEffect {
					status = "fluent"
				} else if pred.PosEffect {
					status = "neg. inertia"
				} else if pred.NegEffect {
					status = "pos. inertia"
				}
				fmt.Fprintf(w, "%s\t%t\t%t\t%s\n", pred,
					pred.PosEffect, pred.NegEffect, status)
			}
			w.Flush()
		}
		file.Close()
	}
}
