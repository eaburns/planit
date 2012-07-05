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

// defs has the set of definitons for a domain/problem.
type defs struct {
	reqs map[string]bool
	types map[string]*TypedName
	consts map[string]*TypedName
	preds map[string]*Predicate
	funcs map[string]*Function
}

// makeDefs returns an empty defs.
func makeDefs() defs {
	return defs{
		reqs:   map[string]bool{},
		types:  map[string]*TypedName{},
		consts: map[string]*TypedName{},
		preds:  map[string]*Predicate{},
		funcs:  map[string]*Function{},
	}
}

// defs returns the set of definitions for a domain.
// If there is an error in the definitions, it is returned.
func (d *Domain) defs() (defs, error) {
	ds := makeDefs()
	for _, r := range d.Requirements {
		if !supportedReqs[r.Str] {
			return ds, errorf(r.Loc, "%s is not a supported requirement", r.Str)
		}
		if ds.reqs[r.Str] {
			return ds, errorf(r.Loc, "%s is dsared multiple times", r.Str)
		}
		ds.reqs[r.Str] = true
	}
	for i, t := range d.Types {
		if ds.types[t.Str] != nil {
			return ds, errorf(t.Loc, "%s is dsared multiple times", t.Str)
		}
		ds.types[t.Str] = &d.Types[i]
		d.Types[i].Num = i
	}
	if ds.reqs[":typing"] && ds.types["object"] == nil {
		ds.types["object"] = &TypedName{ Name: Name{ Str: "object" }, Num: len(ds.types) }
	}
	for i, c := range d.Constants {
		if ds.consts[c.Str] != nil {
			return ds, errorf(c.Loc, "%s is dsared multiple times", c.Str)
		}
		ds.consts[c.Str] = &d.Constants[i]
		d.Constants[i].Num = i
	}
	for i, p := range d.Predicates {
		if ds.preds[p.Str] != nil {
			return ds, errorf(p.Loc, "%s is dsared multiple times", p.Str)
		}
		ds.preds[p.Str] = &d.Predicates[i]
		d.Predicates[i].Num = i
	}
	for i, f := range d.Functions {
		if ds.funcs[f.Str] != nil {
			return ds, errorf(f.Loc, "%s is dsared multiple times", f.Str)
		}
		ds.funcs[f.Str] = &d.Functions[i]
		d.Functions[i].Num = i
	}
	return ds, nil
}

// CheckDomain returns an error if there are
// any semantic errors in the domain, otherwise
// all definitions are numbered and indentifiers
// in the domain are linked to their definition, via
// the appropriate pointers.
func CheckDomain(d *Domain) error {
	ds, err := d.defs()
	if err != nil {
		return err
	}

	if len(d.Types) > 0 && !ds.reqs[":typing"] {
		return errorf(d.Types[0].Loc, ":types requires :typing")
	}
	if err := checkTypedNames(&ds, d.Types); err != nil {
		return err
	}
	for _, t := range d.Types {
		if len(t.Types) > 1 {
			return errorf(t.Loc, "'either' supertypes are not semantically defined")
		}
	}

	if err := checkTypedNames(&ds, d.Constants); err != nil {
		return err
	}

	for _, pred := range d.Predicates {
		if err := checkTypedNames(&ds, pred.Parameters); err != nil {
			return err
		}
	}

	if len(d.Functions) > 0 && !ds.reqs[":action-costs"] {
		return errorf(d.Functions[0].Loc, ":functions requires :action-costs")
	}
	for _, fun := range d.Functions {
		if fun.Str != "total-cost" || len(fun.Parameters) > 0 {
			return errorf(fun.Loc, ":action-costs only allows a 0-ary total-cost function")
		}
		if err := checkTypedNames(&ds, fun.Parameters); err != nil {
			return err
		}
	}

	for _, act := range d.Actions {
		if err := checkTypedNames(&ds, act.Parameters); err != nil {
			return err
		}
	}
	return nil
}

// checkTypedNames returns an error if there is
// type are used when :typing is not required, or
// if there is an undeclared type in the list.
func checkTypedNames(d *defs, lst []TypedName) error {
	for i, ent := range lst {
		if len(ent.Types) > 0 && !d.reqs[":typing"] {
			return errorf(ent.Loc, "typse used but :typing is not required")
		}
		for j, typ := range ent.Types {
			switch def := d.types[typ.Str]; def {
			case nil:
				return errorf(typ.Loc, "type %s is not declared", typ.Str)
			default:
				lst[i].Types[j].Definition = def
			}
		}
	}
	return nil
}

// errorf returns an error with the string
// based on a formate and a location.
func errorf(loc Loc, f string, vs ...interface{}) error {
	return fmt.Errorf("%s: %s", loc.String(), fmt.Sprintf(f, vs...))
}
