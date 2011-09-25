package lifted

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
	return nil
}

func (e *ExprExists) expandQuants(s *Symtab, f *expFrame) Expr {
	return nil
}

func (e *ExprLiteral) expandQuants(s *Symtab, f *expFrame) Expr {
	return nil
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