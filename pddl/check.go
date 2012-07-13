package pddl

import (
	"fmt"
)

const (
	objectTypeName = "object"
	totalCostName  = "total-cost"
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
	// defs aggregates all of the different definition classes.
	defs struct {
		reqs   map[string]bool
		types  map[string]*Type
		consts map[string]*TypedEntry
		preds  map[string]*Predicate
		funcs  map[string]*Function
		vars   *varDefs
	}

	// varDefs is a linked list of variable definitions.
	varDefs struct {
		up         *varDefs
		name       string
		definition *TypedEntry
	}
)

// Check returns an error if there are any
// semantic errors in a domain or problem,
// otherwise all definitions are numbered and
// indentifiers are linked to their definition.
//
// If the problem is nil then only the domain
// is checked.  The domain must not be nil.
func Check(d *Domain, p *Problem) (err error) {
	defs, err := checkDomain(d)
	if err != nil || p == nil {
		return
	}
	if p.Domain.Str != d.Str {
		return fmt.Errorf("problem %s expects domain %s, but got %s",
			p.Name, p.Domain, d.Name)
	}
	if err = checkReqsDef(defs, p.Requirements); err != nil {
		return
	}
	if err := checkConstsDef(defs, p.Objects); err != nil {
		return err
	}
	for i := range p.Init {
		if err := p.Init[i].check(defs); err != nil {
			return err
		}
	}
	if err := p.Goal.check(defs); err != nil {
		return err
	}
	// check the metric
	return
}

func checkDomain(d *Domain) (defs, error) {
	defs := defs{
		reqs:   make(map[string]bool),
		types:  make(map[string]*Type),
		consts: make(map[string]*TypedEntry),
		preds:  make(map[string]*Predicate),
		funcs:  make(map[string]*Function),
	}
	if err := checkReqsDef(defs, d.Requirements); err != nil {
		return defs, err
	}
	if err := checkTypesDef(defs, d); err != nil {
		return defs, err
	}
	if err := checkConstsDef(defs, d.Constants); err != nil {
		return defs, err
	}
	if err := checkPredsDef(defs, d); err != nil {
		return defs, err
	}
	if err := checkFuncsDef(defs, d.Functions); err != nil {
		return defs, err
	}
	for _, act := range d.Actions {
		if err := checkTypedEntries(defs, act.Parameters); err != nil {
			return defs, err
		}
		for i := range act.Parameters {
			defs.vars = defs.vars.push(&act.Parameters[i])
		}
		if act.Precondition != nil {
			if err := act.Precondition.check(defs); err != nil {
				return defs, err
			}
		}
		if act.Effect != nil {
			if err := act.Effect.check(defs); err != nil {
				return defs, err
			}
		}
		for _ = range act.Parameters {
			defs.vars.pop()
		}
	}
	return defs, nil
}

// push returns a new varDefs with the given
// definitions defined.
func (v *varDefs) push(d *TypedEntry) *varDefs {
	return &varDefs{
		up:         v,
		name:       d.Str,
		definition: d,
	}
}

func checkReqsDef(defs defs, rs []Name) error {
	for _, r := range rs {
		if !supportedReqs[r.Str] {
			return makeError(r, "requirement %s is not supported", r)
		}
		if defs.reqs[r.Str] {
			return makeError(r, "%s is defined multiple times", r)
		}
		defs.reqs[r.Str] = true
	}
	if defs.reqs[":adl"] {
		defs.reqs[":strips"] = true
		defs.reqs[":typing"] = true
		defs.reqs[":negative-preconditions"] = true
		defs.reqs[":disjunctive-preconditions"] = true
		defs.reqs[":equality"] = true
		defs.reqs[":quantified-preconditions"] = true
		defs.reqs[":conditional-effects"] = true
	}
	if defs.reqs[":quantified-preconditions"] {
		defs.reqs[":existential-preconditions"] = true
		defs.reqs[":universal-preconditions"] = true
	}
	return nil
}

