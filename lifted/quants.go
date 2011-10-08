package lifted

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

	for i := range saved.Type {
		tnum := saved.Type[i].Num
		for _, obj := range s.typeObjs[tnum] {
			if seen.Test(uint(obj)) {
				continue
			}
			a.Parameters[pnum].Name.Num = obj
			a.Parameters[pnum].Name.Str = s.constNames[obj]
			g := f.push(saved.Name.Num, obj)
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

func (l *LiteralNode) expandQuants(s *Symtab, f *expFrame) Formula {
	parms := make([]Term, len(l.Parameters))
	copy(parms, l.Parameters)

	for i := range parms {
		if parms[i].Kind == TermConstant {
			continue
		}
		vl, ok := f.lookup(parms[i].Name.Num)
		if !ok {
			// Must be replaced in another pass
			continue
		}
		parms[i].Kind = TermConstant
		parms[i].Name.Num = vl
		parms[i].Name.Str = s.constNames[vl]
	}

	return &LiteralNode{
		Positive:   l.Positive,
		Name:       l.Name,
		Parameters: parms,
	}
}

func (e TrueNode) expandQuants(*Symtab, *expFrame) Formula {
	return e
}

func (e FalseNode) expandQuants(*Symtab, *expFrame) Formula {
	return e
}

func (e *AndNode) expandQuants(s *Symtab, f *expFrame) (res Formula) {
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

func (e *OrNode) expandQuants(s *Symtab, f *expFrame) (res Formula) {
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

func (e *NotNode) expandQuants(s *Symtab, f *expFrame) Formula {
	return Negate(e.Formula.expandQuants(s, f))
}

func (e *ForallNode) expandQuants(s *Symtab, f *expFrame) Formula {
	seen := bitset.New(uint(len(s.constNames)))
	vr := e.Variable.Name.Num
	conj := Formula(TrueNode{})
	for i := range e.Variable.Type {
		for _, obj := range s.typeObjs[e.Variable.Type[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}
			frame := f.push(vr, obj)
			conj = Conjunct(conj, e.Formula.expandQuants(s, frame))
			if _, ok := conj.(FalseNode); ok {
				return FalseNode{}
			}
		}
	}
	return conj
}

func (e *ExistsNode) expandQuants(s *Symtab, f *expFrame) Formula {
	seen := bitset.New(uint(len(s.constNames)))
	vr := e.Variable.Name.Num
	disj := Formula(FalseNode{})
	for i := range e.Variable.Type {
		for _, obj := range s.typeObjs[e.Variable.Type[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}
			frame := f.push(vr, obj)
			disj = Disjunct(disj, e.Formula.expandQuants(s, frame))
			if _, ok := disj.(TrueNode); ok {
				return TrueNode{}
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

func (e *AssignNode) expandQuants(*Symtab, *expFrame) Formula {
	// For now there is nothing to substitute because the
	// assignment can only be to total-cost
	return e
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
