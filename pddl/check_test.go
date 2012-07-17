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

var constsDefTests = []checkDomainTest {
	{`(define (domain d) (:constants))`, "", nil},
	{`(define (domain d) (:constants a b c))`, "", nil},
	{`(define (domain d) (:constants a b c a))`, "multiple", nil},
	{`(define (domain d) (:requirements :typing) (:constants a - t))`, "undefined", nil},
	{`(define (domain d) (:constants a b))`, "",
		domainChecks(checkTypeConsts("object", []string{"a", "b"}),
			checkConstsTypes("a", []string{"object"}),
			checkConstsTypes("b", []string{"object"}))},
	{`(define (domain d) (:requirements :typing) (:constants a b - object))`, "",
		domainChecks(checkTypeConsts("object", []string{"a", "b"}),
			checkConstsTypes("a", []string{"object"}),
			checkConstsTypes("b", []string{"object"}))},
	{`(define (domain d) (:requirements :typing) (:types t) (:constants a - t))`, "",
		domainChecks(checkTypeConsts("object", []string{"a"}),
			checkTypeConsts("t", []string{"a"}),
			checkConstsTypes("a", []string{"t"}))},
	{`(define (domain d) (:requirements :typing) (:types t - s s) (:constants a - t))`, "",
		domainChecks(checkTypeConsts("object", []string{"a"}),
			checkTypeConsts("t", []string{"a"}),
			checkTypeConsts("s", []string{"a"}),
			checkConstsTypes("a", []string{"t"}))},
	{`(define (domain d) (:requirements :typing) (:types s t) (:constants a - (either s t)))`, "",
		domainChecks(checkTypeConsts("object", []string{"a"}),
			checkTypeConsts("t", []string{"a"}),
			checkTypeConsts("s", []string{"a"}),
			checkConstsTypes("a", []string{"t", "s"}))},
	{`(define (domain d) (:requirements :typing) (:types t - s s) (:constants a - (either s t)))`, "",
		domainChecks(checkTypeConsts("object", []string{"a"}),
			checkTypeConsts("t", []string{"a"}),
			checkTypeConsts("s", []string{"a"}),
			checkConstsTypes("a", []string{"t", "s"}))},
	{`(define (domain d) (:requirements :typing) (:types t) (:constants a b - t))`, "",
		domainChecks(checkTypeConsts("object", []string{"a", "b"}),
			checkTypeConsts("t", []string{"a", "b"}),
			checkConstsTypes("a", []string{"t"}),
			checkConstsTypes("b", []string{"t"}))},
	{`(define (domain d) (:requirements :typing) (:types t) (:constants a - t b))`, "",
		domainChecks(checkTypeConsts("object", []string{"a", "b"}),
			checkTypeConsts("t", []string{"a"}),
			checkConstsTypes("a", []string{"t"}),
			checkConstsTypes("b", []string{"object"}))},
}

// checkTypeConsts returns a function that
// checks that the given type has all of the
// assigned constants.
func checkTypeConsts(tName string, consts []string) func(string, *Domain, *testing.T) {
	return func(pddl string, d *Domain, t *testing.T) {
		typ := findType(tName, d.Types)
		if typ == nil {
			t.Fatalf("%s\ntype %s: not found", pddl, tName)
		}
		for _, con := range consts {
			found := false
			for _, obj := range typ.Objects {
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

var predsDefTests = []checkDomainTest {
	{`(define (domain d) (:predicates (p)))`, "", nil},
	{`(define (domain d) (:predicates (p) (q)))`, "", nil},
	{`(define (domain d) (:predicates (p ?a ?b)))`, "",
		domainChecks(checkParamTypes("p", "?a", []string{"object"}),
		checkParamTypes("p", "?b", []string{"object"}))},
	{`(define (domain d) (:predicates (p ?a ?a)))`, "multiple", nil},
	{`(define (domain d) (:requirements :typing) (:types t) (:predicates (p ?a - t)))`, "",
		checkParamTypes("p", "?a", []string{"t"})},
	{`(define (domain d) (:requirements :typing) (:types s t) (:predicates (p ?a - (either s t))))`, "",
		checkParamTypes("p", "?a", []string{"t", "s"})},
	{`(define (domain d) (:requirements :typing) (:types t) (:predicates (p ?a - t ?b)))`, "",
		domainChecks(checkParamTypes("p", "?a", []string{"t"}),
			checkParamTypes("p", "?b", []string{"object"}))},
}

// checkPredParamTypes returns a function
// that checks the types assigned to given
// predicate's parameter.
func checkParamTypes(pred string, parm string, types []string) func(string, *Domain, *testing.T) {
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
