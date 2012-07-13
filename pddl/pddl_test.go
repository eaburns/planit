package pddl

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestPrintDomain(t *testing.T) {
	dom, err := Parse("", strings.NewReader(dom))
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer([]byte{})
	PrintDomain(buf, dom.(*Domain))
	if _, err := Parse("", strings.NewReader(buf.String())); err != nil {
		t.Fatal(err)
	}
}

type test struct {
	pddl   string
	errMsg string
}

// checkPddlDomain checks a set of tests by
// calling CheckDomain on the pddl and verifying
// that the error message matches the regular
// expression.
func checkPddlDomain(tests []test, t *testing.T) {
	for _, test := range tests {
		dom, err := Parse("", strings.NewReader(test.pddl))
		if err != nil {
			t.Errorf("%s\n%s", test.pddl, err)
			continue
		}
		err = Check(dom.(*Domain), nil)
		if test.errMsg == "" {
			if err != nil {
				t.Errorf("%s\nunexpected error message: %s",
					test.pddl, err.Error())
			}
			continue
		}
		if err == nil {
			t.Errorf("%s\nexpected error message matching: %s",
				test.pddl, test.errMsg)
			continue
		}
		re := regexp.MustCompile(test.errMsg)
		if !re.Match([]byte(err.Error())) {
			t.Errorf("%s\nexpected error message matching %s, got %s",
				test.pddl, test.errMsg, err.Error())
		}
	}
}

func TestCheckConstantsDef(t *testing.T) {
	checkPddlDomain([]test{
		{`(define (domain x) (:requirements :typing) (:constants c c))`,
			"multiple"},
		{`(define (domain x) (:requirements :typing) (:constants c - undecl))`,
			"undefined type: undecl"},
		{`(define (domain x) (:requirements :typing) (:constants c - object))`, ""},
		{`(define (domain x) (:constants c - unreqd))`,
			":typing is not required"},
		{`(define (domain x) (:requirements :typing) (:types t) (:constants c - t ))`, ""},
		{`(define (domain x) (:requirements :typing) (:types t) (:constants c - (either t undecl) ))`,
			"undefined type: undecl"},
		{`(define (domain x) (:requirements :typing) (:types s t) (:constants c - (either s t) ))`, ""},
	}, t)
}

func TestCheckPredicatesDef(t *testing.T) {
	checkPddlDomain([]test{
		{`(define (domain x) (:predicates (p ?parm)))`, ""},
		{`(define (domain x) (:predicates (p ?parm - unreqd)))`,
			":typing is not required"},
		{`(define (domain x) (:requirements :typing) (:predicates (p ?parm - object)))`, ""},
		{`(define (domain x) (:requirements :typing) (:types t) (:predicates (p ?parm - t)))`, ""},
		{`(define (domain x) (:requirements :typing) (:predicates (p ?parm - undecl)))`,
			"undefined type: undecl"},
		{`(define (domain x) (:requirements :typing) (:types t) (:predicates (p ?parm - (either t undecl))))`,
			"undefined type: undecl"},
		{`(define (domain x) (:requirements :typing) (:types s t) (:predicates (p ?parm - (either s t))))`, ""},
	}, t)
}

func TestCheckActionDefs(t *testing.T) {
	checkPddlDomain([]test{
		{`(define (domain x) (:action a :parameters (?p - unreq)))`,
			":typing is not required"},
		{`(define (domain x) (:action a :parameters (?p)))`, ""},
		{`(define (domain x) (:requirements :typing) (:action a :parameters (?p)))`, ""},
		{`(define (domain x) (:requirements :typing) (:action a :parameters (?p - object)))`, ""},
		{`(define (domain x) (:requirements :typing) (:action a :parameters (?p - undecl)))`,
			"undefined type: undecl"},
		{`(define (domain x) (:requirements :typing) (:types t) (:action a :parameters (?p - t)))`, ""},
		{`(define (domain x) (:requirements :typing) (:types t)
			(:action a :parameters (?p - (either t undecl) )))`,
			"undefined type: undecl"},
		{`(define (domain x) (:requirements :typing) (:types s t)
			(:action a :parameters (?p - (either s t))))`, ""},
	}, t)
}

