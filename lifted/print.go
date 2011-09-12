package lifted

import "fmt"

func (k TermKind) String() string {
	switch k {
	case TermVariable:
		return "TermVariable"
	case TermConstant:
		return "TermConstant"
	}

	return fmt.Sprintf("%d", int(k))
}

func (lit *Literal) String() string {
	return fmt.Sprintf("Literal{Positive:%t, Name:%s, Parameters:%v}",
		lit.Positive, lit.Name, lit.Parameters)
}

func (e *ExprBinary) String() string {
	return fmt.Sprintf("Left:%v, Right:%v", e.Left, e.Right)
}

func (e *ExprUnary) String() string {
	return fmt.Sprintf("Expr:%v", e.Expr)
}

func (e *ExprQuant) String() string {
	return fmt.Sprintf("Variable:%v, %v", e.Variable, e.ExprUnary)
}

func (ExprTrue) String() string {
	return "ExprTrue"
}

func (ExprFalse) String() string {
	return "ExprFalse"
}

func (e *ExprAnd) String() string {
	return fmt.Sprintf("ExprAnd{%v}", (*ExprBinary)(e))
}

func (e *ExprOr) String() string {
	return fmt.Sprintf("ExprOr{%v}", (*ExprBinary)(e))
}

func (e *ExprNot) String() string {
	return fmt.Sprintf("ExprNot{%v}", (*ExprUnary)(e))
}

func (e *ExprForall) String() string {
	return fmt.Sprintf("ExprForall{%v}", (*ExprQuant)(e))
}

func (e *ExprExists) String() string {
	return fmt.Sprintf("ExprExists[{%v}", (*ExprQuant)(e))
}

func (e *ExprLiteral) String() string {
	return fmt.Sprintf("%v", (*Literal)(e))
}

func (eff *EffectBinary) String() string {
	return fmt.Sprintf("Left:%v, Right:%v", eff.Left, eff.Right)
}

func (eff *EffectUnary) String() string {
	return fmt.Sprintf("Effect:%v", eff.Effect)
}

func (EffectNone) String() string {
	return "effNone"
}

func (eff *EffectAnd) String() string {
	return fmt.Sprintf("EffAnd{%v}", (*EffectBinary)(eff))
}

func (eff *EffectForall) String() string {
	return fmt.Sprintf("EffForall{Variable:%v, }", eff.Variable,
		eff.EffectUnary)
}

func (eff *EffectWhen) String() string {
	return fmt.Sprintf("EffWhen{Condition:%v, }", eff.Condition,
		eff.EffectUnary)
}

func (eff *EffectAssign) String() string {
	return fmt.Sprintf("EffAssign{Op:%v, Lval:%v, Rval:%v}",
		eff.Op, eff.Lval, eff.Rval)
}

var assignOpNames = map[AssignOp]string{
	OpAssign:    "OpAssign",
	OpScaleUp:   "OpScaleUp",
	OpScaleDown: "OpScaleDown",
	OpIncrease:  "OpIncrease",
	OpDecrease:  "OpDecrease",
}

func (o AssignOp) String() string {
	return assignOpNames[o]
}
