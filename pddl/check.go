package pddl

import (
	"fmt"
	"strings"
)

const (
	// objectTypeName is the name of the default
	// object type.
	objectTypeName = "object"

	// totalCostName is the name of the total-cost
	// function.
	totalCostName  = "total-cost"
)

// Check returns the first semantic error that
// is encountered, if there are any, otherwise
// it returns no error.
//
// If the problem is nil then only the domain
// is checked.  The domain must not be nil.
func Check(d *Domain, p *Problem) (errs Errors) {
	defs := checkDomain(d, &errs)
	if p == nil {
		return
	}
	if p.Domain.Str != d.Str {
		errs.cat(Errors([]error{fmt.Errorf("problem %s expects domain %s, but got %s",
			p.Name, p.Domain, d.Name)}))
	}
	checkReqsDef(defs, p.Requirements, &errs)
	checkConstsDef(defs, p.Objects, &errs)
	for i := range p.Init {
		p.Init[i].check(defs, &errs)
	}
	p.Goal.check(defs, &errs)
	// check the metric
	return
}

type (
	// defs contains a map from names to their
	// corresponding definitions.
	defs struct {
		reqs   map[string]bool
		types  map[string]*Type
		consts map[string]*TypedEntry
		preds  map[string]*Predicate
		funcs  map[string]*Function
		vars   *varDefs
	}

	// varDefs implements a stack of variable
	// definitions.
	varDefs struct {
		up         *varDefs
		name       string
		definition *TypedEntry
	}
)

func checkDomain(d *Domain, errs *Errors) defs {
	defs := defs{
		reqs:   make(map[string]bool),
		types:  make(map[string]*Type),
		consts: make(map[string]*TypedEntry),
		preds:  make(map[string]*Predicate),
		funcs:  make(map[string]*Function),
	}
	checkReqsDef(defs, d.Requirements, errs)
	checkTypesDef(defs, d, errs)
	checkConstsDef(defs, d.Constants, errs)
	checkPredsDef(defs, d, errs)
	checkFuncsDef(defs, d.Functions, errs)
	for i := range d.Actions {
		checkActionDef(defs, &d.Actions[i], errs)
	}
	return defs
}

var (
	// supportedReqs maps the supported
	// requirement names to true, and
	// everything else to false.
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
		":action-costs": true,
	}
)