// checkTypesDef checks a list of type definitions
// and maps type names to their definitions, and
// builds the list of all super types of each type. 
// If the implicit object type was not defined then
// it is added.
func checkTypesDef(defs defs, d *Domain) error {
	if len(d.Types) > 0 && !defs.reqs[":typing"] {
		return makeError(d.Types[0], ":types requires :typing")
	}

	// Ensure that object is defined
	if !objectDefined(d.Types) {
		d.Types = append(d.Types, Type{
			TypedEntry: TypedEntry{
				Name: Name{Str: objectTypeName},
				Num:  len(d.Types),
			},
		})
	}

	// Map type names to their definitions
	for i, t := range d.Types {
		if len(t.Types) > 1 {
			return makeError(t, "either super types are not semantically defined")
		}
		if defs.types[t.Str] != nil {
			return makeError(t, "%s is defined multiple times", t)
		}
		defs.types[t.Str] = &d.Types[i]
		d.Types[i].Num = i
	}

	// Link parent types to their definitions
	for i := range d.Types {
		if err := checkTypeNames(defs, d.Types[i].Types); err != nil {
			return err
		}
	}

	// Build super type lists
	for i := range d.Types {
		d.Types[i].Supers = superTypes(defs, &d.Types[i])
		if len(d.Types[i].Supers) <= 0 {
			panic("no supers!")
		}
	}
	return nil
}

// objectDefined returns true if the object type
// is in the list of defined types.
func objectDefined(ts []Type) bool {
	for _, t := range ts {
		if t.Str == objectTypeName {
			return true
		}
	}
	return false
}

// superTypes returns a slice of the parent types
// of the given type, including the type itself.
func superTypes(defs defs, t *Type) (supers []*Type) {
	seen := make([]bool, len(defs.types))
	stk := []*Type{t}
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
	if obj := defs.types[objectTypeName]; !seen[obj.Num] {
		supers = append(supers, obj)
	}
	return
}

// checkConstsDef checks a list of constant or
// object definitions and maps names to their
// definitions.
func checkConstsDef(defs defs, objs []TypedEntry) error {
	for i, obj := range objs {
		if defs.consts[obj.Str] != nil {
			return makeError(obj, "%s is defined multiple times", obj)
		}
		objs[i].Num = len(defs.consts)
		defs.consts[obj.Str] = &objs[i]
	}
	if err := checkTypedEntries(defs, objs); err != nil {
		return err
	}

	// Add the object to the list of objects for its type
	for i := range objs {
		obj := &objs[i]
		for _, t := range obj.Types {
			for _, s := range t.Definition.Supers {
				s.addObject(obj)
			}
		}
	}
	return nil
}

// addObject adds an object to the list of all
// objects of the given type.  If the object has
// already been added then it is not added
// again.
func (t *Type) addObject(obj *TypedEntry) {
	for _, o := range t.Objects {
		if o == obj {
			return
		}
	}
	t.Objects = append(t.Objects, obj)
}

// checkPredsDef checks a list of predicate definitions
// and maps predicate names to their definitions.
// If :equality is required and the implicit = predicate
// was not defined then it is added.
func checkPredsDef(defs defs, d *Domain) error {
	if defs.reqs[":equality"] && !equalDefined(d.Predicates) {
		d.Predicates = append(d.Predicates, Predicate{
			Name: Name{Str: "="},
			Num:  len(defs.preds),
			Parameters: []TypedEntry{
				{Name: Name{Str: "?x"}},
				{Name: Name{Str: "?y"}},
			},
		})
	}
	for i, p := range d.Predicates {
		if defs.preds[p.Str] != nil {
			return makeError(p, "%s is defined multiple times", p)
		}
		if err := checkTypedEntries(defs, p.Parameters); err != nil {
			return err
		}
		defs.preds[p.Str] = &d.Predicates[i]
		d.Predicates[i].Num = i
	}
	return nil
}

// equalDefined returns true if the = predicate
// is in the list of defined predicates.
func equalDefined(ps []Predicate) bool {
	for _, p := range ps {
		if p.Str == "=" {
			return true
		}
	}
	return false
}