func TestCheckQuantifiers(t *testing.T) {
	checkPddlDomain([]test{
		{`(define (domain x) (:requirements :universal-preconditions)
			(:action a :parameters () :precondition (forall (?v - notypes) (and))))`,
			":typing is not required"},
		{`(define (domain x) (:requirements :typing :universal-preconditions)
			(:action a :parameters () :precondition (forall (?v - undef) (and))))`,
			"undefined type: undef"},
		{`(define (domain x) (:requirements :typing :universal-preconditions) (:types t)
			(:action a :parameters () :precondition (forall (?v - t) (and))))`,
			""},
	}, t)
}

func TestCheckProposition(t *testing.T) {
	checkPddlDomain([]test{
		{`(define (domain x) (:predicates (p)) (:action a :parameters () :precondition (p)))`, ""},
		{`(define (domain x) (:action a :parameters () :precondition (p)))`,
			"undefined predicate: p"},
		{`(define (domain x) (:predicates (p ?x)) (:action a :parameters () :precondition (p ?x)))`,
			"undefined variable: \\?x"},
		{`(define (domain x) (:predicates (p ?x)) (:action a :parameters () :precondition (p x)))`,
			"undefined constant: x"},
		{`(define (domain x) (:predicates (p ?x)) (:action a :parameters (?x) :precondition (p ?x)))`, ""},
		{`(define (domain x) (:requirements :universal-preconditions) (:predicates (p ?x))
			(:action a :parameters () :precondition (forall (?x) (p ?x))))`, ""},
		{`(define (domain x) (:constants c) (:predicates (p ?x))
			(:action a :parameters () :precondition (p c)))`, ""},
		{`(define (domain x) (:constants c) (:predicates (p))
			(:action a :parameters () :precondition (p c)))`,
			"requires 0 arguments"},
		{`(define (domain x) (:constants c d) (:predicates (p ?x))
			(:action a :parameters () :precondition (p c d)))`,
			"requires 1 argument"},
		{`(define (domain x) (:requirements :typing) (:types t)
			(:constants c - t) (:predicates (p ?x - t))
			(:action a :parameters () :precondition (p c)))`, ""},
		{`(define (domain x) (:requirements :typing) (:types s t)
			(:constants c - t) (:predicates (p ?x - (either t s)))
			(:action a :parameters () :precondition (p c)))`, ""},
		{`(define (domain x) (:requirements :typing) (:types s t)
			(:constants c - (either t s)) (:predicates (p ?x - (either t s)))
			(:action a :parameters () :precondition (p c)))`, ""},
		{`(define (domain x) (:requirements :typing) (:types s t)
			(:constants c - t) (:predicates (p ?x - s))
			(:action a :parameters () :precondition (p c)))`,
			"incompatible"},
		{`(define (domain x) (:requirements :typing) (:types s t)
			(:constants c - (either t s)) (:predicates (p ?x - t))
			(:action a :parameters () :precondition (p c)))`,
			"incompatible"},
		{`(define (domain x) (:requirements :typing) (:types s t u)
			(:constants c - (either t s)) (:predicates (p ?x - (either s t u)))
			(:action a :parameters () :precondition (p c)))`,
			""},
		{`(define (domain x) (:requirements :typing) (:types s t u)
			(:constants c - (either t s)) (:predicates (p ?x - (either t u)))
			(:action a :parameters () :precondition (p c)))`,
			"incompatible"},
		{`(define (domain x) (:constants c) (:predicates (p ?x))
			(:action a :parameters () :precondition (p c)))`,
			""},
	}, t)
}

func TestPrintProblem(t *testing.T) {
	pro, err := Parse("", strings.NewReader(prob))
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer([]byte{})
	PrintProblem(buf, pro.(*Problem))
	if _, err := Parse("", strings.NewReader(buf.String())); err != nil {
		t.Fatal(err)
	}
}

func TestCheck(t *testing.T) {
	dom, err := Parse("<domain>", strings.NewReader(dom))
	if err != nil {
		t.Fatal(err)
	}
	pro, err := Parse("<problem>", strings.NewReader(prob))
	if err != nil {
		t.Fatal(err)
	}
	if err := Check(dom.(*Domain), pro.(*Problem)); err != nil {
		t.Error(err)
	}
}

const (
	dom = `
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
	prob = `
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
)
