package pddl

import (
	"os"
	"testing"
	"strings"
)

const (
	testDomainFile  = "p01-domain.pddl"
	testProblemFile = "p01.pddl"
)

func TestParseDomain(t *testing.T) {
	file, err := os.Open(testDomainFile)
	if err != nil {
		t.Error(err)
	}
	_, err = ParseDomain(testDomainFile, file)
	if err != nil {
		t.Error(err)
	}
}

/*
func TestPrintDomain(t *testing.T) {
	file, err := os.Open(testDomainFile)
	if err != nil {
		t.Error(err)
	}
	d, err := ParseDomain(testDomainFile, file)
	PrintDomain(os.Stdout, d)
}
*/

func TestCheckDomain(t *testing.T) {
	tests := [...]struct{
		pddl string
		ok bool
	}{
		{ "(define (domain x) (:requirements :strips))", true },
		{ "(define (domain x) (:requirements :foobar))", false },

		{ "(define (domain x) (:types t))", false },
		{ "(define (domain x) (:requirements :typing) (:types t))", true },
		{ "(define (domain x) (:requirements :typing) (:types t - undecl))", false },
		{ "(define (domain x) (:requirements :typing) (:types s t - s))", true },
		{ "(define (domain x) (:requirements :typing) (:types t - u u))", true },
		{ "(define (domain x) (:requirements :typing) (:types t - object))", true },
		{ "(define (domain x) (:requirements :typing) (:types u - (either s t) s t))", false },

		{ "(define (domain x) (:requirements :typing) (:constants c - undecl))", false },
		{ "(define (domain x) (:requirements :typing) (:constants c - object))", true },
		{ "(define (domain x) (:constants c - unreqd))", false },
		{ "(define (domain x) (:requirements :typing) (:types t) (:constants c - t ))", true },
		{ "(define (domain x) (:requirements :typing) (:types t) (:constants c - (either t undecl) ))", false },
		{ "(define (domain x) (:requirements :typing) (:types s t) (:constants c - (either s t) ))", true },

		{ "(define (domain x) (:predicates (p ?parm)))", true },
		{ "(define (domain x) (:predicates (p ?parm - unreqd)))", false },
		{ "(define (domain x) (:requirements :typing) (:predicates (p ?parm - object)))", true },
		{ "(define (domain x) (:requirements :typing) (:types t) (:predicates (p ?parm - t)))", true },
		{ "(define (domain x) (:requirements :typing) (:predicates (p ?parm - undecl)))", false },
		{ "(define (domain x) (:requirements :typing) (:types t) (:predicates (p ?parm - (either t undecl))))", false },
		{ "(define (domain x) (:requirements :typing) (:types s t) (:predicates (p ?parm - (either s t))))", true },

		{ "(define (domain x) (:action a :parameters (?p - unreq)))", false },
		{ "(define (domain x) (:action a :parameters (?p)))", true },
		{ "(define (domain x) (:requirements :typing) (:action a :parameters (?p)))", true },
		{ "(define (domain x) (:requirements :typing) (:action a :parameters (?p - object)))", true },
		{ "(define (domain x) (:requirements :typing) (:action a :parameters (?p - undecl)))", false },
		{ "(define (domain x) (:requirements :typing) (:types t) (:action a :parameters (?p - t)))", true },
		{ "(define (domain x) (:requirements :typing) (:types t) (:action a :parameters (?p - (either t undecl) )))", false },
		{ "(define (domain x) (:requirements :typing) (:types s t) (:action a :parameters (?p - (either s t))))", true },
	}

	for _, test := range tests {
		d, err := ParseDomain("", strings.NewReader(test.pddl))
		if err != nil {
			t.Fatalf("%s\n%s", test.pddl, err)
		}
		switch err := CheckDomain(d); {
		case err != nil && test.ok:
			t.Errorf("%s\nunexpected error %s", test.pddl, err)
		case err == nil && !test.ok:
			t.Errorf("%s\nexpected error", test.pddl)
		}
	}
}

func TestParseProblem(t *testing.T) {
	file, err := os.Open(testProblemFile)
	if err != nil {
		t.Error(err)
	}
	_, err = ParseProblem(testProblemFile, file)
	if err != nil {
		t.Error(err)
	}
}

/*
func TestPrintProblem(t *testing.T) {
	file, err := os.Open(testProblemFile)
	if err != nil {
		t.Error(err)
	}
	p, err := ParseProblem(testProblemFile, file)
	if err != nil {
		t.Error(err)
	}
	PrintProblem(os.Stdout, p)
}
*/