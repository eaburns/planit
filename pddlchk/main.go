package main

import (
	"log"
	"os"
	"planit/pddl"
)

func main() {
	doms := map[string]*pddl.Domain{}
	probs := []*pddl.Problem{}

	for i := 1; i < len(os.Args); i++ {
		f := os.Args[i]
		in, err := os.Open(f)
		if err != nil {
			log.Println(err)
			continue
		}
		d, p, err := pddl.Parse(f, in)
		if err != nil {
			log.Println(err)
			continue
		}
		if d != nil {
			if doms[d.Name] != nil {
				log.Printf("domain %s: specified multiple times\n", d.Name)
				continue
			}
			doms[d.Name] = d
		} else {
			probs = append(probs, p)
		}
	}
	chkd := map[string]bool{}
	for _, p := range probs {
		d := doms[p.Domain]
		if d == nil {
			log.Printf("problem %s: missing domain %s\n", p.Name, p.Domain)
			continue
		}
		log.Printf("checking %s and %s", d.Name, p.Name)
		if err := pddl.Check(d, p); err != nil {
			log.Println(err)
		}
		chkd[d.Name] = true
	}
	for _, d := range doms {
		if chkd[d.Name] {
			continue
		}
		log.Printf("checking %s", d.Name)
		if _, err := pddl.CheckDomain(d); err != nil {
			log.Println(err)
		}
	}
}
