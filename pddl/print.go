package pddl

import "fmt"

func (lit Literal) String() string {
	return fmt.Sprintf("Literal{Positive:%t, Name:%s, Parameters:%v}",
		lit.Positive, lit.Name, lit.Parameters)
}

func (gd gdBinary) String() string {
	return fmt.Sprintf("Left:%v, Right:%v", gd.Left, gd.Right)
}

func (gd gdUnary) String() string {
	return fmt.Sprintf("Expr:%v", gd.Expr)
}

func (gd gdQuant) String() string {
	return fmt.Sprintf("Variable:%v, %v", gd.Variable, gd.gdUnary)
}

func (GdTrue) String() string {
	return "GdTrue"
}

func (GdFalse) String() string {
	return "GdFalse"
}

func (gd GdAnd) String() string {
	return fmt.Sprintf("GdAnd{%v}", gdBinary(gd))
}

func (gd GdOr) String() string {
	return fmt.Sprintf("GdOr{%v}", gdBinary(gd))
}

func (gd GdNot) String() string {
	return fmt.Sprintf("GdNot{%v}", gdUnary(gd))
}

func (gd GdForall) String() string {
	return fmt.Sprintf("GdForall{%v}", gdQuant(gd))
}

func (gd GdExists) String() string {
	return fmt.Sprintf("GdExists[{%v}", gdQuant(gd))
}

func (gd GdLiteral) String() string {
	return fmt.Sprintf("%v", Literal(gd))
}

func (eff effBinary) String() string {
	return fmt.Sprintf("Left:%v, Right:%v", eff.Left, eff.Right)
}

func (eff effUnary) String() string {
	return fmt.Sprintf("Effect:%v", eff.Effect)
}

func (EffNone) String() string {
	return "effNone"
}

func (eff EffAnd) String() string {
	return fmt.Sprintf("EffAnd{%v}", effBinary(eff))
}

func (eff EffForall) String() string {
	return fmt.Sprintf("EffForall{Variable:%v, }", eff.Variable, eff.effUnary)
}

func (eff EffWhen) String() string {
	return fmt.Sprintf("EffWhen{Condition:%v, }", eff.Condition, eff.effUnary)
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

func (eff EffAssign) String() string {
	return fmt.Sprintf("EffAssign{Op:%v, Lval:%v, Rval:%v}",
		eff.Op, eff.Lval, eff.Rval)
}
