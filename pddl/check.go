package pddl

import (
	"fmt"
)

var (
	// supportedReqs is a list of the requirement
	// flags that are supported by planit.
	supportedReqs = map[string]bool{
		":adl":                       true,
		":strips":                    true,
		":typing":                    true,
		":negative-preconditions":    true,
		":disjunctive-preconditions": true,
		":equality":                  true,
		":quantified-preconditions":  true,
		":universal-preconditions":   true,
		":existential-preconditions": true,
		":conditional-effects":       true,

		// http://ipc.informatik.uni-freiburg.de/PddlActionCosts
		":action-costs": true,
	}
)

type (
	// reqDefs is a requirement definition set.
	reqDefs map[string]bool

	// typeDefs maps a type name to its definition.
	typeDefs map[string]*Type

	// constDefs maps a constant or object name to
	// its definition.
	constDefs map[string]*TypedIdentifier

	// predDefs maps a predicate name to its definition.
	predDefs map[string]*Predicate

	// funcDefs maps a function name to its definition.
	funcDefs map[string]*Function

	// varDefs is a linked list of variable definitions.
	varDefs struct {
		up         *varDefs
		name       string
		definition *TypedIdentifier
	}

	// defs aggregates all of the different definition
	// classes.
	defs struct {
		reqs   reqDefs
		types  typeDefs
		consts constDefs
		preds  predDefs
		funcs  funcDefs
		vars   *varDefs
	}
)

