package prob

import "github.com/willf/bitset"

func (d *Domain) ExpandQuants(s *Symtab) {
	acts := make([]Action, 0, len(d.Actions))
	for i := range d.Actions {
		a := &d.Actions[i]
		a.Precondition = a.Precondition.expandQuants(s, nil)
		a.Effect = a.Effect.expandQuants(s, nil)
		acts = append(acts, a.expandParms(s, nil, a.Parameters)...)
	}
	d.Actions = acts
}

func (a *Action) expandParms(s *Symtab, f *expFrame, ps []TypedName) (acts []Action) {
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
func (a *Action) groundedParms(s *Symtab, f *expFrame) []Action {
	prec := a.Precondition.expandQuants(s, f)
	if _, ok := prec.(*falseNode); ok {
		return make([]Action, 0)
	}
	eff := a.Effect.expandQuants(s, f)
	if _, ok := eff.(*trueNode); ok {
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

func (l *Literal) expandQuants(s *Symtab, f *expFrame) Formula {
	parms := make([]Term, len(l.Parameters))
	copy(parms, l.Parameters)

	for i := range parms {
		if term, ok := parms[i].(Variable); ok {
			vl, ok := f.lookup(term.Num)
			if !ok {	// Must be replaced in another pass
				continue
			}
			term.Num = vl
			term.Str = s.constNames[vl]
			parms[i] = Constant{term.Name}
		}
	}

	return &Literal{
		Name:       l.Name,
		Positive:   l.Positive,
		Parameters: parms,
	}
}

func (n *LeafNode) expandQuants(*Symtab, *expFrame) Formula {
	return n
}

func (e *AndNode) expandQuants(s *Symtab, f *expFrame) (res Formula) {
	switch l := e.Left.expandQuants(s, f).(type) {
	case *trueNode:
		res = e.Right.expandQuants(s, f)
	case *falseNode:
		res = MakeFalse()
	default:
		res = Conjunct(l, e.Right.expandQuants(s, f))
	}
	return
}

func (e *OrNode) expandQuants(s *Symtab, f *expFrame) (res Formula) {
	switch l := e.Left.expandQuants(s, f).(type) {
	case *trueNode:
		res = MakeTrue()
	case *falseNode:
		res = e.Right.expandQuants(s, f)
	default:
		res = Disjunct(l, e.Right.expandQuants(s, f))
	}
	return
}

func (e *NotNode) expandQuants(s *Symtab, f *expFrame) Formula {
	return Negate(e.Formula.expandQuants(s, f))
}

func (e *ForallNode) expandQuants(s *Symtab, f *expFrame) Formula {
	seen := bitset.New(uint(len(s.constNames)))
	vr := e.Variable.Num
	conj := Formula(MakeTrue())

	for i := range e.Variable.Types {
		for _, obj := range s.typeObjs[e.Variable.Types[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}

			seen.Set(uint(obj))
			frame := f.push(vr, obj)
			conj = Conjunct(conj, e.Formula.expandQuants(s, frame))
			if _, ok := conj.(*falseNode); ok {
				return MakeFalse()
			}
		}
	}
	return conj
}

func (e *ExistsNode) expandQuants(s *Symtab, f *expFrame) Formula {
	seen := bitset.New(uint(len(s.constNames)))
	vr := e.Variable.Num
	disj := Formula(MakeFalse())

	for i := range e.Variable.Types {
		for _, obj := range s.typeObjs[e.Variable.Types[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}

			seen.Set(uint(obj))
			frame := f.push(vr, obj)
			disj = Disjunct(disj, e.Formula.expandQuants(s, frame))
			if _, ok := disj.(*trueNode); ok {
				return MakeTrue()
			}
		}
	}
	return disj
}

func (e *WhenNode) expandQuants(s *Symtab, f *expFrame) Formula {
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
