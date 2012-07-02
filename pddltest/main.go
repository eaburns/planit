package main

import (
	"flag"
	"fmt"
	"planit/pddl"
	"planit/prob"
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

	parser, err := pddl.NewParserFile(*dpath)
	if err != nil {
		panic(err)
	}
	d := parser.ParseDomain()
	if *dump {
		fmt.Printf("%+v\n", d)
	}

	parser, err = pddl.NewParserFile(*ppath)
	if err != nil {
		panic(err)
	}
	p := parser.ParseProblem()
	if *dump {
		fmt.Printf("%+v\n", p)
	}

	var _ = prob.Ground(d, p)
}