// checkFuncsDef checks a list of function definitions,
// and maps function names to their definitions.
func checkFuncsDef(defs defs, fs []Function) error {
	if len(fs) > 0 && !defs.reqs[":action-costs"] {
		return makeError(fs[0], ":functions requires :action-costs")
	}
	for i, f := range fs {
		if defs.funcs[f.Str] != nil {
			return makeError(f, "%s is defined multiple times", f)
		}
		if err := checkTypedEntries(defs, f.Parameters); err != nil {
			return err
		}
		defs.funcs[f.Str] = &fs[i]
		fs[i].Num = i
	}
	return nil
}

// checkTypedEntries ensures that the types
// of a list of typed indentifiers are valid.  If they
// are valid then they are linked to their type
// definitions.  All identifiers that have no declared
// type are linked to the object type.
func checkTypedEntries(defs defs, lst []TypedEntry) error {
	for i := range lst {
		if err := checkTypeNames(defs, lst[i].Types); err != nil {
			return err
		}
		if len(lst[i].Types) == 0 {
			lst[i].Types = []TypeName{{
				Name:       Name{Str: objectTypeName},
				Definition: defs.types[objectTypeName],
			}}
		}
	}
	return nil
}

// checkTypeNames checks that all of the type
// names are defined.  Each defined type name
// is linked to its type definition.
func checkTypeNames(defs defs, ts []TypeName) error {
	if len(ts) > 0 && !defs.reqs[":typing"] {
		return badReq(ts[0], "types", ":typing")
	}
	for j, t := range ts {
		switch def := defs.types[t.Str]; def {
		case nil:
			return makeError(t, "undefined type: %s", t)
		default:
			ts[j].Definition = def
		}
	}
	return nil
}

func (u *UnaryNode) check(defs defs) error {
	return u.Formula.check(defs)
}

func (b *BinaryNode) check(defs defs) error {
	if err := b.Left.check(defs); err != nil {
		return err
	}
	return b.Right.check(defs)
}

func (m *MultiNode) check(defs defs) error {
	for i := range m.Formula {
		if err := m.Formula[i].check(defs); err != nil {
			return err
		}
	}
	return nil
}

