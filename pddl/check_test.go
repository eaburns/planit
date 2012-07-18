package pddl

import (
	"regexp"
	"strings"
	"testing"
)

var reqsDefTests = []checkDomainTest{
	{`(define (domain d) (:requirements))`, "", nil},
	{`(define (domain d) (:requirements :strips))`, "", nil},
	{`(define (domain d) (:requirements :foo))`, "not supported", nil},
	{`(define (domain d) (:requirements :strips :strips))`, "multiple", nil},
	{`(define (domain d) (:requirements :strips :adl :strips))`, "multiple", nil},
}

func TestCheckReqsDef(t *testing.T) {
	for _, test := range reqsDefTests {
		test.run(t)
	}
}

var requirementsTests = []checkDomainTest{
	// typing
	{`(define (domain d) (:types t))`, ":typing", nil},
	{`(define (domain d) (:requirements :typing) (:types t))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types t))`, "", nil},

	// negative-preconditions
	{`(define (domain d)
		(:predicates (p))
		(:action a :parameters () :precondition (not (p))))`,
		":negative-preconditions", nil},
	{`(define (domain d)
		(:requirements :negative-preconditions)
		(:predicates (p))
		(:action a :parameters () :precondition (not (p))))`,
		"", nil},

	// disjunctive-preconditions
	{`(define (domain d)
		(:predicates (p))
		(:action a :parameters () :precondition (or (p) (p))))`,
		":disjunctive-preconditions", nil},
	{`(define (domain d)
		(:predicates (p))
		(:action a :parameters () :precondition (imply (p) (p))))`,
		":disjunctive-preconditions", nil},
	{`(define (domain d)
		(:predicates (p))
		(:action a :parameters () :precondition (not (and (p) (p)))))`,
		":disjunctive-preconditions", nil},
	{`(define (domain d)
		(:requirements :disjunctive-preconditions)
		(:predicates (p))
		(:action a :parameters () :precondition (or (p) (p))))`,
		"", nil},
	{`(define (domain d)
		(:requirements :disjunctive-preconditions)
		(:predicates (p))
		(:action a :parameters () :precondition (imply (p) (p))))`,
		"", nil},
	{`(define (domain d)
		(:requirements :disjunctive-preconditions)
		(:predicates (p))
		(:action a :parameters () :precondition (not (and (p) (p)))))`,
		"", nil},

	// equality
	{`(define (domain d)
		(:constants c)
		(:action a :parameters () :precondition (=  c c)))`,
		"undefined", nil},
	{`(define (domain d)
		(:requirements :equality)
		(:constants c)
		(:action a :parameters () :precondition (=  c c)))`,
		"", nil},

	// universal-preconditions
	{`(define (domain d)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (forall (?x) (p ?x))))`,
		":universal-preconditions", nil},
	{`(define (domain d)
		(:requirements :universal-preconditions)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (forall (?x) (p ?x))))`,
		"", nil},
	{`(define (domain d)
		(:requirements :quantified-preconditions)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (forall (?x) (p ?x))))`,
		"", nil},

	// existential-preconditions
	{`(define (domain d)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (exists (?x) (p ?x))))`,
		":existential-preconditions", nil},
	{`(define (domain d)
		(:requirements :existential-preconditions)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (exists (?x) (p ?x))))`,
		"", nil},
	{`(define (domain d)
		(:requirements :quantified-preconditions)
		(:predicates (p ?x))
		(:action a :parameters () :precondition (exists (?x) (p ?x))))`,
		"", nil},

	// conditional-effects
	{`(define (domain d)
		(:predicates (p ?x))
		(:action a :parameters () :effect (forall (?x) (p ?x))))`,
		":conditional-effects", nil},
	{`(define (domain d)
		(:predicates (p) (q))
		(:action a :parameters () :effect (when (p) (q))))`,
		":conditional-effects", nil},
	{`(define (domain d)
		(:requirements :conditional-effects)
		(:predicates (p ?x))
		(:action a :parameters () :effect (forall (?x) (p ?x))))`,
		"", nil},
	{`(define (domain d)
		(:requirements :conditional-effects)
		(:predicates (p) (q))
		(:action a :parameters () :effect (when (p) (q))))`,
		"", nil},

	// :action-costs
	{`(define (domain d)
		(:functions (total-cost)))`,
		":action-costs", nil},
	{`(define (domain d)
		(:predicates (p) (q))
		(:action a :parameters () :effect (increase total-cost 1)))`,
		":action-costs", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost ?x))
		(:action a :parameters (?x) :effect (increase (total-cost ?x) 1)))`,
		"0-ary total-cost", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:functions (foo-bar))
		(:action a :parameters (?x) :effect (increase foo-bar 1)))`,
		"0-ary total-cost", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost))
		(:action a :parameters (?x) :effect (increase total-cost -1)))`,
		"negative", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost))
		(:action a :parameters (?x) :effect (increase total-cost -5)))`,
		"negative", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost))
		(:action a :parameters (?x) :effect (increase total-cost --1)))`,
		"", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost))
		(:action a :parameters (?x) :effect (increase total-cost (total-cost))))`,
		"total-cost", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:functions (total-cost)))`,
		"", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:predicates (p) (q))
		(:functions (total-cost))
		(:action a :parameters () :effect (increase total-cost 1)))`,
		"", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:predicates (p) (q))
		(:functions (total-cost))
		(:action a :parameters () :effect (increase total-cost 500)))`,
		"", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:predicates (p) (q))
		(:functions (total-cost) (f))
		(:action a :parameters () :effect (increase total-cost (f))))`,
		"", nil},
	{`(define (domain d)
		(:requirements :action-costs)
		(:predicates (p) (q))
		(:functions (total-cost) (f ?x))
		(:action a :parameters (?x) :effect (increase total-cost (f ?x))))`,
		"", nil},
}

