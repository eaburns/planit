package lifted

import (
	"github.com/willf/bitset"
	"log"
)

func (l *LiteralNode) expandQuants(s *Symtab, f *expFrame) Formula {
	parms := make([]Term, len(l.Parameters))
	copy(parms, l.Parameters)

	for i, _ := range parms {
		if parms[i].Kind == TermConstant {
			continue
		}
		vl, ok := f.lookup(parms[i].Name.Num)
		if !ok {
			// Previous pass should have ensured
			// that this is already bound.
			log.Fatalf("%s: Unbound variable: %s", parms[i].Name.Loc,
				parms[i].Name.Str)
		}
		parms[i].Kind = TermConstant
		parms[i].Name.Num = vl
		parms[i].Name.Str = s.constNames[vl]
	}

	return &LiteralNode{
		Positive: l.Positive,
		Name: l.Name,
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
		res = FalseNode(0)
	default:
		res = Conjunct(l, e.Right.expandQuants(s, f))
	}
	return
}

func (e *OrNode) expandQuants(s *Symtab, f *expFrame) (res Formula) {
	switch l := e.Left.expandQuants(s, f).(type) {
	case TrueNode:
		res = TrueNode(1)
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
	conj := Formula(TrueNode(1))
	for i, _ := range e.Variable.Type {
		for obj := range s.typeObjs[e.Variable.Type[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}
			frame := f.push(vr, obj)
			conj = Conjunct(conj, e.Formula.expandQuants(s, frame))
			if _, ok := conj.(FalseNode); ok {
				return FalseNode(0)
			}
		}
	}
	return conj
}

func (e *ExistsNode) expandQuants(s *Symtab, f *expFrame) Formula {
	seen := bitset.New(uint(len(s.constNames)))
	vr := e.Variable.Name.Num
	disj := Formula(FalseNode(0))
	for i, _ := range e.Variable.Type {
		for obj := range s.typeObjs[e.Variable.Type[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}
			frame := f.push(vr, obj)
			disj = Disjunct(disj, e.Formula.expandQuants(s, frame))
			if _, ok := disj.(TrueNode); ok {
				return TrueNode(1)
			}
		}
	}
	return disj
}

func (e *WhenNode) expandQuants(*Symtab, *expFrame) Formula {
	return nil
}

func (e *AssignNode) expandQuants(*Symtab, *expFrame) Formula {
	// For now there is nothing to substitute because the
	// assignment can only be to total-cost
	return e
}

type expFrame struct {
	vr, vl	int
	up	*expFrame
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
	return &expFrame { vr: vr, vl: vl, up: f }
}