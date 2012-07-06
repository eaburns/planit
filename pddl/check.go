package pddl

import (
	"fmt"
	"log"
)

var (
	// supportedReqs is a list of the requirement
	// flags that are supported by planit.
	supportedReqs = map[string]bool{
		":adl":	true,
		":strips":                    true,
		":typing":                    true,
		":negative-preconditions":    true,
		":disjunctive-preconditions": true,
		":equality":                  true,
		":quantified-preconditions":  true,
		":universal-preconditions":  true,
		":existential-preconditions":  true,
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
		up         *varDefs
		name       string
		definition *TypedName
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
	if p.Domain != d.Name {
		return fmt.Errorf("problem %s expects domain %s, but got %s",
			p.Name, p.Domain, d.Name)
	}
	reqs, err := checkReqsDef(p.Requirements)
	if err != nil {
		return
	}
	for req := range reqs {
		if defs.reqs[req] {
			return fmt.Errorf("problem requirement %s is already a domain requirement", req)
		}
		defs.reqs[req] = true
	}
	objs, err := checkConstsDef(defs.reqs, defs.types, p.Objects)
	if err != nil {
		return
	}
	for o, def := range objs {
		if defs.consts[o] != nil {
			return errorf(def.Loc, "object %s is already a domain constant", o)
		}
		def.Num = len(defs.consts)
		defs.consts[o] = def
	}
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
	defs.reqs, err = checkReqsDef(d.Requirements)
	if err != nil {
		return
	}
	defs.types, err = checkTypesDef(defs.reqs, d.Types)
	if err != nil {
		return
	}
	defs.consts, err = checkConstsDef(defs.reqs, defs.types, d.Constants)
	if err != nil {
		return
	}
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
	return
}

// checkReqsDef checks requirement definitions.
// On success, a set of requirements is returned,
// else an error is returned.
func checkReqsDef(rs []Name) (reqDefs, error) {
	reqs := reqDefs{}
	for _, r := range rs {
		if !supportedReqs[r.Str] {
			return reqs, errorf(r.Loc, "%s is not a supported requirement", r.Str)
		}
		if reqs[r.Str] {
			return reqs, errorf(r.Loc, "%s is defined multiple times", r.Str)
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
	return reqs, nil
}

// checkTypesDef returns a mapping from type names
// to their definition, or an error if there is a semantic
// error in the type definitions.
func checkTypesDef(reqs reqDefs, ts []TypedName) (typeDefs, error) {
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

// checkConstsDef returns the map from constant
// names to their definition if there are no semantic
// errors, otherwise an error is returned.
func checkConstsDef(reqs reqDefs, types typeDefs, cs []TypedName) (constDefs, error) {
	consts := constDefs{}
	for i, c := range cs {
		if consts[c.Str] != nil {
			return consts, errorf(c.Loc, "%s is defined multiple times", c.Str)
		}
		consts[c.Str] = &cs[i]
		cs[i].Num = i
	}
	return consts, checkTypedNames(reqs, types, cs)
}

// checkPredsDef returns a map from a predicate
// name to its definition, or an error if there is a
// semantic error.
func checkPredsDef(reqs reqDefs, types typeDefs, ps []Predicate) (predDefs, error) {
	preds := predDefs{}
	for i, p := range ps {
		if preds[p.Str] != nil {
			return preds, errorf(p.Loc, "%s is defined multiple times", p.Str)
		}
		if err := checkTypedNames(reqs, types, p.Parameters); err != nil {
			return preds, err
		}
		preds[p.Str] = &ps[i]
		ps[i].Num = i
	}
	if reqs[":equality"] && preds["="] == nil {
		preds["="] = &Predicate{
			Name: Name{ Str: "=" },
			Num: len(preds),
			Parameters: []TypedName {
				{ Name: Name{Str: "a"} },
				{ Name: Name{Str: "b"} },
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
	for i, f := range fs {
		if funcs[f.Str] != nil {
			return funcs, errorf(f.Loc, "%s is defined multiple times", f.Str)
		}
		if f.Str != "total-cost" || len(f.Parameters) > 0 {
			return funcs, errorf(f.Loc, ":action-costs only allows a 0-ary total-cost function")
		}
		funcs[f.Str] = &fs[i]
		fs[i].Num = i
	}
	if len(fs) > 0 && !reqs[":action-costs"] {
		return funcs, errorf(fs[0].Loc, ":functions requires :action-costs")
	}
	return funcs, nil
}

// checkTypedNames returns an error if there is
// type are used when :typing is not required, or
// if there is an undeclared type in the list.
func checkTypedNames(reqs reqDefs, types typeDefs, lst []TypedName) error {
	for i, ent := range lst {
		if len(ent.Types) > 0 && !reqs[":typing"] {
			return errorf(ent.Loc, "types used but :typing is not required")
		}
		for j, typ := range ent.Types {
			switch def := types[typ.Str]; def {
			case nil:
				return errorf(typ.Loc, "undefined type: %s", typ.Str)
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

func (n *OrNode) check(defs *defs) error {
	if !defs.reqs[":disjunctive-preconditions"] {
		return errorf(n.Loc, "or used but :disjunctive-preconditions is not required")
	}
	return n.MultiNode.check(defs)
}

func (n *NotNode) check(defs *defs) error {
	switch _, ok := n.Formula.(*PropositionNode); {
	case ok && !defs.reqs[":negative-preconditions"]:
		return errorf(n.Loc, "negative literal used but :negative-preconditions is not required")
	case !ok && !defs.reqs[":disjunctive-preconditions"]:
		return errorf(n.Loc, "not used but :disjunctive-preconditions is not required")
	}
	return n.UnaryNode.check(defs)
}

func (i *ImplyNode) check(defs *defs) error {
	if !defs.reqs[":disjunctive-preconditions"] {
		return errorf(i.Loc, "imply used but :disjunctive-preconditions is not required")
	}
	return i.BinaryNode.check(defs)
}

func (f *ForallNode) check(defs *defs) error {
	switch {
	case !f.Effect && !defs.reqs[":universal-preconditions"]:
		return errorf(f.Loc, "forall used but :universal-preconditions is not required")
	case  f.Effect && !defs.reqs[":conditional-effects"]:
		return errorf(f.Loc, "forall used but :conditional-effects is not required")
	}
	return f.QuantNode.check(defs)
}

func (e *ExistsNode) check(defs *defs) error {
	if !defs.reqs[":existential-preconditions"] {
		return errorf(e.Loc, "exists used but :existential-preconditions is not required")
	}
	return e.QuantNode.check(defs)
}

func (w *WhenNode) check(defs *defs) error {
	if !defs.reqs[":conditional-effects"] {
		return errorf(w.Loc, "when used but :conditional-effects is not required")
	}
	if err := w.Condition.check(defs); err != nil {
		return err
	}
	return w.UnaryNode.check(defs)
}

func (p *PropositionNode) check(defs *defs) error {
	switch pred := defs.preds[p.Predicate]; {
	case pred == nil:
		return errorf(p.Loc, "undefined predicate: %s", p.Predicate)
	default:
		p.Definition = pred
	}
	if len(p.Arguments) != len(p.Definition.Parameters) {
		var arg = "arguments"
		if len(p.Definition.Parameters) == 1 {
			arg = arg[:len(arg)-1]
		}
		return errorf(p.Loc, "predicate %s requires %d %s",
			p.Definition.Str, len(p.Definition.Parameters), arg)
	}
	for i := range p.Arguments {
		kind := "constant"
		arg := defs.consts[p.Arguments[i].Str]
		if p.Arguments[i].Variable {
			arg = defs.vars.find(p.Arguments[i].Str)
			kind = "variable"
		}
		if arg == nil {
			return errorf(p.Arguments[i].Loc, "undefined %s: %s",
				kind, p.Arguments[i].Str)
		}
		p.Arguments[i].Definition = arg

		parm := p.Definition.Parameters[i]
		if !compatTypes(defs.types, parm.Types, arg.Types) {
			return errorf(p.Arguments[i].Loc,
				"%s [type %s] is incompatible with parameter %s [type %s] of predicate %s",
				arg.Str, typeString(arg.Types), parm.Str,
 				typeString(parm.Types), p.Definition.Str)
		}
	}
	return nil
}

func (a *AssignNode) check(defs *defs) error {
	switch {
	case !defs.reqs[":action-costs"]:
		return errorf(a.Loc, "%s used but :action-costs is not required", a.Op)
	case defs.funcs[a.Lval.Str] == nil:
		return errorf(a.Lval.Loc, "undefined function: %s", a.Lval.Str)
	}
	log.Printf("TODO: check assignment for numeric Rval\n")
	return nil
}

// compatTypes returns true r is convertable to l via widening.
func compatTypes(types typeDefs, left, right []Type) bool {
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
		if !compatTypes(types, left, []Type{r}) {
			return false
		}
	}
	return true
}

// superTypes returns a slice of the parent types
// of the given type, including the type itself.
func superTypes(types typeDefs, t Type) (supers []*TypedName) {
	seen := make([]bool, len(types))
	cur := t.Definition
	for !seen[cur.Num] {
		seen[cur.Num] = true
		supers = append(supers, cur)
		if len(cur.Types) > 1 {
			// This should have been caught already
			panic(cur.Str + " has more than one parent type")
		}
		if len(cur.Types) > 0 {
			cur = cur.Types[0].Definition
		}
	}
	return
}