func (q *QuantNode) check(defs defs) error {
	if err := checkTypedEntries(defs, q.Variables); err != nil {
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

// pop returns a varDefs with the latest definition
// removed.
func (v *varDefs) pop() *varDefs {
	return v.up
}

func (n *OrNode) check(defs defs) error {
	if !defs.reqs[":disjunctive-preconditions"] {
		return badReq(n, "or", ":disjunctive-preconditions")
	}
	return n.MultiNode.check(defs)
}

func (n *NotNode) check(defs defs) error {
	switch _, ok := n.Formula.(*LiteralNode); {
	case ok && !defs.reqs[":negative-preconditions"]:
		return badReq(n, "negative literal", ":negative-preconditions")
	case !ok && !defs.reqs[":disjunctive-preconditions"]:
		return badReq(n, "not", ":disjunctive-preconditions")
	}
	return n.UnaryNode.check(defs)
}

func (i *ImplyNode) check(defs defs) error {
	if !defs.reqs[":disjunctive-preconditions"] {
		return badReq(i, "imply", ":disjunctive-preconditions")
	}
	return i.BinaryNode.check(defs)
}

func (f *ForallNode) check(defs defs) error {
	switch {
	case !f.IsEffect && !defs.reqs[":universal-preconditions"]:
		return badReq(f, "forall", ":universal-preconditions")
	case f.IsEffect && !defs.reqs[":conditional-effects"]:
		return badReq(f, "forall", ":conditional-effects")
	}
	return f.QuantNode.check(defs)
}

func (e *ExistsNode) check(defs defs) error {
	if !defs.reqs[":existential-preconditions"] {
		return badReq(e, "exists", ":existential-preconditions")
	}
	return e.QuantNode.check(defs)
}

func (w *WhenNode) check(defs defs) error {
	if !defs.reqs[":conditional-effects"] {
		return badReq(w, "when", ":conditional-effects")
	}
	if err := w.Condition.check(defs); err != nil {
		return err
	}
	return w.UnaryNode.check(defs)
}

func (lit *LiteralNode) check(defs defs) error {
	if lit.Definition = defs.preds[lit.Predicate.Str]; lit.Definition == nil {
		return makeError(lit, "undefined predicate: %s", lit.Predicate)
	}
	if lit.IsEffect {
		if lit.Negative {
			lit.Definition.NegEffect = true
		} else {
			lit.Definition.PosEffect = true
		}
	}
	return checkInst(defs, lit.Predicate, lit.Arguments, lit.Definition.Parameters)
}

// checkInst checks the arguments match the parameters
// of a predicate or function instantiation.
func checkInst(defs defs, n Name, args []Term, parms []TypedEntry) error {
	if len(args) != len(parms) {
		var argStr = "arguments"
		if len(parms) == 1 {
			argStr = argStr[:len(argStr)-1]
		}
		return makeError(n, "%s requires %d %s", n, len(parms), argStr)
	}

	for i := range args {
		kind := "constant"
		args[i].Definition = defs.consts[args[i].Str]
		if args[i].Variable {
			args[i].Definition = defs.vars.find(args[i].Str)
			kind = "variable"
		}
		if args[i].Definition == nil {
			return makeError(args[i], "undefined %s: %s", kind, args[i])
		}
		if !compatTypes(parms[i].Types, args[i].Definition.Types) {
			return makeError(args[i],
				"%s [type %s] is incompatible with parameter %s [type %s] of %s",
				args[i], typeString(args[i].Definition.Types),
				parms[i], typeString(parms[i].Types), n)
		}
	}
	return nil
}

// find returns the definition of the variable
// or nil if it was undefined.
func (v *varDefs) find(n string) *TypedEntry {
	if v == nil {
		return nil
	}
	if v.name == n {
		return v.definition
	}
	return v.up.find(n)
}

// compatTypes returns true if each type on the right
// is convertable to each type on the left.
func compatTypes(left, right []TypeName) bool {
	for _, r := range right {
		ok := false
		for _, l := range left {
			for _, s := range r.Definition.Supers {
				if s == l.Definition {
					ok = true
					break
				}
			}
			if ok {
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func (a *AssignNode) check(defs defs) error {
	if !defs.reqs[":action-costs"] {
		return badReq(a, a.Op.Str, ":action-costs")
	}
	if err := a.Lval.check(defs); err != nil {
		return err
	}
	if a.Lval.Definition.Str != totalCostName || len(a.Lval.Definition.Parameters) > 0 {
		return makeError(a.Lval, ":action-costs only allows the 0-ary total-cost function as the target of an assignments")
	}
	if a.IsNumber {
		if negative(a.Number) {
			return makeError(a, ":action-costs disallows negative numbers as the right-hand-side of an assignment")
		}
	} else {
		if a.Fhead.Str == "total-cost" {
			return makeError(a.Fhead, ":action-costs does not allow total-cost as the right-hand-side of an assignment")
		}
		if err := a.Fhead.check(defs); err != nil {
			return err
		}
	}
	return nil
}

// negative returns true if the string is a negative number.
func negative(n string) bool {
	neg := false
	for _, s := range n {
		if s != '-' {
			break
		}
		neg = !neg
	}
	return neg
}

func (h *Fhead) check(defs defs) error {
	if h.Definition = defs.funcs[h.Str]; h.Definition == nil {
		return makeError(h, "undefined function: %s", h)
	}
	return checkInst(defs, h.Name, h.Arguments, h.Definition.Parameters)
}

// badReq returns a requirement error for the case
// when something was used but its requirement
// was not defined.
func badReq(l Locer, used, reqd string) error {
	return makeError(l, "%s used but %s is not required", used, reqd)
}