// find returns the definition of the variable
// or nil if it was undefined.
func (v *varDefs) find(n string) *TypedIdentifier {
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
func (v *varDefs) push(d *TypedIdentifier) *varDefs {
	return &varDefs{
		up:         v,
		name:       d.Str,
		definition: d,
	}
}

// pop returns a varDefs with the latest definition
// removed.
func (v *varDefs) pop() *varDefs {
	return v.up
}

// Check returns an error if there are any
// semantic errors in a domain or problem,
// otherwise all definitions are numbered and
// indentifiers are linked to their definition.
func Check(d *Domain, p *Problem) (err error) {
	defs, err := CheckDomain(d)
	if err != nil {
		return
	}
	if p.Domain.Str != d.Str {
		return fmt.Errorf("problem %s expects domain %s, but got %s",
			p.Identifier, p.Domain, d.Identifier)
	}
	if err = checkReqsDef(defs.reqs, p.Requirements); err != nil {
		return
	}
	if err := checkConstsDef(&defs, p.Objects); err != nil {
		return err
	}
	addObjectsToTypes(defs.types, p.Objects)
	for i := range p.Init {
		if err := p.Init[i].check(&defs); err != nil {
			return err
		}
	}
	if err := p.Goal.check(&defs); err != nil {
		return err
	}
	// check the metric
	return
}

// CheckDomain returns an error if there are
// any semantic errors in the domain, otherwise
// all definitions are numbered and indentifiers
// are linked to their definition.
func CheckDomain(d *Domain) (defs defs, err error) {
	defs.reqs = reqDefs{}
	if err = checkReqsDef(defs.reqs, d.Requirements); err != nil {
		return
	}
	defs.types, err = checkTypesDef(d, defs.reqs)
	if err != nil {
		return
	}
	defs.consts = constDefs{}
	if err = checkConstsDef(&defs, d.Constants); err != nil {
		return
	}
	addObjectsToTypes(defs.types, d.Constants)
	defs.preds, err = checkPredsDef(defs.reqs, defs.types, d.Predicates)
	if err != nil {
		return
	}
	defs.funcs, err = checkFuncsDef(defs.reqs, defs.types, d.Functions)
	if err != nil {
		return
	}
	for _, act := range d.Actions {
		parms := act.Parameters
		if err = checkTypedIdentifiers(defs.reqs, defs.types, parms); err != nil {
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
	return
}

func addObjectsToTypes(types typeDefs, objs []TypedIdentifier) {
	for i := range objs {
		obj := &objs[i]
		for _, t := range obj.Types {
			for _, s := range superTypes(types, t) {
				s.addObject(obj)
			}
		}
	}
}

// addObject adds an object definition to the list
// of all objects for the receiver type if it is not
// already there.
func (t *Type)  addObject(obj *TypedIdentifier) {
	for _, o := range t.Objects {
		if o == obj {
			return
		}
	}
	t.Objects = append(t.Objects, obj)
}

// checkReqsDef checks requirement definitions.
// On success, a set of requirements is returned,
// else an error is returned.
func checkReqsDef(reqs reqDefs, rs []Identifier) error {
	for _, r := range rs {
		if !supportedReqs[r.Str] {
			return makeError(r, "%s is not a supported requirement", r)
		}
		if reqs[r.Str] {
			return makeError(r, "%s is defined multiple times", r)
		}
		reqs[r.Str] = true
	}
	if reqs[":adl"] {
		reqs[":strips"] = true
		reqs[":typing"] = true
		reqs[":negative-preconditions"] = true
		reqs[":disjunctive-preconditions"] = true
		reqs[":equality"] = true
		reqs[":quantified-preconditions"] = true
		reqs[":conditional-effects"] = true
	}
	if reqs[":quantified-preconditions"] {
		reqs[":existential-preconditions"] = true
		reqs[":universal-preconditions"] = true
	}
	return nil
}

// checkTypesDef returns a mapping from type names
// to their definition, or an error if there is a semantic
// error in the type definitions.
func checkTypesDef(d *Domain, reqs reqDefs) (typeDefs, error) {
	types := typeDefs{}
	if len(d.Types) > 0 && !reqs[":typing"] {
		return nil, makeError(d.Types[0], ":types requires :typing")
	}
	for i, t := range d.Types {
		if len(t.Types) > 1 {
			return nil, makeError(t, "either super types are not semantically defined")
		}
		if types[t.Str] != nil {
			return nil, makeError(t, "%s is defined multiple times", t)
		}
		types[t.Str] = &d.Types[i]
		d.Types[i].Num = i
	}
	if types["object"] == nil {
		obj := Type{
			TypedIdentifier: TypedIdentifier{
				Identifier: Identifier{Str:"object"},
				Num: len(d.Types),
			},
		}
		d.Types = append(d.Types, obj)
		types["object"] = &d.Types[len(d.Types)-1]
	}
	for i := range d.Types {
		if err := checkTypeNames(reqs, types, d.Types[i].Types); err != nil {
			return nil, err
		}
	}
	return types, nil
}

// checkConstsDef adds to the map from constant
// names to their definition.
func checkConstsDef(defs *defs, cs []TypedIdentifier) error {
	for i, c := range cs {
		if defs.consts[c.Str] != nil {
			return makeError(c, "%s is defined multiple times", c)
		}
		cs[i].Num = len(defs.consts)
		defs.consts[c.Str] = &cs[i]
	}
	return checkTypedIdentifiers(defs.reqs, defs.types, cs)
}

// checkPredsDef returns a map from a predicate
// name to its definition, or an error if there is a
// semantic error.
func checkPredsDef(reqs reqDefs, types typeDefs, ps []Predicate) (predDefs, error) {
	preds := predDefs{}
	for i, p := range ps {
		if preds[p.Str] != nil {
			return preds, makeError(p, "%s is defined multiple times", p)
		}
		if err := checkTypedIdentifiers(reqs, types, p.Parameters); err != nil {
			return preds, err
		}
		preds[p.Str] = &ps[i]
		ps[i].Num = i
	}
	if reqs[":equality"] && preds["="] == nil {
		preds["="] = &Predicate{
			Identifier: Identifier{Str: "="},
			Num:  len(preds),
			Parameters: []TypedIdentifier{
				{Identifier: Identifier{Str: "a"}},
				{Identifier: Identifier{Str: "b"}},
			},
		}
	}
	return preds, nil
}

// checkFuncs returns a map from a function name
// to its definition, or an error if there is a semantic
// error.
func checkFuncsDef(reqs reqDefs, types typeDefs, fs []Function) (funcDefs, error) {
	funcs := funcDefs{}
	if len(fs) > 0 && !reqs[":action-costs"] {
		return funcs, makeError(fs[0], ":functions requires :action-costs")
	}
	for i, f := range fs {
		if funcs[f.Str] != nil {
			return funcs, makeError(f, "%s is defined multiple times", f)
		}
		if f.Str != "total-cost" || len(f.Parameters) > 0 {
			return funcs, makeError(f, ":action-costs only allows a 0-ary total-cost function")
		}
		funcs[f.Str] = &fs[i]
		fs[i].Num = i
	}
	return funcs, nil
}

// checkTypedIdentifiers returns an error if there is
// type are used when :typing is not required, or
// if there is an undeclared type in the list.
func checkTypedIdentifiers(reqs reqDefs, types typeDefs, lst []TypedIdentifier) error {
	for i := range lst {
		if err := checkTypeNames(reqs, types, lst[i].Types); err != nil {
			return err
		}
		if len(lst[i].Types) == 0 {
			lst[i].Types = []TypeName{{
				Identifier: Identifier{ "object", lst[i].Loc() },
				Definition: types["object"],
			 }}
		}
	}
	return nil
}

func checkTypeNames(reqs reqDefs, types typeDefs, ts []TypeName) error {
	if len(ts) > 0 && !reqs[":typing"] {
		return makeError(ts[0], "types used but :typing is not required")
	}
	for j, t := range ts {
		switch def := types[t.Str]; def {
		case nil:
			return makeError(t, "undefined type: %s", t)
		default:
			ts[j].Definition = def
		}
	}
	return nil
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
	if err := checkTypedIdentifiers(defs.reqs, defs.types, q.Variables); err != nil {
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

func (n *OrNode) check(defs *defs) error {
	if !defs.reqs[":disjunctive-preconditions"] {
		return makeError(n, "or used but :disjunctive-preconditions is not required")
	}
	return n.MultiNode.check(defs)
}

func (n *NotNode) check(defs *defs) error {
	switch _, ok := n.Formula.(*PropositionNode); {
	case ok && !defs.reqs[":negative-preconditions"]:
		return makeError(n, "negative literal used but :negative-preconditions is not required")
	case !ok && !defs.reqs[":disjunctive-preconditions"]:
		return makeError(n, "not used but :disjunctive-preconditions is not required")
	}
	return n.UnaryNode.check(defs)
}

func (i *ImplyNode) check(defs *defs) error {
	if !defs.reqs[":disjunctive-preconditions"] {
		return makeError(i, "imply used but :disjunctive-preconditions is not required")
	}
	return i.BinaryNode.check(defs)
}

func (f *ForallNode) check(defs *defs) error {
	switch {
	case !f.Effect && !defs.reqs[":universal-preconditions"]:
		return makeError(f, "forall used but :universal-preconditions is not required")
	case f.Effect && !defs.reqs[":conditional-effects"]:
		return makeError(f, "forall used but :conditional-effects is not required")
	}
	return f.QuantNode.check(defs)
}

func (e *ExistsNode) check(defs *defs) error {
	if !defs.reqs[":existential-preconditions"] {
		return makeError(e, "exists used but :existential-preconditions is not required")
	}
	return e.QuantNode.check(defs)
}

func (w *WhenNode) check(defs *defs) error {
	if !defs.reqs[":conditional-effects"] {
		return makeError(w, "when used but :conditional-effects is not required")
	}
	if err := w.Condition.check(defs); err != nil {
		return err
	}
	return w.UnaryNode.check(defs)
}

func (p *PropositionNode) check(defs *defs) error {
	switch pred := defs.preds[p.Predicate.Str]; {
	case pred == nil:
		return makeError(p, "undefined predicate: %s", p.Predicate)
	default:
		p.Definition = pred
	}
	if len(p.Arguments) != len(p.Definition.Parameters) {
		var arg = "arguments"
		if len(p.Definition.Parameters) == 1 {
			arg = arg[:len(arg)-1]
		}
		return makeError(p, "predicate %s requires %d %s",
			p.Definition, len(p.Definition.Parameters), arg)
	}
	for i := range p.Arguments {
		kind := "constant"
		arg := defs.consts[p.Arguments[i].Str]
		if p.Arguments[i].Variable {
			arg = defs.vars.find(p.Arguments[i].Str)
			kind = "variable"
		}
		if arg == nil {
			return makeError(p.Arguments[i], "undefined %s: %s",
				kind, p.Arguments[i])
		}
		p.Arguments[i].Definition = arg

		parm := p.Definition.Parameters[i]
		if !compatTypes(defs.types, parm.Types, arg.Types) {
			return makeError(p.Arguments[i],
				"%s [type %s] is incompatible with parameter %s [type %s] of predicate %s",
				arg, typeString(arg.Types), parm, typeString(parm.Types), p.Definition)
		}
	}
	return nil
}

func (a *AssignNode) check(defs *defs) error {
	switch {
	case !defs.reqs[":action-costs"]:
		return makeError(a, "%s used but :action-costs is not required", a.Op)
	case defs.funcs[a.Lval.Str] == nil:
		return makeError(a.Lval, "undefined function: %s", a.Lval)
	}
	return nil
}

// compatTypes returns true r is convertable to l via widening.
func compatTypes(types typeDefs, left, right []TypeName) bool {
	if len(right) == 1 {
		for _, l := range left {
			for _, s := range superTypes(types, l) {
				if s == right[0].Definition {
					return true
				}
			}
		}
		return false
	}
	for _, r := range right {
		if !compatTypes(types, left, []TypeName{r}) {
			return false
		}
	}
	return true
}

// superTypes returns a slice of the parent types
// of the given type, including the type itself.
func superTypes(types typeDefs, t TypeName) (supers []*Type) {
	seen := make([]bool, len(types))
	stk := []*Type{ t.Definition }
	for len(stk) > 0 {
		t := stk[len(stk)-1]
		stk = stk[:len(stk)-1]
		if seen[t.Num] {
			continue
		}
		seen[t.Num] = true
		supers = append(supers, t)
		for _, s := range t.Types {
			stk = append(stk, s.Definition)
		}
	}
	if obj := types["object"]; !seen[obj.Num] {
		supers = append(supers, obj)
	}
	return
}
