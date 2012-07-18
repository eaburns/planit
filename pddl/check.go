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

// Check returns the first semantic error that
// is encountered, if there are any, otherwise
// it returns no error.
//
// If the problem is nil then only the domain
// is checked.  The domain must not be nil.
func Check(d *Domain, p *Problem) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if e, ok := r.(Error); ok {
			err = e
		} else {
			panic(r)
		}
	}()
	defs := checkDomain(d)
	if p == nil {
		return
	}
	if p.Domain.Str != d.Str {
		return fmt.Errorf("problem %s expects domain %s, but got %s",
			p.Name, p.Domain, d.Name)
	}
	checkReqsDef(defs, p.Requirements)
	checkConstsDef(defs, p.Objects)
	for i := range p.Init {
		p.Init[i].check(defs)
	}
	p.Goal.check(defs)
	// check the metric
	return
}

func checkDomain(d *Domain) defs {
	defs := defs{
		reqs:   make(map[string]bool),
		types:  make(map[string]*Type),
		consts: make(map[string]*TypedEntry),
		preds:  make(map[string]*Predicate),
		funcs:  make(map[string]*Function),
	}
	checkReqsDef(defs, d.Requirements)
	checkTypesDef(defs, d)
	checkConstsDef(defs, d.Constants)
	checkPredsDef(defs, d)
	checkFuncsDef(defs, d.Functions)
	for i := range d.Actions {
		checkActionDef(defs, &d.Actions[i])
	}
	return defs
}

