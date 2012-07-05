package pddl

import (
	"fmt"
)

var (
	// supportedReqs is a list of the requirement
	// flags that are supported by planit.
	supportedReqs = map[string]bool{
		":strips":                    true,
		":typing":                    true,
		":negative-preconditions":    true,
		":disjunctive-preconditions": true,
		":equality":                  true,
		":quantified-preconditions":  true,
		":conditional-effects":       true,

		// http://ipc.informatik.uni-freiburg.de/PddlActionCosts
		":action-costs": true,
	}
)

// declarations are names that have been declared.
type declarations struct {
	reqs, types, consts, preds, funcs map[string]bool
}

// CheckDomain returns an error if there are
// any semantic errors in the domain.
func CheckDomain(d *Domain) error {
	decl, err := d.declarations()
	if err != nil {
		return err
	}
	if len(d.Types) > 0 && !decl.reqs[":typing"] {
		return errorf(d.Types[0].Loc, ":types requires :typing")
	}
	if err := checkTypedNames(&decl, d.Types); err != nil {
		return err
	}
	for _, t := range d.Types {
		if len(t.Types) > 1 {
			return errorf(t.Loc, "'either' supertypes are not semantically defined")
		}
	}
	if err := checkTypedNames(&decl, d.Constants); err != nil {
		return err
	}
	for _, pred := range d.Predicates {
		if err := checkTypedNames(&decl, pred.Parameters); err != nil {
			return err
		}
	}
	if len(d.Functions) > 0 && !decl.reqs[":action-costs"] {
		return errorf(d.Functions[0].Loc, ":functions requires :action-costs")
	}
	for _, fun := range d.Functions {
		if fun.Str != "total-cost" || len(fun.Parameters) > 0 {
			return errorf(fun.Loc, ":action-costs only allows a 0-ary total-cost function")
		}
		if err := checkTypedNames(&decl, fun.Parameters); err != nil {
			return err
		}
	}
	for _, act := range d.Actions {
		if err := checkTypedNames(&decl, act.Parameters); err != nil {
			return err
		}
	}
	return nil
}

// checkTypedNames returns an error if there is
// type are used when :typing is not required, or
// if there is an undeclared type in the list.
func checkTypedNames(d *declarations, lst []TypedName) error {
	for _, ent := range lst {
		if len(ent.Types) > 0 && !d.reqs[":typing"] {
			return errorf(ent.Loc, "typse used but :typing is not required")
		}
		for _, typ := range ent.Types {
			if !d.types[typ.Str] {
				return errorf(typ.Loc, "type %s is not declared", typ.Str)
			}
		}
	}
	return nil
}

// declarations returns the set of declarations for a
// domain.  If there is an error in the declarations
// for the domain then it is returned.
func (d *Domain) declarations() (declarations, error) {
	decl := declarations{
		reqs:   map[string]bool{},
		types:  map[string]bool{},
		consts: map[string]bool{},
		preds:  map[string]bool{},
		funcs:  map[string]bool{},
	}
	for _, r := range d.Requirements {
		if !supportedReqs[r.Str] {
			return decl, errorf(r.Loc, "%s is not a supported requirement", r.Str)
		}
		if decl.reqs[r.Str] {
			return decl, errorf(r.Loc, "%s is declared multiple times", r.Str)
		}
		decl.reqs[r.Str] = true
	}
	for _, t := range d.Types {
		if decl.types[t.Str] {
			return decl, errorf(t.Loc, "%s is declared multiple times", t.Str)
		}
		decl.types[t.Str] = true
	}
	if decl.reqs[":typing"] {
		decl.types["object"] = true
	}
	for _, c := range d.Constants {
		if decl.consts[c.Str] {
			return decl, errorf(c.Loc, "%s is declared multiple times", c.Str)
		}
		decl.consts[c.Str] = true
	}
	for _, p := range d.Predicates {
		if decl.preds[p.Str] {
			return decl, errorf(p.Loc, "%s is declared multiple times", p.Str)
		}
		decl.preds[p.Str] = true
	}
	for _, f := range d.Functions {
		if decl.funcs[f.Str] {
			return decl, errorf(f.Loc, "%s is declared multiple times", f.Str)
		}
		decl.funcs[f.Str] = true
	}
	return decl, nil
}

// errorf returns an error with the string
// based on a formate and a location.
func errorf(loc Loc, f string, vs ...interface{}) error {
	return fmt.Errorf("%s: %s", loc.String(), fmt.Sprintf(f, vs...))
}
