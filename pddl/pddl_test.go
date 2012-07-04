package pddl

import (
	"testing"
	"os"
)

func TestParseDomain(t *testing.T) {
	_, err := ParseDomain("p01-domain.pddl")
	if err != nil {
		t.Error(err)
	}
}

func TestPrintDomain(t *testing.T) {
	d, err := ParseDomain("p01-domain.pddl")
	if err != nil {
		t.Error(err)
	}
	PrintDomain(os.Stdout, d)
}

func TestParseProblem(t *testing.T) {
	_, err := ParseProblem("p01.pddl")
	if err != nil {
		t.Error(err)
	}
}


func TestPrintProblem(t *testing.T) {
	p, err := ParseProblem("p01.pddl")
	if err != nil {
		t.Error(err)
	}
	PrintProblem(os.Stdout, p)
}
