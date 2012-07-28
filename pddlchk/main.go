package main

import (
	"flag"
	"log"
	"os"
	"planit/pddl"
	"runtime/pprof"
)

var (
	cpuProfile = flag.String("cpuprof", "", "write CPU profile to this file")
	memProfile = flag.String("memprof", "", "write memory profile to this file")
	ignoreReqs = flag.Bool("ignore-missing-reqs", false, "ignore missing requirement errors")
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	var dom *pddl.Domain
	var prob *pddl.Problem

	if len(flag.Args()) == 0 {
		return
	}

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}

	ast, err := parseFile(flag.Arg(0))
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

	if len(flag.Args()) > 1 {
		ast, err := parseFile(flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}
		switch r := ast.(type) {
		case *pddl.Domain:
			if dom != nil {
				log.Fatal("two domains specified")
			}
			dom = r
		case *pddl.Problem:
			if dom == nil {
				log.Fatal("no domain specified")
			}
			prob = r
		default:
			panic("Impossible")
		}
	}

	const maxErrors = 5
	if errs := pddl.Check(dom, prob); len(errs) > 0 {
		if *ignoreReqs {
			var report []error
			for _, e := range errs {
				if _, ok := e.(pddl.MissingRequirementError); ok {
					continue
				}
				report = append(report, e)
			}
			errs = report
		}
		for i := 0; i < maxErrors && i < len(errs); i++ {
			log.Printf(errs[i].Error())
		}
		if len(errs) > maxErrors {
			log.Print("too many errors, truncating list")
		}
		if len(errs) > 0 {
			errors := "errors"
			if len(errs) == 1 {
				errors = "error"
			}
			log.Fatalf("%d %s\n", len(errs), errors)
		}
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