func checkReqsDef(defs defs, rs []Name, errs *Errors) {
	for _, r := range rs {
		req := strings.ToLower(r.Str)
		if !supportedReqs[req] {
			errs.add(r, "requirement %s is not supported", r)
		}
		if defs.reqs[req] {
			errs.multiDef(r, "requirement")
		}
		defs.reqs[req] = true
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
func checkTypesDef(defs defs, d *Domain, errs *Errors) {
	if len(d.Types) > 0 && !defs.reqs[":typing"] {
		errs.badReq(d.Types[0], ":types",  ":typing")
	}
	// Ensure that object is defined
	if !objectDefined(d.Types) {
		d.Types = append(d.Types, Type{
			TypedEntry: TypedEntry{
				Name: Name{Str: objectTypeName},
			},
		})
	}

	// Map type names to their definitions
	for i, t := range d.Types {
		if len(t.Types) > 1 {
			errs.add(t, "either super types are not semantically defined")
			continue
		}
		if defs.types[t.Str] != nil {
			errs.multiDef(t.Name, "type")
			continue
		}
		d.Types[i].Num = len(defs.types)
		defs.types[t.Str] = &d.Types[i]
	}

	// Link parent types to their definitions
	for i := range d.Types {
		checkTypeNames(defs, d.Types[i].Types, errs)
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
			if s.Definition != nil {
				stk = append(stk, s.Definition)
			}
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
func checkConstsDef(defs defs, objs []TypedEntry, errs *Errors) {
	for i, obj := range objs {
		if defs.consts[obj.Str] != nil {
			errs.multiDef(obj.Name, "object")
			continue
		}
		objs[i].Num = len(defs.consts)
		defs.consts[obj.Str] = &objs[i]
	}
	checkTypedEntries(defs, objs, errs)

	// Add the object to the list of objects for its type
	for i := range objs {
		obj := &objs[i]
		for _, t := range obj.Types {
			if t.Definition == nil {
				continue
			}
			for _, s := range t.Definition.Supers {
				s.addToDomain(obj)
			}
		}
	}
}

// addToDomain adds an object to the list of all
// objects of the given type.  If the object has
// already been added then it is not added
// again.
func (t *Type) addToDomain(obj *TypedEntry) {
	for _, o := range t.Domain {
		if o == obj {
			return
		}
	}
	t.Domain = append(t.Domain, obj)
}

// checkPredsDef checks a list of predicate definitions
// and maps predicate names to their definitions.
// If :equality is required and the implicit = predicate
// was not defined then it is added.
func checkPredsDef(defs defs, d *Domain, errs *Errors) {
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
			errs.multiDef(p.Name, "predicate")
			continue
		}
		checkTypedEntries(defs, p.Parameters, errs)
		counts := make(map[string]int, len(p.Parameters))
		for _, parm := range p.Parameters {
			if counts[parm.Str] > 0 {
				errs.multiDef(parm.Name, "parameter")
			}
			counts[parm.Str]++
		}
		d.Predicates[i].Num = len(defs.preds)
		defs.preds[p.Str] = &d.Predicates[i]
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
func checkFuncsDef(defs defs, fs []Function, errs *Errors) {
	if len(fs) > 0 && !defs.reqs[":action-costs"] {
		errs.badReq(fs[0], ":functions", ":action-costs")
	}
	for i, f := range fs {
		if defs.funcs[f.Str] != nil {
			errs.multiDef(f.Name, "function")
			continue
		}
		checkTypedEntries(defs, f.Parameters, errs)
		counts := make(map[string]int, len(f.Parameters))
		for _, parm := range f.Parameters {
			if counts[parm.Str] > 0 {
				errs.multiDef(parm.Name, "parameter")
			}
			counts[parm.Str]++
		}
		fs[i].Num = len(defs.funcs)
		defs.funcs[f.Str] = &fs[i]
	}
}

func checkActionDef(defs defs, act *Action, errs *Errors) {
	checkTypedEntries(defs, act.Parameters, errs)
	counts := make(map[string]int, len(act.Parameters))
	for i, parm := range act.Parameters {
		if counts[parm.Str] > 0 {
			errs.multiDef(parm.Name, "parameter")
		}
		counts[parm.Str]++
		defs.vars = defs.vars.push(&act.Parameters[i])
	}
	if act.Precondition != nil {
		act.Precondition.check(defs, errs)
	}
	if act.Effect != nil {
		act.Effect.check(defs, errs)
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
func checkTypedEntries(defs defs, lst []TypedEntry, errs *Errors) {
	for i := range lst {
		checkTypeNames(defs, lst[i].Types, errs)
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
func checkTypeNames(defs defs, ts []TypeName, errs *Errors) {
	if len(ts) > 0 && !defs.reqs[":typing"] {
		errs.badReq(ts[0], "types", ":typing")
	}
	for j, t := range ts {
		switch def := defs.types[t.Str]; def {
		case nil:
			errs.undefined(t.Name, "type")
		default:
			ts[j].Definition = def
		}
	}
}

func (u *UnaryNode) check(defs defs, errs *Errors) {
	u.Formula.check(defs, errs)
}

func (b *BinaryNode) check(defs defs, errs *Errors) {
	b.Left.check(defs, errs)
	b.Right.check(defs, errs)
}

func (m *MultiNode) check(defs defs, errs *Errors) {
	for i := range m.Formula {
		m.Formula[i].check(defs, errs)
	}
}

func (q *QuantNode) check(defs defs, errs *Errors) {
	checkTypedEntries(defs, q.Variables, errs)
	counts := make(map[string]int, len(q.Variables))
	for i, v := range q.Variables {
		if counts[v.Str] > 0 {
			errs.multiDef(v.Name, "variable")
		}
		counts[v.Str]++
		defs.vars = defs.vars.push(&q.Variables[i])
	}
	q.UnaryNode.check(defs, errs)
	for _ = range q.Variables {
		defs.vars = defs.vars.pop()
	}
}

// pop returns a varDefs with the latest definition
// removed.
func (v *varDefs) pop() *varDefs {
	return v.up
}

func (n *OrNode) check(defs defs, errs *Errors) {
	if !defs.reqs[":disjunctive-preconditions"] {
		errs.badReq(n, "or", ":disjunctive-preconditions")
	}
	n.MultiNode.check(defs, errs)
}

func (n *NotNode) check(defs defs, errs *Errors) {
	switch _, ok := n.Formula.(*LiteralNode); {
	case ok && !defs.reqs[":negative-preconditions"]:
		errs.badReq(n, "negative literal", ":negative-preconditions")
	case !ok && !defs.reqs[":disjunctive-preconditions"]:
		errs.badReq(n, "not", ":disjunctive-preconditions")
	}
	n.UnaryNode.check(defs, errs)
}

func (i *ImplyNode) check(defs defs, errs *Errors) {
	if !defs.reqs[":disjunctive-preconditions"] {
		errs.badReq(i, "imply", ":disjunctive-preconditions")
	}
	i.BinaryNode.check(defs, errs)
}

func (f *ForallNode) check(defs defs, errs *Errors) {
	switch {
	case !f.IsEffect && !defs.reqs[":universal-preconditions"]:
		errs.badReq(f, "forall", ":universal-preconditions")
	case f.IsEffect && !defs.reqs[":conditional-effects"]:
		errs.badReq(f, "forall", ":conditional-effects")
	}
	f.QuantNode.check(defs, errs)
}

func (e *ExistsNode) check(defs defs, errs *Errors) {
	if !defs.reqs[":existential-preconditions"] {
		errs.badReq(e, "exists", ":existential-preconditions")
	}
	e.QuantNode.check(defs, errs)
}

func (w *WhenNode) check(defs defs, errs *Errors) {
	if !defs.reqs[":conditional-effects"] {
		errs.badReq(w, "when", ":conditional-effects")
	}
	w.Condition.check(defs, errs)
	w.UnaryNode.check(defs, errs)
}

func (lit *LiteralNode) check(defs defs, errs *Errors) {
	if lit.Definition = defs.preds[lit.Predicate.Str]; lit.Definition == nil {
		errs.undefined(lit.Predicate, "predicate")
		return
	}
	if lit.IsEffect {
		if lit.Negative {
			lit.Definition.NegEffect = true
		} else {
			lit.Definition.PosEffect = true
		}
	}
	checkInst(defs, lit.Predicate, lit.Arguments, lit.Definition.Parameters, errs)
}

// checkInst checks the arguments match the parameters
// of a predicate or function instantiation.
func checkInst(defs defs, n Name, args []Term, parms []TypedEntry, errs *Errors) {
	if len(args) != len(parms) {
		var argStr = "arguments"
		if len(parms) == 1 {
			argStr = argStr[:len(argStr)-1]
		}
		errs.add(n, "%s requires %d %s", n, len(parms), argStr)
	}

	for i := range args {
		kind := "constant"
		args[i].Definition = defs.consts[args[i].Str]
		if args[i].Variable {
			args[i].Definition = defs.vars.find(args[i].Str)
			kind = "variable"
		}
		if args[i].Definition == nil {
			errs.undefined(args[i].Name, kind)
			return
		}
		if !compatTypes(parms[i].Types, args[i].Definition.Types) {
			errs.add(args[i],
				"%s [type %s] is incompatible with parameter %s [type %s] of %s",
				args[i], typeString(args[i].Definition.Types),
				parms[i], typeString(parms[i].Types), n)
		}
	}
}

// find returns the definition of the variable
// or nil if it was not defined.
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
		if r.Definition == nil {
			// undefined, don't report a new error.
			return true
		}
		ok := false
		for _, l := range left {
			if l.Definition == nil {
				// undefined, don't report a new error.
				return true
			}
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

func (a *AssignNode) check(defs defs, errs *Errors) {
	if !defs.reqs[":action-costs"] {
		errs.badReq(a, a.Op.Str, ":action-costs")
	}
	a.Lval.check(defs, errs)
	if a.IsNumber {
		if negative(a.Number) {
			errs.add(a, "assigned value must not be negative with :action-costs")
		}
	} else {
		a.Fhead.check(defs, errs)
	}

	if !a.IsInit {
		if a.Lval.Definition != nil && !a.Lval.Definition.isTotalCost() {
			errs.add(a.Lval, "assignment target must be a 0-ary total-cost function with :action-costs")
		}
		if !a.IsNumber && a.Fhead.Definition != nil && a.Fhead.Definition.isTotalCost() {
			errs.add(a.Fhead, "assigned value must not be total-cost with :action-costs")
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

func (h *Fhead) check(defs defs, errs *Errors) {
	if h.Definition = defs.funcs[h.Str]; h.Definition == nil {
		errs.undefined(h.Name, " function")
		return
	}
	checkInst(defs, h.Name, h.Arguments, h.Definition.Parameters, errs)
}

type Errors []error

func (es *Errors) add(l Locer, f string, vs ...interface{}) {
	*es = append(*es, Error{l.Loc(), fmt.Sprintf(f, vs...)})
}

func (es *Errors) cat(errs Errors) {
	*es = append(*es, errs...)
}

func (es *Errors) badReq(l Locer, used, reqd string) {
	*es = append(*es, MissingRequirementError{
		Location: l.Loc(),
		Cause: used,
		Requirement: reqd,
	})
}

func (es *Errors) multiDef(name Name, kind string) {
	es.add(name, "%s %s is defined multiple times", kind, name)
}

func (es *Errors) undefined(name Name, kind string) {
	es.add(name, "undefined %s %s", kind, name.Str)
}

type MissingRequirementError struct {
	Location
	Cause, Requirement string
}

func (r MissingRequirementError) Error() string {
	return r.Loc().String() + " " + r.Cause + " requires " + r.Requirement
}