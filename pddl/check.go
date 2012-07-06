package pddl

import (
	"fmt"
	"log"
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

type (
	// reqDefs is a requirement definition set.
	reqDefs map[string]bool

	// typeDefs maps a type name to its definition.
	typeDefs map[string]*TypedName

	// constDefs maps a constant or object name to
	// its definition.
	constDefs map[string]*TypedName

	// predDefs maps a predicate name to its definition.
	predDefs map[string]*Predicate

	// funcDefs maps a function name to its definition.
	funcDefs map[string]*Function

	// varDefs is a linked list of variable definitions.
	varDefs struct {
		up *varDefs
		name string
		definition *TypedName
	}

	// defs aggregates all of the different definition
	// classes.
	defs struct {
		reqs reqDefs
		types typeDefs
		consts constDefs
		preds predDefs
		funcs funcDefs
		vars *varDefs
	}
)

// find returns the definition of the variable
// or nil if it was undefined.
func (v *varDefs) find(n string) *TypedName {
	if v == nil {
		return nil
	}
	if v.name == n {
		return v.definition
	}
	return v.up.find(n)
}

// push returns a new varDefs with the given
// definitions defined.
func (v *varDefs) push(d *TypedName) *varDefs {
	return &varDefs {
		up: v,
		name: d.Str,
		definition: d,
	}
}

// pop returns a varDefs with the latest definition
// removed.
func (v *varDefs) pop() *varDefs {
	return v.up
}

// CheckDomain returns an error if there are
// any semantic errors in the domain, otherwise
// all definitions are numbered and indentifiers
// in the domain are linked to their definition, via
// the appropriate pointers.
func CheckDomain(d *Domain) (err error) {
	defs := defs{}

	defs.reqs, err = checkReqs(d.Requirements)
	if err != nil {
		return
	}
	defs.types, err = checkTypes(defs.reqs, d.Types)
	if err != nil {
		return
	}
	defs.consts, err = checkConsts(defs.reqs, defs.types, d.Constants)
	if err != nil {
		return
	}
	defs.preds, err = checkPreds(defs.reqs, defs.types, d.Predicates)
	if err != nil {
		return
	}
	defs.funcs, err = checkFuncs(defs.reqs, defs.types, d.Functions)
	if err != nil {
		return
	}
	for _, act := range d.Actions {
		parms := act.Parameters
		if err = checkTypedNames(defs.reqs, defs.types, parms); err != nil {
			return
		}
		for i := range act.Parameters {
			defs.vars = defs.vars.push(&act.Parameters[i])
		}
		if act.Precondition != nil {
			err = act.Precondition.check(&defs)
		}
		if err != nil && act.Effect != nil {
			err = act.Effect.check(&defs)
		}
		for _ = range act.Parameters {
			defs.vars.pop()
		}
		if err != nil {
			return
		}
	}
	return nil
}

// checkReqs checks requirement definitions.
// On success, a set of requirements is returned,
// else an error is returned.
func checkReqs(rs []Name) (reqDefs, error) {
	reqs := reqDefs{}
	for _, r := range rs {
		if !supportedReqs[r.Str] {
			return reqs, errorf(r.Loc, "%s is not a supported requirement", r.Str)
		}
		if reqs[r.Str] {
			return reqs, errorf(r.Loc, "%s is dsared multiple times", r.Str)
		}
		reqs[r.Str] = true
	}
	return reqs, nil
}

// checkTypes returns a mapping from type names
// to their definition, or an error if there is a semantic
// error in the type definitions.
func checkTypes(reqs reqDefs, ts []TypedName) (typeDefs, error) {
	types := typeDefs{}
	if len(ts) > 0 && !reqs[":typing"] {
		return types, errorf(ts[0].Loc, ":types requires :typing")
	}
	for i, t := range ts {
		if types[t.Str] != nil {
			return types, errorf(t.Loc, "%s is defined multiple times", t.Str)
		}
		types[t.Str] = &ts[i]
		ts[i].Num = i
	}
	if reqs[":typing"] && types["object"] == nil {
		types["object"] = &TypedName{Name: Name{Str: "object"}, Num: len(types)}
	}
	if err := checkTypedNames(reqs, types, ts); err != nil {
		return types, err
	}
	for _, t := range ts {
		if len(t.Types) > 1 {
			return types, errorf(t.Loc, "'either' supertypes are not semantically defined")
		}
	}
	return types, nil
}

// checkConsts returns the map from constant names
// to their definition if there are no semantic errors,
// otherwise an error is returned.
func checkConsts(reqs reqDefs, types typeDefs, cs []TypedName) (constDefs, error) {
	consts := constDefs{}
	for i, c := range cs {
		if consts[c.Str] != nil {
			return consts, errorf(c.Loc, "%s is dsared multiple times", c.Str)
		}
		consts[c.Str] = &cs[i]
		cs[i].Num = i
	}
	if err := checkTypedNames(reqs, types, cs); err != nil {
 		return consts, err
 	}
	return consts, nil
}

// checkPreds returts a map from a predicate name
// to its definition, or an error if there is a semantic
// error.
func checkPreds(reqs reqDefs, types typeDefs, ps []Predicate) (predDefs, error) {
	preds := predDefs{}
	for i, p := range ps {
		if preds[p.Str] != nil {
			return preds, errorf(p.Loc, "%s is dsared multiple times", p.Str)
		}
		preds[p.Str] = &ps[i]
		ps[i].Num = i
	}
	for _, pred := range ps {
		if err := checkTypedNames(reqs, types, pred.Parameters); err != nil {
			return preds, err
		}
	}
	return preds, nil
}

// checkFuncs returts a map from a function name
// to its definition, or an error if there is a semantic
// error.
func checkFuncs(reqs reqDefs, types typeDefs, fs []Function) (funcDefs, error) {
	funcs := funcDefs{}
	for i, f := range fs {
		if funcs[f.Str] != nil {
			return funcs, errorf(f.Loc, "%s is dsared multiple times", f.Str)
		}
		funcs[f.Str] = &fs[i]
		fs[i].Num = i
	}
	if len(fs) > 0 && !reqs[":action-costs"] {
		return funcs, errorf(fs[0].Loc, ":functions requires :action-costs")
	}
	for _, fun := range fs {
		if fun.Str != "total-cost" || len(fun.Parameters) > 0 {
			return funcs, errorf(fun.Loc, ":action-costs only allows a 0-ary total-cost function")
		}
		if err := checkTypedNames(reqs, types, fun.Parameters); err != nil {
			return funcs ,err
		}
	}
	return funcs, nil
}

// checkTypedNames returns an error if there is
// type are used when :typing is not required, or
// if there is an undeclared type in the list.
func checkTypedNames(reqs reqDefs, types typeDefs, lst []TypedName) error {
	for i, ent := range lst {
		if len(ent.Types) > 0 && !reqs[":typing"] {
			return errorf(ent.Loc, "typse used but :typing is not required")
		}
		for j, typ := range ent.Types {
			switch def := types[typ.Str]; def {
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

func (u *UnaryNode) check(defs *defs) error {
	return u.Formula.check(defs)
}

func (b *BinaryNode) check(defs *defs) error {
	if err := b.Left.check(defs); err != nil {
		return err
	}
	return b.Right.check(defs)
}


func (m *MultiNode) check(defs *defs) error {
	for i := range m.Formula {
		if err := m.Formula[i].check(defs); err != nil {
			return err
		}
	}
	return nil
}

func (q *QuantNode) check(defs *defs) error {
	if err := checkTypedNames(defs.reqs, defs.types, q.Variables); err != nil {
		return err
	}
	for i := range q.Variables {
		defs.vars = defs.vars.push(&q.Variables[i])
	}
	err := q.Formula.check(defs)
	if err == nil {
		err = q.UnaryNode.check(defs)
	}
	for _ = range q.Variables {
		defs.vars = defs.vars.pop()
	}
	return err
}

func (w *WhenNode) check(defs *defs) error {
	if err := w.Condition.check(defs); err != nil {
		return err
	}
	return w.UnaryNode.check(defs)
}

func (p *PropositionNode) check(defs *defs) error {
	switch pred := defs.preds[p.Str]; {
	case pred == nil:
		return errorf(p.Loc, "undefined predicate: %s", p.Str)
	default:
		p.Definition = pred
	}
	for i := range p.Parameters {
		var n *TypedName
		var kind string = "variable"
		if p.Parameters[i].Variable {
			n = defs.vars.find(p.Parameters[i].Str)
		} else {
			n = defs.consts[p.Parameters[i].Str]
			kind = "constant"
		}
		if n == nil {
			return errorf(p.Parameters[i].Loc, "undefined %s: %s", kind, p.Parameters[i].Str)
		}
		p.Parameters[i].Definition = n
	}
	return nil
}

func (a *AssignNode) check(defs *defs) error {
	if defs.funcs[a.Lval.Str] == nil {
		return errorf(a.Lval.Loc, "undefined function: %s", a.Lval.Str)
	}
	log.Printf("TODO: check assignment for numeric Rval\n")
	return nil
}

