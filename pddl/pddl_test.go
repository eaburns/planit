package pddl

import (
	"bytes"
	"os"
	"strings"
	"testing"
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

func TestPrintDomain(t *testing.T) {
	dom := `
;; openstacks, strips version

(define (domain openstacks-sequencedstrips-ADL)
  (:requirements :typing :adl :action-costs)
  (:types order product count)
  (:predicates (includes ?o - order ?p - product)
	       (waiting ?o - order)
	       (started ?o - order)
	       (shipped ?o - order)
	       (made ?p - product)
	       (stacks-avail ?s - count)
	       (next-count ?s ?ns - count))

  (:functions (total-cost) - number)
	       
  (:action make-product
    :parameters (?p - product)
    :precondition (and (not (made ?p))
		       (forall (?o - order)
			       (imply (includes ?o ?p)
				      (started ?o))))
    :effect (made ?p))

  (:action start-order
    :parameters (?o - order ?avail ?new-avail - count)
    :precondition (and (waiting ?o)
		       (stacks-avail ?avail)
		       (next-count ?new-avail ?avail))
    :effect (and (not (waiting ?o))
		 (started ?o)
		 (not (stacks-avail ?avail))
		 (stacks-avail ?new-avail))
    )

  (:action ship-order
    :parameters (?o - order ?avail ?new-avail - count)
    :precondition (and (started ?o)
		       (forall (?p - product)
			       (imply (includes ?o ?p) (made ?p)))
		       (stacks-avail ?avail)
		       (next-count ?avail ?new-avail))
    :effect (and (not (started ?o))
		 (shipped ?o)
		 (not (stacks-avail ?avail))
		 (stacks-avail ?new-avail))
    )

  (:action open-new-stack
    :parameters (?open ?new-open - count)
    :precondition (and (stacks-avail ?open)
		       (next-count ?open ?new-open))
    :effect (and (not (stacks-avail ?open))
		 (stacks-avail ?new-open) (increase (total-cost) 1))
    )

  )`
	ast, err := ParseDomain("", strings.NewReader(dom))
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer([]byte{})
	PrintDomain(buf, ast)
	if _, err := ParseDomain("", strings.NewReader(buf.String())); err != nil {
		t.Fatal(err)
	}
}

func TestCheckDomain(t *testing.T) {
	tests := [...]struct {
		pddl string
		ok   bool
	}{
		{"(define (domain x) (:requirements :strips))", true},
		{"(define (domain x) (:requirements :foobar))", false},

		{"(define (domain x) (:types t))", false},
		{"(define (domain x) (:requirements :typing) (:types t))", true},
		{"(define (domain x) (:requirements :typing) (:types t - undecl))", false},
		{"(define (domain x) (:requirements :typing) (:types s t - s))", true},
		{"(define (domain x) (:requirements :typing) (:types t - u u))", true},
		{"(define (domain x) (:requirements :typing) (:types t - object))", true},
		{"(define (domain x) (:requirements :typing) (:types u - (either s t) s t))", false},

		{"(define (domain x) (:requirements :typing) (:constants c - undecl))", false},
		{"(define (domain x) (:requirements :typing) (:constants c - object))", true},
		{"(define (domain x) (:constants c - unreqd))", false},
		{"(define (domain x) (:requirements :typing) (:types t) (:constants c - t ))", true},
		{"(define (domain x) (:requirements :typing) (:types t) (:constants c - (either t undecl) ))", false},
		{"(define (domain x) (:requirements :typing) (:types s t) (:constants c - (either s t) ))", true},

		{"(define (domain x) (:predicates (p ?parm)))", true},
		{"(define (domain x) (:predicates (p ?parm - unreqd)))", false},
		{"(define (domain x) (:requirements :typing) (:predicates (p ?parm - object)))", true},
		{"(define (domain x) (:requirements :typing) (:types t) (:predicates (p ?parm - t)))", true},
		{"(define (domain x) (:requirements :typing) (:predicates (p ?parm - undecl)))", false},
		{"(define (domain x) (:requirements :typing) (:types t) (:predicates (p ?parm - (either t undecl))))", false},
		{"(define (domain x) (:requirements :typing) (:types s t) (:predicates (p ?parm - (either s t))))", true},

		{"(define (domain x) (:action a :parameters (?p - unreq)))", false},
		{"(define (domain x) (:action a :parameters (?p)))", true},
		{"(define (domain x) (:requirements :typing) (:action a :parameters (?p)))", true},
		{"(define (domain x) (:requirements :typing) (:action a :parameters (?p - object)))", true},
		{"(define (domain x) (:requirements :typing) (:action a :parameters (?p - undecl)))", false},
		{"(define (domain x) (:requirements :typing) (:types t) (:action a :parameters (?p - t)))", true},
		{"(define (domain x) (:requirements :typing) (:types t) (:action a :parameters (?p - (either t undecl) )))", false},
		{"(define (domain x) (:requirements :typing) (:types s t) (:action a :parameters (?p - (either s t))))", true},

		{"(define (domain x) (:functions (foo)))", false},
		{"(define (domain x) (:functions (total-cost)))", false},
		{"(define (domain x) (:requirements :action-costs)))", true},
		{"(define (domain x) (:requirements :action-costs) (:functions (total-cost)))", true},
		{"(define (domain x) (:requirements :action-costs) (:functions (total-cost) - number))", true},
		{"(define (domain x) (:requirements :action-costs) (:functions (total-cost ?foo)))", false},
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

func TestPrintProblem(t *testing.T) {
	prob := `
(define (problem os-sequencedstrips-p5_1)
(:domain openstacks-sequencedstrips-ADL)
(:objects 
n0 n1 n2 n3 n4 n5  - count
o1 o2 o3 o4 o5  - order
p1 p2 p3 p4 p5  - product

)

(:init
(next-count n0 n1) (next-count n1 n2) (next-count n2 n3) (next-count n3 n4) (next-count n4 n5) 
(stacks-avail n0)

(waiting o1)
(includes o1 p2)

(waiting o2)
(includes o2 p1)(includes o2 p2)

(waiting o3)
(includes o3 p3)

(waiting o4)
(includes o4 p3)(includes o4 p4)

(waiting o5)
(includes o5 p5)

(= (total-cost) 0)

)

(:goal
(and
(shipped o1)
(shipped o2)
(shipped o3)
(shipped o4)
(shipped o5)
))

(:metric minimize (total-cost))

)`
	ast, err := ParseProblem("", strings.NewReader(prob))
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer([]byte{})
	PrintProblem(buf, ast)
	if _, err := ParseProblem("", strings.NewReader(buf.String())); err != nil {
		t.Fatal(err)
	}
}
