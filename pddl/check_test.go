package pddl

import (
	"testing"
	"strings"
	"regexp"
)

var reqsDefTests = []checkDomainTest{
	{ `(define (domain d) (:requirements :strips))`, "", nil },
	{ `(define (domain d) (:requirements :foo))`, "not supported", nil },
	{ `(define (domain d) (:requirements :strips :strips))`, "multiple", nil },
	{ `(define (domain d) (:requirements :strips :adl :strips))`, "multiple", nil },
}

func TestCheckReqsDef(t *testing.T) {
	for _, test := range reqsDefTests {
		test.run(t)
	}
}

var requirementsTests = []checkDomainTest{
	// typing
	{ `(define (domain d) (:types t))`, ":typing", nil },
	{ `(define (domain d) (:requirements :typing) (:types t))`, "", nil },
	{ `(define (domain d) (:requirements :adl) (:types t))`, "", nil },

	// negative-preconditions
	{ `(define (domain d)
		(:predicates (p))
		(:action a :parameters () :precondition (not (p))))`,
		":negative-preconditions", nil },
	{ `(define (domain d)
		(:requirements :negative-preconditions)
		(:predicates (p))
		(:action a :parameters () :precondition (not (p))))`,
		"", nil },

	// disjunctive-preconditions
	{ `(define (domain d)
		(:predicates (p))
		(:action a :parameters () :precondition (or (p) (p))))`,
		":disjunctive-preconditions", nil },
	{ `(define (domain d)
		(:predicates (p))
		(:action a :parameters () :precondition (imply (p) (p))))`,
		":disjunctive-preconditions", nil },
	{ `(define (domain d)
		(:predicates (p))
		(:action a :parameters () :precondition (not (and (p) (p)))))`,
		":disjunctive-preconditions", nil },
	{ `(define (domain d)
		(:requirements :disjunctive-preconditions)
		(:predicates (p))
		(:action a :parameters () :precondition (or (p) (p))))`,
		"", nil },
	{ `(define (domain d)
		(:requirements :disjunctive-preconditions)
		(:predicates (p))
		(:action a :parameters () :precondition (imply (p) (p))))`,
		"", nil },
	{ `(define (domain d)
		(:requirements :disjunctive-preconditions)
		(:predicates (p))
		(:action a :parameters () :precondition (not (and (p) (p)))))`,
		"", nil },

	// equality
/*	// This doesn't parse since = is not allowed as an identifier…
	{ `(define (domain d)
		(:constants c)
		(:action a :parameters () :precondition (=  c c)))`,
		"", nil },
*/

	// universal-preconditions
	{ `(define (domain d)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (forall (?x) (p ?x))))`,
		":universal-preconditions", nil },
	{ `(define (domain d)
		(:requirements :universal-preconditions)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (forall (?x) (p ?x))))`,
		"", nil },
	{ `(define (domain d)
		(:requirements :quantified-preconditions)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (forall (?x) (p ?x))))`,
		"", nil },

	// existential-preconditions
	{ `(define (domain d)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (exists (?x) (p ?x))))`,
		":existential-preconditions", nil },
	{ `(define (domain d)
		(:requirements :existential-preconditions)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (exists (?x) (p ?x))))`,
		"", nil },
	{ `(define (domain d)
		(:requirements :quantified-preconditions)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (exists (?x) (p ?x))))`,
		"", nil },

	// conditional-effects
	{ `(define (domain d)
		(:predicates (p ?x))
		(:action a :parameters () :effect (forall (?x) (p ?x))))`,
		":conditional-effects", nil },
	{ `(define (domain d)
		(:predicates (p) (q))
		(:action a :parameters () :effect (when (p) (q))))`,
		":conditional-effects", nil },
	{ `(define (domain d)
		(:requirements :conditional-effects)
		(:predicates (p ?x))
		(:action a :parameters () :effect (forall (?x) (p ?x))))`,
		"", nil },
	{ `(define (domain d)
		(:requirements :conditional-effects)
		(:predicates (p) (q))
		(:action a :parameters () :effect (when (p) (q))))`,
		"", nil },

	// :action-costs
	{ `(define (domain d)
		(:functions (total-cost)))`,
		":action-costs", nil },
	{ `(define (domain d)
		(:predicates (p) (q))
		(:action a :parameters () :effect (increase total-cost 1)))`,
		":action-costs", nil },
	{ `(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost ?x))
		(:action a :parameters (?x) :effect (increase (total-cost ?x) 1)))`,
		"0-ary total-cost", nil },
	{ `(define (domain d)
		(:requirements :action-costs)
		(:functions (foo-bar))
		(:action a :parameters (?x) :effect (increase foo-bar 1)))`,
		"0-ary total-cost", nil },
/*
	// Why does the parser reject a negative number?
	{ `(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost))
		(:action a :parameters (?x) :effect (increase total-cost -1)))`,
		"negative", nil },
*/
	{ `(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost))
		(:action a :parameters (?x) :effect (increase total-cost (total-cost))))`,
		"total-cost", nil },
	{ `(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost)))`,
		"", nil },
	{ `(define (domain d)
		(:requirements :action-costs)
		(:predicates (p) (q))
		(:functions (total-cost))
		(:action a :parameters () :effect (increase total-cost 1)))`,
		"", nil },
	{ `(define (domain d)
		(:requirements :action-costs)
		(:predicates (p) (q))
		(:functions (total-cost))
		(:action a :parameters () :effect (increase total-cost 500)))`,
		"", nil },
	{ `(define (domain d)
		(:requirements :action-costs)
		(:predicates (p) (q))
		(:functions (total-cost) (f))
		(:action a :parameters () :effect (increase total-cost (f))))`,
		"", nil },
	{ `(define (domain d)
		(:requirements :action-costs)
		(:predicates (p) (q))
		(:functions (total-cost) (f ?x))
		(:action a :parameters (?x) :effect (increase total-cost (f ?x))))`,
		"", nil },
}

