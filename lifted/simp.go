package lifted

func ExprConj(l Expr, r Expr) Expr {
	switch l.(type) {
	case ExprTrue:
		return r
	case ExprFalse:
		return ExprFalse(0)
	}
	switch r.(type) {
	case ExprTrue:
		return l
	case ExprFalse:
		return ExprFalse(0)
	}
	return &ExprAnd{Left: l, Right: r}
}

func ExprDisj(l Expr, r Expr) Expr {
	switch l.(type) {
	case ExprTrue:
		return ExprTrue(1)
	case ExprFalse:
		return r
	}
	switch r.(type) {
	case ExprTrue:
		return ExprTrue(1)
	case ExprFalse:
		return l
	}
	return &ExprOr{Left: l, Right: r}
}

func ExprNeg(e Expr) Expr {
	switch e.(type) {
	case *ExprNot:
		return e.(*ExprNot).Expr
	}
	return &ExprNot{Expr: e}
}

func EffectConj(l Expr, r Expr) Expr {
	switch l.(type) {
	case EffectNone:
		return r
	}
	switch r.(type) {
	case EffectNone:
		return l
	}
	return &ExprAnd{Left: l, Right: r}
}