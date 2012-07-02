package main

import (
	"flag"
	"fmt"
	"planit/pddl"
	"planit/prob"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
)

var dpath = flag.String("d", "", "The PDDL domain file")
var ppath = flag.String("p", "", "The PDDL problem file")
var dump = flag.Bool("dump", false, "Dump ground planning problem")
var cpuprofile = flag.String("cpuprofile", "", "Write CPU profile to this file")
var memprofile = flag.String("memprofile", "", "Write memory profile to this file")

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}

	d, err := domain()
	if err != nil {
		panic(err)
	}
	if *dump {
		fmt.Printf("%+v\n", d)
	}

	p, err := problem()
	if err != nil {
		panic(err)
	}
	if *dump {
		fmt.Printf("%+v\n", p)
	}

	var _ = prob.Ground(d, p)
}

func domain() (*prob.Domain, error) {
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

func problem() (*prob.Problem, error) {
	file, err := os.Open(*ppath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open problem file %s: %s", *ppath, err)
	}
	s, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read problem file %s: %s", *ppath, err)
	}
	p := pddl.Parse(pddl.Lex(*ppath, string(s)))
	return p.ParseProblem(), nil
}