func TestRequirements(t *testing.T) {
	for _, test := range requirementsTests {
		test.run(t)
	}
}

var typesDefTests = []checkDomainTest{
	{`(define (domain d) (:types))`, "", nil},

	// undefined type
	{`(define (domain d) (:requirements :typing) (:types t - s))`, "undefined", nil},

	// object is not undefined
	{`(define (domain d) (:requirements :typing) (:types t - object))`, "", nil},

	// multiple type definitions
	{`(define (domain d) (:requirements :typing) (:types t t))`, "multiple", nil},
	{`(define (domain d) (:requirements :typing) (:types t s t))`, "multiple", nil},

	// Everything's OK
	{`(define (domain d) (:requirements :typing) (:types t))`, "", nil},
	{`(define (domain d) (:requirements :typing) (:types t s))`, "", nil},
	{`(define (domain d) (:requirements :typing) (:types t s - object))`, "", nil},
	{`(define (domain d) (:requirements :typing) (:types t - s s - object))`, "", nil},

	// OK, but make sure we only have the types we expect.
	{`(define (domain d) (:requirements :typing) (:types object))`, "",
		checkTypes([]string{"object"})},
	{`(define (domain d) (:requirements :typing))`, "",
		checkTypes([]string{"object"})},
	{`(define (domain d) (:requirements :typing) (:types t))`, "",
		checkTypes([]string{"object", "t"})},

	// OK, but make sure we only have the super typse we expect.
	{`(define (domain d))`, "",
		checkSupers("object", []string{"object"})},
	{`(define (domain d) (:requirements :typing) (:types object))`, "",
		checkSupers("object", []string{"object"})},
	{`(define (domain d) (:requirements :typing) (:types t))`, "",
		checkSupers("t", []string{"t", "object"})},
	{`(define (domain d) (:requirements :typing) (:types t - s s))`, "",
		checkSupers("t", []string{"t", "s", "object"})},
	{`(define (domain d) (:requirements :typing) (:types t - s s - t))`, "",
		checkSupers("t", []string{"t", "s", "object"})},
	{`(define (domain d) (:requirements :typing) (:types t - s s - u u))`, "",
		checkSupers("t", []string{"t", "s", "u", "object"})},
	{`(define (domain d) (:requirements :typing) (:types t - s s - u u))`, "",
		checkSupers("s", []string{"s", "u", "object"})},
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

var constsDefTests = []checkDomainTest{
	{`(define (domain d) (:constants))`, "", nil},
	{`(define (domain d) (:constants a b c))`, "", nil},
	{`(define (domain d) (:constants a b c a))`, "multiple", nil},
	{`(define (domain d) (:requirements :typing) (:constants a - t))`, "undefined", nil},
	{`(define (domain d) (:constants a b))`, "",
		domainChecks(checkTypeDomain("object", []string{"a", "b"}),
			checkConstsTypes("a", []string{"object"}),
			checkConstsTypes("b", []string{"object"}))},
	{`(define (domain d) (:requirements :typing) (:constants a b - object))`, "",
		domainChecks(checkTypeDomain("object", []string{"a", "b"}),
			checkConstsTypes("a", []string{"object"}),
			checkConstsTypes("b", []string{"object"}))},
	{`(define (domain d) (:requirements :typing) (:types t) (:constants a - t))`, "",
		domainChecks(checkTypeDomain("object", []string{"a"}),
			checkTypeDomain("t", []string{"a"}),
			checkConstsTypes("a", []string{"t"}))},
	{`(define (domain d) (:requirements :typing) (:types t - s s) (:constants a - t))`, "",
		domainChecks(checkTypeDomain("object", []string{"a"}),
			checkTypeDomain("t", []string{"a"}),
			checkTypeDomain("s", []string{"a"}),
			checkConstsTypes("a", []string{"t"}))},
	{`(define (domain d) (:requirements :typing) (:types s t) (:constants a - (either s t)))`, "",
		domainChecks(checkTypeDomain("object", []string{"a"}),
			checkTypeDomain("t", []string{"a"}),
			checkTypeDomain("s", []string{"a"}),
			checkConstsTypes("a", []string{"t", "s"}))},
	{`(define (domain d) (:requirements :typing) (:types t - s s) (:constants a - (either s t)))`, "",
		domainChecks(checkTypeDomain("object", []string{"a"}),
			checkTypeDomain("t", []string{"a"}),
			checkTypeDomain("s", []string{"a"}),
			checkConstsTypes("a", []string{"t", "s"}))},
	{`(define (domain d) (:requirements :typing) (:types t) (:constants a b - t))`, "",
		domainChecks(checkTypeDomain("object", []string{"a", "b"}),
			checkTypeDomain("t", []string{"a", "b"}),
			checkConstsTypes("a", []string{"t"}),
			checkConstsTypes("b", []string{"t"}))},
	{`(define (domain d) (:requirements :typing) (:types t) (:constants a - t b))`, "",
		domainChecks(checkTypeDomain("object", []string{"a", "b"}),
			checkTypeDomain("t", []string{"a"}),
			checkConstsTypes("a", []string{"t"}),
			checkConstsTypes("b", []string{"object"}))},
}

// checkTypeDomain returns a function that
// checks that the given type has all of the
// assigned constants.
func checkTypeDomain(tName string, consts []string) func(string, *Domain, *testing.T) {
	return func(pddl string, d *Domain, t *testing.T) {
		typ := findType(tName, d.Types)
		if typ == nil {
			t.Fatalf("%s\ntype %s: not found", pddl, tName)
		}
		for _, con := range consts {
			found := false
			for _, obj := range typ.Domain {
				if con == obj.Str {
					if found {
						t.Errorf("%s\ntype %s: has constant %s multiple times",
							pddl, typ, obj)
					}
					found = true
				}
			}
			if !found {
				t.Errorf("%s\ntype %s: no object %s", pddl, tName, con)
			}
		}
	}
}

// checkConstsTypes returns a function
// that checks the types assigned to the
// given constant in a consts definition.
func checkConstsTypes(eName string, types []string) func(string, *Domain, *testing.T) {
	return func(pddl string, d *Domain, t *testing.T) {
		checkEntryTypes(pddl, eName, types, d.Constants, t)
	}
}

// checkEntryTypes checks that the given
// typed entry has all of the given types.
func checkEntryTypes(pddl, eName string, types []string, lst []TypedEntry, t *testing.T) {
	ent := findTypedEntry(eName, lst)
	if ent == nil {
		t.Fatalf("%s\ntyped entry %s: not found", pddl, eName)
	}
	if len(ent.Types) != len(types) {
		t.Fatalf("typed entry %s: expected %d types, got %d",
			eName, len(types), len(ent.Types))
	}
	for _, typ := range types {
		found := false
		for _, tp := range ent.Types {
			if tp.Str == typ {
				if tp.Definition.Str != typ {
					t.Errorf("%s\ntyped entry %s: type %s linked to wrong definition: %s",
						pddl, eName, typ, tp.Definition.Str)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s\ntyped entry %s: missing type %s", pddl, eName, typ)
		}
	}
}

// findTypedEntry returns a pointer to
// the typed entry with the given name,
// or nil if there is no entry with the
// given name.
func findTypedEntry(eName string, lst []TypedEntry) *TypedEntry {
	for i := range lst {
		if lst[i].Str == eName {
			return &lst[i]
		}
	}
	return nil
}

// domainChecks returns a domain checking function
// that sequences two domain checking
// functions.
func domainChecks(fs ...func(string, *Domain, *testing.T)) func(string, *Domain, *testing.T) {
	return func(pddl string, d *Domain, t *testing.T) {
		for _, f := range fs {
			f(pddl, d, t)
		}
	}
}

func TestCheckConstsDef(t *testing.T) {
	for _, test := range constsDefTests {
		test.run(t)
	}
}

var predsDefTests = []checkDomainTest{
	{`(define (domain d) (:predicates (p)))`, "", nil},
	{`(define (domain d) (:predicates (p) (q)))`, "", nil},
	{`(define (domain d) (:predicates (p ?a ?b)))`, "",
		domainChecks(checkPredParamTypes("p", "?a", []string{"object"}),
			checkPredParamTypes("p", "?b", []string{"object"}))},
	{`(define (domain d) (:predicates (p ?a ?a)))`, "multiple", nil},
	{`(define (domain d) (:requirements :typing) (:types t) (:predicates (p ?a - t)))`, "",
		checkPredParamTypes("p", "?a", []string{"t"})},
	{`(define (domain d) (:requirements :typing) (:types s t) (:predicates (p ?a - (either s t))))`, "",
		checkPredParamTypes("p", "?a", []string{"t", "s"})},
	{`(define (domain d) (:requirements :typing) (:types t) (:predicates (p ?a - t ?b)))`, "",
		domainChecks(checkPredParamTypes("p", "?a", []string{"t"}),
			checkPredParamTypes("p", "?b", []string{"object"}))},

	// undefined paramater types
	{`(define (domain d) (:requirements :typing) (:predicates (p ?a - t)))`, "undefined", nil},
}

// checkPredParamTypes returns a function
// that checks the types assigned to given
// predicate's parameter.
func checkPredParamTypes(pred string, parm string, types []string) func(string, *Domain, *testing.T) {
	return func(pddl string, d *Domain, t *testing.T) {
		p := findPred(pred, d.Predicates)
		if p == nil {
			t.Fatalf("%s\npredicate %s: not found", pddl, pred)
		}
		checkEntryTypes(pddl, parm, types, p.Parameters, t)
	}
}

// findPred returns a pointer to the predicate with
// the matching name, if one exists.
func findPred(pred string, preds []Predicate) *Predicate {
	for i := range preds {
		if preds[i].Str == pred {
			return &preds[i]
		}
	}
	return nil
}

func TestCheckPredsDef(t *testing.T) {
	for _, test := range predsDefTests {
		test.run(t)
	}
}

var funcsDefTests = []checkDomainTest{
	{`(define (domain d) (:requirements :action-costs) (:functions (f)))`, "", nil},
	{`(define (domain d) (:requirements :action-costs) (:functions (f) (g)))`, "", nil},
	{`(define (domain d) (:requirements :action-costs) (:functions (f ?a ?b)))`, "",
		domainChecks(checkFuncParamTypes("f", "?a", []string{"object"}),
			checkFuncParamTypes("f", "?b", []string{"object"}))},
	{`(define (domain d) (:requirements :action-costs) (:functions (f ?a ?a)))`, "multiple", nil},

	// ensure parameter types
	{`(define (domain d) (:requirements :typing :action-costs) (:types t) (:functions (f ?a - t)))`, "",
		checkFuncParamTypes("f", "?a", []string{"t"})},
	{`(define (domain d) (:requirements :typing :action-costs) (:types s t) (:functions (f ?a - (either s t))))`, "",
		checkFuncParamTypes("f", "?a", []string{"t", "s"})},
	{`(define (domain d) (:requirements :typing :action-costs) (:types t) (:functions (f ?a - t ?b)))`, "",
		domainChecks(checkFuncParamTypes("f", "?a", []string{"t"}),
			checkFuncParamTypes("f", "?b", []string{"object"}))},

	// undefined paramater types
	{`(define (domain d) (:requirements :typing :action-costs) (:functions (f ?a - t)))`, "undefined", nil},
}

// checkFuncParamTypes returns a function
// that checks the types assigned to given
// functions's parameter.
func checkFuncParamTypes(fun string, parm string, types []string) func(string, *Domain, *testing.T) {
	return func(pddl string, d *Domain, t *testing.T) {
		f := findFunc(fun, d.Functions)
		if f == nil {
			t.Fatalf("%s\nfunction %s: not found", pddl, fun)
		}
		checkEntryTypes(pddl, parm, types, f.Parameters, t)
	}
}

// findFunc returns a pointer to the function
// with the matching name, if one exists.
func findFunc(fun string, funcs []Function) *Function {
	for i := range funcs {
		if funcs[i].Str == fun {
			return &funcs[i]
		}
	}
	return nil
}

func TestCheckFuncsDef(t *testing.T) {
	for _, test := range funcsDefTests {
		test.run(t)
	}
}

var actionsDefTests = []checkDomainTest{
	{`(define (domain d) (:action a :parameters ()))`, "", nil},
	{`(define (domain d) (:action a :parameters (?a)))`, "",
		checkActionParamTypes("a", "?a", []string{"object"})},

	{`(define (domain d) (:action a :parameters (?a ?b)))`, "",
		domainChecks(checkActionParamTypes("a", "?a", []string{"object"}),
			checkActionParamTypes("a", "?b", []string{"object"}))},

	{`(define (domain d) (:action a :parameters (?a ?a)))`, "multiple", nil},

	{`(define (domain d) (:requirements :typing) (:types t) (:action a :parameters (?a - t ?b)))`, "",
		domainChecks(checkActionParamTypes("a", "?a", []string{"t"}),
			checkActionParamTypes("a", "?b", []string{"object"}))},

	{`(define (domain d) (:requirements :typing) (:types t s) (:action a :parameters (?a - (either s t))))`, "",
		checkActionParamTypes("a", "?a", []string{"t", "s"})},

	// undefined action parameter type
	{`(define (domain d) (:requirements :typing) (:action a :parameters (?a - t)))`, "undefined", nil},
}

// checkActionParamTypes returns a function
// that checks the types assigned to given
// action's parameter.
func checkActionParamTypes(act string, parm string, types []string) func(string, *Domain, *testing.T) {
	return func(pddl string, d *Domain, t *testing.T) {
		a := findAction(act, d.Actions)
		if a == nil {
			t.Fatalf("%s\naction %s: not found", pddl, act)
		}
		checkEntryTypes(pddl, parm, types, a.Parameters, t)
	}
}

// findAction returns a pointer to the action
// with the matching name, if one exists.
func findAction(act string, acts []Action) *Action {
	for i := range acts {
		if acts[i].Str == act {
			return &acts[i]
		}
	}
	return nil
}

func TestCheckActionsDef(t *testing.T) {
	for _, test := range actionsDefTests {
		test.run(t)
	}
}

var quantNodeTests = []checkDomainTest{
	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (forall () (and))))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (forall (?a) (and))))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (forall (?a ?b) (and))))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (forall (?a ?a) (and))))`, "multiple", nil},
	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (forall (?a - t) (and))))`, "undefined", nil},
	{`(define (domain d) (:requirements :adl) (:types t) (:action a :parameters ()
		:precondition (forall (?a - (either s t)) (and))))`, "undefined", nil},

	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (exists () (and))))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (exists (?a) (and))))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (exists (?a ?b) (and))))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (exists (?a ?a) (and))))`, "multiple", nil},
	{`(define (domain d) (:requirements :adl) (:action a :parameters ()
		:precondition (exists (?a - t) (and))))`, "undefined", nil},
	{`(define (domain d) (:requirements :adl) (:types t) (:action a :parameters ()
		:precondition (exists (?a - (either s t)) (and))))`, "undefined", nil},
}

func TestCheckQuantNode(t *testing.T) {
	for _, test := range quantNodeTests {
		test.run(t)
	}
}

var literalNodeTests = []checkDomainTest{
	{`(define (domain d) (:requirements :adl) (:predicates (p))
		(:action a :parameters () :precondition (p)))`, "", nil},

	{`(define (domain d) (:requirements :adl)
		(:action a :parameters () :precondition (p)))`, "undefined", nil},

	// OK untyped parameter
	{`(define (domain d) (:requirements :adl) (:constants c) (:predicates (p ?a))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:predicates (p ?a))
		(:action a :parameters (?a) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:predicates (p ?a))
		(:action a :parameters () :precondition (forall (?a) (p ?a))))`, "", nil},

	// OK with a typed parameter
	{`(define (domain d) (:requirements :adl) (:types t) (:constants c - t) (:predicates (p ?a - t))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types t) (:predicates (p ?a - t))
		(:action a :parameters (?a - t) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types t) (:predicates (p ?a - t))
		(:action a :parameters () :precondition (forall (?a - t) (p ?a))))`, "", nil},

	// Incompatible parameter types: wants t, gets object
	{`(define (domain d) (:requirements :adl) (:types t) (:constants c) (:predicates (p ?a - t))
		(:action a :parameters () :precondition (p c)))`, "incompatible", nil},
	{`(define (domain d) (:requirements :adl) (:types t) (:predicates (p ?a - t))
		(:action a :parameters (?a) :precondition (p ?a)))`, "incompatible", nil},
	{`(define (domain d) (:requirements :adl) (:types t) (:predicates (p ?a - t))
		(:action a :parameters () :precondition (forall (?a) (p ?a))))`, "incompatible", nil},

	// OK parameter types: wants object, gets t
	{`(define (domain d) (:requirements :adl) (:types t) (:constants c - t) (:predicates (p ?a - object))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types t) (:predicates (p ?a - object))
		(:action a :parameters (?a - t) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types t) (:predicates (p ?a - object))
		(:action a :parameters () :precondition (forall (?a - t) (p ?a))))`, "", nil},

	// OK parameter types: wants s, gets t - s
	{`(define (domain d) (:requirements :adl) (:types t - s s) (:constants c - t) (:predicates (p ?a - s))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types t - s s) (:predicates (p ?a - s))
		(:action a :parameters (?a - t) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types t - s s) (:predicates (p ?a - s))
		(:action a :parameters () :precondition (forall (?a - t) (p ?a))))`, "", nil},

	// OK parameter types: wants u, gets t - s - u
	{`(define (domain d) (:requirements :adl) (:types t - s s - u u) (:constants c - t) (:predicates (p ?a - u))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types t - s s - u u) (:predicates (p ?a - u))
		(:action a :parameters (?a - t) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types t - s s - u u) (:predicates (p ?a - u))
		(:action a :parameters () :precondition (forall (?a - t) (p ?a))))`, "", nil},

	// OK parameter types: wants (either s t), gets s or t
	{`(define (domain d) (:requirements :adl) (:types s t) (:constants c - t)
		(:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:predicates (p ?a - (either s t)))
		(:action a :parameters (?a - t) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (forall (?a - t) (p ?a))))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:constants c - s)
		(:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:predicates (p ?a - (either s t)))
		(:action a :parameters (?a - s) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (forall (?a - s) (p ?a))))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:constants c - (either s t))
		(:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:predicates (p ?a - (either s t)))
		(:action a :parameters (?a - (either s t)) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (forall (?a - (either s t)) (p ?a))))`, "", nil},

	// Incompatible parameter types: wants (either s t) gets u
	{`(define (domain d) (:requirements :adl) (:types s t u) (:constants c - u)
		(:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (p c)))`, "incompatible", nil},
	{`(define (domain d) (:requirements :adl) (:types s t u) (:predicates (p ?a - (either s t)))
		(:action a :parameters (?a - u) :precondition (p ?a)))`, "incompatible", nil},
	{`(define (domain d) (:requirements :adl) (:types s t u) (:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (forall (?a - u) (p ?a))))`, "incompatible", nil},

	// OK parameter types: wants (either u t) gets s - u
	{`(define (domain d) (:requirements :adl) (:types s - u t u) (:constants c - s)
		(:predicates (p ?a - (either u t)))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s - u t u) (:predicates (p ?a - (either u t)))
		(:action a :parameters (?a - s) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s - u t u) (:predicates (p ?a - (either u t)))
		(:action a :parameters () :precondition (forall (?a - s) (p ?a))))`, "", nil},

	// Incompatible parameter types: wants (either s t) gets (either t u)
	{`(define (domain d) (:requirements :adl) (:types s t u) (:constants c - (either t u))
		(:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (p c)))`, "incompatible", nil},
	{`(define (domain d) (:requirements :adl) (:types s t u) (:predicates (p ?a - (either s t)))
		(:action a :parameters (?a - (either t u)) :precondition (p ?a)))`, "incompatible", nil},
	{`(define (domain d) (:requirements :adl) (:types s t u) (:predicates (p ?a - (either s t)))
		(:action a :parameters () :precondition (forall (?a - (either t u)) (p ?a))))`, "incompatible", nil},

	// OK parameter types: wants (either u t) gets (either t s - u)
	{`(define (domain d) (:requirements :adl) (:types s - u t u) (:constants c - (either t s))
		(:predicates (p ?a - (either u t)))
		(:action a :parameters () :precondition (p c)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s - u t u) (:predicates (p ?a - (either u t)))
		(:action a :parameters (?a - (either t s)) :precondition (p ?a)))`, "", nil},
	{`(define (domain d) (:requirements :adl) (:types s - u t u) (:predicates (p ?a - (either u t)))
		(:action a :parameters () :precondition (forall (?a - (either t s)) (p ?a))))`, "", nil},

	// Incompatible parameter types: wants t gets (either s t)
	{`(define (domain d) (:requirements :adl) (:types s t) (:constants c - (either s t))
		(:predicates (p ?a - t))
		(:action a :parameters () :precondition (p c)))`, "incompatible", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:predicates (p ?a - t))
		(:action a :parameters (?a - (either s t)) :precondition (p ?a)))`, "incompatible", nil},
	{`(define (domain d) (:requirements :adl) (:types s t) (:predicates (p ?a - t))
		(:action a :parameters () :precondition (forall (?a - (either s t)) (p ?a))))`, "incompatible", nil},
}

func TestCheckLiteralNode(t *testing.T) {
	for _, test := range literalNodeTests {
		test.run(t)
	}
}

type checkDomainTest struct {
	pddl        string
	errorRegexp string
	test        func(string, *Domain, *testing.T)
}

func (c checkDomainTest) run(t *testing.T) {
	dom, err := Parse("", strings.NewReader(c.pddl))
	if err != nil {
		t.Fatalf("%s\nparse error: %s", c.pddl, err)
	}
	switch err := Check(dom.(*Domain), nil); {
	case err == nil && c.errorRegexp == "":
		if c.test != nil {
			c.test(c.pddl, dom.(*Domain), t)
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
