package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"planit/pddl"
	"planit/prob"
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

	d, err := pddl.ParseDomain(*dpath)
	if err != nil {
		panic(err)
	}
	if *dump {
		fmt.Printf("%+v\n", d)
	}

	p, err := pddl.ParseProblem(*ppath)
	if err != nil {
		panic(err)
	}
	if *dump {
		fmt.Printf("%+v\n", p)
	}

	var _ = prob.Ground(d, p)
}
