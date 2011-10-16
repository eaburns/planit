package prob

import "github.com/willf/bitset"

func (d *Domain) expandQuants(s *symtab) (acts []Action) {
	for _, a := range d.Actions {
		a.Precondition = a.Precondition.expandQuants(s, nil)
		a.Effect = a.Effect.expandQuants(s, nil)
		acts = append(acts, a.expandParms(s, nil, a.Parameters)...)
	}
	return
}

func (a *Action) expandParms(s *symtab, f *expFrame, ps []TypedName) (acts []Action) {
	if len(ps) == 0 {
		return a.groundedParms(s, f)
	}

	pnum := len(a.Parameters) - len(ps)
	saved := a.Parameters[pnum]
	seen := bitset.New(uint(len(s.constNames)))

	for i := range saved.Types {
		tnum := saved.Types[i].Num
		for _, obj := range s.typeObjs[tnum] {
			if seen.Test(uint(obj)) {
				continue
			}
			a.Parameters[pnum].Num = obj
			a.Parameters[pnum].Str = s.constNames[obj]
			g := f.push(saved.Num, obj)
			acts = append(acts, a.expandParms(s, g, ps[1:])...)
		}
	}

	a.Parameters[pnum] = saved
	return
}

// Return a ground instance of the given action
// which has all of its parameters replaced with
// constants
func (a *Action) groundedParms(s *symtab, f *expFrame) []Action {
	prec := a.Precondition.expandQuants(s, f)
	if _, ok := prec.(FalseNode); ok {
		return make([]Action, 0)
	}
	eff := a.Effect.expandQuants(s, f)
	if _, ok := eff.(TrueNode); ok {
		return make([]Action, 0)
	}
	act := Action{
		Name:         a.Name,
		Parameters:   make([]TypedName, len(a.Parameters)),
		Precondition: prec,
		Effect:       eff,
	}
	copy(act.Parameters, a.Parameters)
	return []Action{act}
}

func (l *Literal) expandQuants(s *symtab, f *expFrame) Formula {
	parms := make([]Term, len(l.Parameters))
	copy(parms, l.Parameters)

	varFree := true

	for i := range parms {
		if parms[i].Variable {
			vl, ok := f.lookup(parms[i].Num)
			if !ok { // Must be replaced in another pass
				varFree = false
				continue
			}
			parms[i].Num = vl
			parms[i].Str = s.constNames[vl]
			parms[i].Variable = false
		}
	}

	newLit := &Literal{
		Predicate:  l.Predicate,
		Positive:   l.Positive,
		Parameters: parms,
	}

	if varFree {
		return s.lits.intern(newLit)
	} else {
		newLit.Num = -1
	}

	return newLit
}

func (n TrueNode) expandQuants(*symtab, *expFrame) Formula {
	return n
}

func (n FalseNode) expandQuants(*symtab, *expFrame) Formula {
	return n
}

func (n *AssignNode) expandQuants(*symtab, *expFrame) Formula {
	return n
}

func (e *AndNode) expandQuants(s *symtab, f *expFrame) (res Formula) {
	switch l := e.Left.expandQuants(s, f).(type) {
	case TrueNode:
		res = e.Right.expandQuants(s, f)
	case FalseNode:
		res = FalseNode{}
	default:
		res = Conjunct(l, e.Right.expandQuants(s, f))
	}
	return
}

func (e *OrNode) expandQuants(s *symtab, f *expFrame) (res Formula) {
	switch l := e.Left.expandQuants(s, f).(type) {
	case TrueNode:
		res = TrueNode{}
	case FalseNode:
		res = e.Right.expandQuants(s, f)
	default:
		res = Disjunct(l, e.Right.expandQuants(s, f))
	}
	return
}

func (e *NotNode) expandQuants(s *symtab, f *expFrame) Formula {
	return Negate(e.Formula.expandQuants(s, f))
}

func (e *ForallNode) expandQuants(s *symtab, f *expFrame) Formula {
	seen := bitset.New(uint(len(s.constNames)))
	vr := e.Variable.Num
	conj := Formula(TrueNode{})

	for i := range e.Variable.Types {
		for _, obj := range s.typeObjs[e.Variable.Types[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}

			seen.Set(uint(obj))
			frame := f.push(vr, obj)
			conj = Conjunct(conj, e.Formula.expandQuants(s, frame))
			if _, ok := conj.(FalseNode); ok {
				return FalseNode{}
			}
		}
	}
	return conj
}

func (e *ExistsNode) expandQuants(s *symtab, f *expFrame) Formula {
	seen := bitset.New(uint(len(s.constNames)))
	vr := e.Variable.Num
	disj := Formula(FalseNode{})

	for i := range e.Variable.Types {
		for _, obj := range s.typeObjs[e.Variable.Types[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}

			seen.Set(uint(obj))
			frame := f.push(vr, obj)
			disj = Disjunct(disj, e.Formula.expandQuants(s, frame))
			if _, ok := disj.(TrueNode); ok {
				return TrueNode{}
			}
		}
	}
	return disj
}

func (e *WhenNode) expandQuants(s *symtab, f *expFrame) Formula {
	return &WhenNode{
		e.Condition.expandQuants(s, f),
		UnaryNode{e.Formula.expandQuants(s, f)},
	}
}

type expFrame struct {
	vr, vl int
	up     *expFrame
}

func (f *expFrame) lookup(vr int) (int, bool) {
	if f == nil {
		return -1, false
	}
	if f.vr == vr {
		return f.vl, true
	}
	return f.up.lookup(vr)
}

func (f *expFrame) push(vr int, vl int) *expFrame {
	return &expFrame{vr: vr, vl: vl, up: f}
}