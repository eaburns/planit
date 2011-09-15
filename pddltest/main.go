package main

import (
	"fmt"
	"os"
	"flag"
	"io/ioutil"
	"log"
	"runtime/pprof"
	"goplan/pddl"
	"goplan/lifted"
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

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}

	nacts := len(dom.Actions)
	//	dom.ExpandQuants(append(dom.Constants, prob.Objects...))
	if *dump {
		fmt.Printf("%+v\n\n%+v\n", dom, prob)
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
