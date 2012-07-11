// inertia prints the inertial status of each predicate.
// The table that is printed, for each domain, mimics
// that of Figure 5 in:
// On the Instantiation of ADL Operators Involving
// Arbitrary First-Order Formulas, by Koehler and
// Hoffmann, 2000.
package main

import (
	"log"
	"os"
	"planit/pddl"
	"text/tabwriter"
	"fmt"
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
		case p != nil:
			log.Println(path, "is a problem file, skipping")
		case d != nil:
			if err := pddl.Check(d, nil); err != nil {
				log.Println(err)
				continue
			}
			fmt.Println(d.Identifier)
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
			fmt.Fprintln(w, "predicate\tpos. effect\tneg. effect\tstatus\n")
			for _, pred := range d.Predicates {
				status := "inertia"
				if pred.PosEffect && pred.NegEffect {
					status = "fluent"
				} else if pred.PosEffect {
					status = "neg. inertia"
				} else if pred.NegEffect {
					status = "pos. inertia"
				}
				fmt.Fprintf(w, "%s\t%t\t%t\t%s\n", pred.Identifier,
					pred.PosEffect, pred.NegEffect, status)
			}
			w.Flush()
		}
		file.Close()
	}
}