func TestRequirements(t *testing.T) {
	for _, test := range requirementsTests {
		test.run(t)
	}
}

var typesDefTests = []checkDomainTest{
	// undefined type
	{ `(define (domain d) (:requirements :typing) (:types t - s))`, "undefined", nil },

	// object is not undefined
	{ `(define (domain d) (:requirements :typing) (:types t - object))`, "", nil },

	// multiple type definitions
	{ `(define (domain d) (:requirements :typing) (:types t t))`, "multiple", nil },
	{ `(define (domain d) (:requirements :typing) (:types t s t))`, "multiple", nil },

	// Everything's OK
	{ `(define (domain d) (:requirements :typing) (:types t))`, "", nil },
	{ `(define (domain d) (:requirements :typing) (:types t s))`, "", nil },
	{ `(define (domain d) (:requirements :typing) (:types t s - object))`, "", nil },
	{ `(define (domain d) (:requirements :typing) (:types t - s s - object))`, "", nil },

	// OK, but make sure we only have the types we expect.
	{ `(define (domain d) (:requirements :typing) (:types object))`, "",
		checkTypes([]string{"object"}) },
	{ `(define (domain d) (:requirements :typing))`, "",
		checkTypes([]string{"object"}) },
	{ `(define (domain d) (:requirements :typing) (:types t))`, "",
		checkTypes([]string{"object", "t"}) },

	// OK, but make sure we only have the super typse we expect.
	{ `(define (domain d))`, "",
		checkSupers("object", []string{"object"}) },
	{ `(define (domain d) (:requirements :typing) (:types object))`, "",
		checkSupers("object", []string{"object"}) },
	{ `(define (domain d) (:requirements :typing) (:types t))`, "",
		checkSupers("t", []string{"t", "object"}) },
	{ `(define (domain d) (:requirements :typing) (:types t - s s))`, "",
		checkSupers("t", []string{"t", "s", "object"}) },
	{ `(define (domain d) (:requirements :typing) (:types t - s s - t))`, "",
		checkSupers("t", []string{"t", "s", "object"}) },
	{ `(define (domain d) (:requirements :typing) (:types t - s s - u u))`, "",
		checkSupers("t", []string{"t", "s", "u", "object"}) },
	{ `(define (domain d) (:requirements :typing) (:types t - s s - u u))`, "",
		checkSupers("s", []string{"s", "u", "object"}) },
}

// checkTypes returns a function that scans through
// a domain's type list looking for each type in the
// given slice.
func checkTypes(types []string) func(string, *Domain, *testing.T) {
	return func(pddl string, d *Domain, t *testing.T) {
		if len(types) != len(d.Types) {
			t.Errorf("%s\nexpected %d types, got %d: %v",
				pddl, len(types), len(d.Types), d.Types)
			return
		}
		for _, typ := range types {
			if findType(typ, d.Types) != nil {
				continue
			}
			t.Errorf("%s\nexpected type %s: %v", pddl, typ, d.Types)
			return
		}
	}
}

// checkSupers returns a function that checks
// that the supers of the named type match
// the list of super types.
func checkSupers(name string, supers []string) func(string, *Domain, *testing.T) {
	return func(pddl string, d *Domain, t *testing.T) {
		typ := findType(name, d.Types)
		if typ == nil {
			t.Fatalf("%s\ntype %s not found", pddl, name)
		}
		if len(typ.Supers) != len(supers) {
			t.Errorf("%s\nincorrect number of super types, expected %d: %v",
				pddl, len(supers), typ.Supers)
			return
		}
		for _, s := range supers {
			if findTypePtr(s, typ.Supers) == nil {
				t.Errorf("%s\nexpected super type %s: %v",
					pddl, s, typ.Supers)
				return
			}
		}
	}
}

func findType(name string, ts []Type) *Type {
	for i := range ts {
		if ts[i].Str == name {
			return &ts[i]
		}
	}
	return nil
}

func findTypePtr(name string, ts []*Type) *Type {
	for i := range ts {
		if ts[i].Str == name {
			return ts[i]
		}
	}
	return nil
}

func TestCheckTypesDef(t *testing.T) {
	for _, test := range typesDefTests {
		test.run(t)
	}
}

type checkDomainTest struct {
	pddl string
	errorRegexp string
	test func(string, *Domain, *testing.T)
}

func (c checkDomainTest) run(t *testing.T) {
	d, p, err := Parse("", strings.NewReader(c.pddl))
	if p != nil {
		t.Fatalf("%s\nis a problem, not a domain", c.pddl)
	}
	if err != nil {
		t.Fatalf("%s\nparse error: %s", c.pddl, err)
	}
	switch err := Check(d, nil); {
	case err == nil && c.errorRegexp == "":
		if c.test != nil {
			c.test(c.pddl, d, t)
		}
	case err == nil && c.errorRegexp != "":
		t.Errorf("%s\nexpected error matching '%s'", c.pddl, c.errorRegexp)
	case err != nil && c.errorRegexp == "":
		t.Errorf("%s\nunexpected error '%s'", c.pddl, err)
	case err != nil && c.errorRegexp != "":
		re := regexp.MustCompile(c.errorRegexp)
		if !re.Match([]byte(err.Error())) {
			t.Errorf("%s\nexpected error matching '%s', got '%s'",
				c.pddl, c.errorRegexp, err.Error())
		}
	}
}