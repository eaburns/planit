package lifted

import "github.com/willf/bitset"

func (l *Literal) expandQuants(s *Symtab, f *expFrame) *Literal {
	return nil
}

func (e ExprTrue) expandQuants(*Symtab, *expFrame) Expr {
	return e
}

func (e ExprFalse) expandQuants(*Symtab, *expFrame) Expr {
	return e
}

func (e *ExprAnd) expandQuants(s *Symtab, f *expFrame) (res Expr) {
	switch l := e.Left.expandQuants(s, f).(type) {
	case ExprTrue:
		res = e.Right.expandQuants(s, f)
	case ExprFalse:
		res = ExprFalse(0)
	default:
		res = ExprConj(l, e.Right.expandQuants(s, f))
	}
	return
}

func (e *ExprOr) expandQuants(s *Symtab, f *expFrame) (res Expr) {
	switch l := e.Left.expandQuants(s, f).(type) {
	case ExprTrue:
		res = ExprTrue(1)
	case ExprFalse:
		res = e.Right.expandQuants(s, f)
	default:	
		res = ExprDisj(l, e.Right.expandQuants(s, f))
	}
	return
}

func (e *ExprNot) expandQuants(s *Symtab, f *expFrame) Expr {
	return ExprNeg(e.Expr.expandQuants(s, f))
}

func (e *ExprForall) expandQuants(s *Symtab, f *expFrame) Expr {
	seen := bitset.New(uint(len(s.constNames)))
	vr := e.Variable.Name.Num
	conj := Expr(ExprTrue(1))
	for i, _ := range e.Variable.Type {
		for obj := range s.typeObjs[e.Variable.Type[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}
			frame := f.push(vr, obj)
			conj = ExprConj(conj, e.Expr.expandQuants(s, frame))
			if _, ok := conj.(ExprFalse); ok {
				return ExprFalse(0)
			}
		}
	}
	return conj
}

func (e *ExprExists) expandQuants(s *Symtab, f *expFrame) Expr {
	seen := bitset.New(uint(len(s.constNames)))
	vr := e.Variable.Name.Num
	disj := Expr(ExprFalse(0))
	for i, _ := range e.Variable.Type {
		for obj := range s.typeObjs[e.Variable.Type[i].Num] {
			if seen.Test(uint(obj)) {
				continue
			}
			frame := f.push(vr, obj)
			disj = ExprDisj(disj, e.Expr.expandQuants(s, frame))
			if _, ok := disj.(ExprTrue); ok {
				return ExprTrue(1)
			}
		}
	}
	return disj
}

func (e *ExprLiteral) expandQuants(s *Symtab, f *expFrame) Expr {
	return (*ExprLiteral)((*Literal)(e).expandQuants(s, f))
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