func checkReqsDef(defs defs, rs []Name) {
	for _, r := range rs {
		if !supportedReqs[r.Str] {
			errorf(r, "requirement %s is not supported", r)
		}
		if defs.reqs[r.Str] {
			errorf(r, "%s is defined multiple times", r)
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
}

// checkTypesDef checks a list of type definitions
// and maps type names to their definitions, and
// builds the list of all super types of each type. 
// If the implicit object type was not defined then
// it is added.
func checkTypesDef(defs defs, d *Domain) {
	if len(d.Types) > 0 && !defs.reqs[":typing"] {
		errorf(d.Types[0], ":types requires :typing")
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
			errorf(t, "either super types are not semantically defined")
		}
		if defs.types[t.Str] != nil {
			errorf(t, "%s is defined multiple times", t)
		}
		defs.types[t.Str] = &d.Types[i]
		d.Types[i].Num = i
	}

	// Link parent types to their definitions
	for i := range d.Types {
		checkTypeNames(defs, d.Types[i].Types)
	}

	// Build super type lists
	for i := range d.Types {
		d.Types[i].Supers = superTypes(defs, &d.Types[i])
		if len(d.Types[i].Supers) <= 0 {
			panic("no supers!")
		}
	}
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
func checkConstsDef(defs defs, objs []TypedEntry) {
	for i, obj := range objs {
		if defs.consts[obj.Str] != nil {
			errorf(obj, "%s is defined multiple times", obj)
		}
		objs[i].Num = len(defs.consts)
		defs.consts[obj.Str] = &objs[i]
	}
	checkTypedEntries(defs, objs)

	// Add the object to the list of objects for its type
	for i := range objs {
		obj := &objs[i]
		for _, t := range obj.Types {
			for _, s := range t.Definition.Supers {
				s.addObject(obj)
			}
		}
	}
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
func checkPredsDef(defs defs, d *Domain) {
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
			errorf(p, "%s is defined multiple times", p)
		}
		checkTypedEntries(defs, p.Parameters)
		counts := make(map[string]int, len(p.Parameters))
		for _, parm := range p.Parameters {
			if counts[parm.Str] > 0 {
				errorf(parm, "%s is defined multiple times", parm)
			}
			counts[parm.Str]++
		}
		defs.preds[p.Str] = &d.Predicates[i]
		d.Predicates[i].Num = i
	}
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
func checkFuncsDef(defs defs, fs []Function) {
	if len(fs) > 0 && !defs.reqs[":action-costs"] {
		errorf(fs[0], ":functions requires :action-costs")
	}
	for i, f := range fs {
		if defs.funcs[f.Str] != nil {
			errorf(f, "%s is defined multiple times", f)
		}
		checkTypedEntries(defs, f.Parameters)
		counts := make(map[string]int, len(f.Parameters))
		for _, parm := range f.Parameters {
			if counts[parm.Str] > 0 {
				errorf(parm, "%s is defined multiple times", parm)
			}
			counts[parm.Str]++
		}
		defs.funcs[f.Str] = &fs[i]
		fs[i].Num = i
	}
}

func checkActionDef(defs defs, act *Action) {
	checkTypedEntries(defs, act.Parameters)
	counts := make(map[string]int, len(act.Parameters))
	for i, parm := range act.Parameters {
		if counts[parm.Str] > 0 {
			errorf(parm, "%s is defined multiple times", parm)
		}
		counts[parm.Str]++
		defs.vars = defs.vars.push(&act.Parameters[i])
	}
	if act.Precondition != nil {
		act.Precondition.check(defs)
	}
	if act.Effect != nil {
		act.Effect.check(defs)
	}
	for _ = range act.Parameters {
		defs.vars.pop()
	}
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

// checkTypedEntries ensures that the types
// of a list of typed indentifiers are valid.  If they
// are valid then they are linked to their type
// definitions.  All identifiers that have no declared
// type are linked to the object type.
func checkTypedEntries(defs defs, lst []TypedEntry) {
	for i := range lst {
		checkTypeNames(defs, lst[i].Types)
		if len(lst[i].Types) == 0 {
			lst[i].Types = []TypeName{{
				Name:       Name{Str: objectTypeName},
				Definition: defs.types[objectTypeName],
			}}
		}
	}
}

// checkTypeNames checks that all of the type
// names are defined.  Each defined type name
// is linked to its type definition.
func checkTypeNames(defs defs, ts []TypeName) {
	if len(ts) > 0 && !defs.reqs[":typing"] {
		badReq(ts[0], "types", ":typing")
	}
	for j, t := range ts {
		switch def := defs.types[t.Str]; def {
		case nil:
			errorf(t, "undefined type: %s", t)
		default:
			ts[j].Definition = def
		}
	}
}

func (u *UnaryNode) check(defs defs) {
	u.Formula.check(defs)
}

func (b *BinaryNode) check(defs defs) {
	b.Left.check(defs)
	b.Right.check(defs)
}

func (m *MultiNode) check(defs defs) {
	for i := range m.Formula {
		m.Formula[i].check(defs)
	}
}

func (q *QuantNode) check(defs defs) {
	checkTypedEntries(defs, q.Variables)
	counts := make(map[string]int, len(q.Variables))
	for i, v := range q.Variables {
		if counts[v.Str] > 0 {
			errorf(v, "%s is defined multiple times", v)
		}
		counts[v.Str]++
		defs.vars = defs.vars.push(&q.Variables[i])
	}
	q.UnaryNode.check(defs)
	for _ = range q.Variables {
		defs.vars = defs.vars.pop()
	}
}

// pop returns a varDefs with the latest definition
// removed.
func (v *varDefs) pop() *varDefs {
	return v.up
}

func (n *OrNode) check(defs defs) {
	if !defs.reqs[":disjunctive-preconditions"] {
		badReq(n, "or", ":disjunctive-preconditions")
	}
	n.MultiNode.check(defs)
}

func (n *NotNode) check(defs defs) {
	switch _, ok := n.Formula.(*LiteralNode); {
	case ok && !defs.reqs[":negative-preconditions"]:
		badReq(n, "negative literal", ":negative-preconditions")
	case !ok && !defs.reqs[":disjunctive-preconditions"]:
		badReq(n, "not", ":disjunctive-preconditions")
	}
	n.UnaryNode.check(defs)
}

func (i *ImplyNode) check(defs defs) {
	if !defs.reqs[":disjunctive-preconditions"] {
		badReq(i, "imply", ":disjunctive-preconditions")
	}
	i.BinaryNode.check(defs)
}

func (f *ForallNode) check(defs defs) {
	switch {
	case !f.IsEffect && !defs.reqs[":universal-preconditions"]:
		badReq(f, "forall", ":universal-preconditions")
	case f.IsEffect && !defs.reqs[":conditional-effects"]:
		badReq(f, "forall", ":conditional-effects")
	}
	f.QuantNode.check(defs)
}

func (e *ExistsNode) check(defs defs) {
	if !defs.reqs[":existential-preconditions"] {
		badReq(e, "exists", ":existential-preconditions")
	}
	e.QuantNode.check(defs)
}

func (w *WhenNode) check(defs defs) {
	if !defs.reqs[":conditional-effects"] {
		badReq(w, "when", ":conditional-effects")
	}
	w.Condition.check(defs)
	w.UnaryNode.check(defs)
}

func (lit *LiteralNode) check(defs defs) {
	if lit.Definition = defs.preds[lit.Predicate.Str]; lit.Definition == nil {
		errorf(lit, "undefined predicate: %s", lit.Predicate)
	}
	if lit.IsEffect {
		if lit.Negative {
			lit.Definition.NegEffect = true
		} else {
			lit.Definition.PosEffect = true
		}
	}
	checkInst(defs, lit.Predicate, lit.Arguments, lit.Definition.Parameters)
}

// checkInst checks the arguments match the parameters
// of a predicate or function instantiation.
func checkInst(defs defs, n Name, args []Term, parms []TypedEntry) {
	if len(args) != len(parms) {
		var argStr = "arguments"
		if len(parms) == 1 {
			argStr = argStr[:len(argStr)-1]
		}
		errorf(n, "%s requires %d %s", n, len(parms), argStr)
	}

	for i := range args {
		kind := "constant"
		args[i].Definition = defs.consts[args[i].Str]
		if args[i].Variable {
			args[i].Definition = defs.vars.find(args[i].Str)
			kind = "variable"
		}
		if args[i].Definition == nil {
			errorf(args[i], "undefined %s: %s", kind, args[i])
		}
		if !compatTypes(parms[i].Types, args[i].Definition.Types) {
			errorf(args[i],
				"%s [type %s] is incompatible with parameter %s [type %s] of %s",
				args[i], typeString(args[i].Definition.Types),
				parms[i], typeString(parms[i].Types), n)
		}
	}
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

func (a *AssignNode) check(defs defs) {
	if !defs.reqs[":action-costs"] {
		badReq(a, a.Op.Str, ":action-costs")
	}
	a.Lval.check(defs)
	if a.IsNumber {
		if negative(a.Number) {
			errorf(a, "assigned value must not be negative with :action-costs")
		}
	} else {
		a.Fhead.check(defs)
	}

	if !a.IsInit {
		if !a.Lval.Definition.isTotalCost() {
			errorf(a.Lval, "assignment target must be a 0-ary total-cost function with :action-costs")
		}
		if !a.IsNumber && a.Fhead.Definition.isTotalCost() {
			errorf(a.Fhead, "assigned value must not be total-cost with :action-costs")
		}
	}
}

func (f *Function) isTotalCost() bool {
	return f.Str == totalCostName && len(f.Parameters) == 0
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

func (h *Fhead) check(defs defs) {
	if h.Definition = defs.funcs[h.Str]; h.Definition == nil {
		errorf(h, "undefined function: %s", h)
	}
	checkInst(defs, h.Name, h.Arguments, h.Definition.Parameters)
}

// badReq panicks a requirement error for the case
// when something was used but its requirement
// was not defined.
func badReq(l Locer, used, reqd string) {
	errorf(l, "%s used but %s is not required", used, reqd)
}
