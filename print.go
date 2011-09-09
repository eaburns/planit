package main

import "fmt"

func (gd gdBinary) String() string {
	return fmt.Sprintf("left:%v, right:%v", gd.left, gd.right)
}

func (gd gdUnary) String() string {
	return fmt.Sprintf("expr:%v", gd.expr)
}

func (gd gdQuant) String() string {
	return fmt.Sprintf("vr:%v, %v", gd.vr, gd.gdUnary)
}

func (gdTrue) String() string {
	return "gdTrue"
}

func (gdFalse) String() string {
	return "gdFalse"
}

func (gd gdAnd) String() string {
	return fmt.Sprintf("gdAnd{%v}", gdBinary(gd))
}

func (gd gdOr) String() string {
	return fmt.Sprintf("gdOr{%v}", gdBinary(gd))
}

func (gd gdNot) String() string {
	return fmt.Sprintf("gdNot{%v}", gdUnary(gd))
}

func (gd gdForall) String() string {
	return fmt.Sprintf("gdForall{%v}", gdQuant(gd))
}

func (gd gdExists) String() string {
	return fmt.Sprintf("gdExists[{%v}", gdQuant(gd))
}

func (gd gdLiteral) String() string {
	return fmt.Sprintf("gdLiteral{pos:%t, name:%s, parms:%v}",
		gd.pos, gd.name, gd.parms)
}

func (eff effBinary) String() string {
	return fmt.Sprintf("left:%v, right:%v", eff.left, eff.right)
}

func (eff effUnary) String() string {
	return fmt.Sprintf("eff:%v", eff.eff)
}

func (effNone) String() string {
	return "effNone"
}

func (eff effAnd) String() string {
	return fmt.Sprintf("effAnd{%v}", effBinary(eff))
}

func (eff effForall) String() string {
	return fmt.Sprintf("effForall{vr:%v, }", eff.vr, eff.effUnary)
}

func (eff effWhen) String() string {
	return fmt.Sprintf("effWhen{gd:%v, }", eff.gd, eff.effUnary)
}

func (eff effLiteral) String() string {
	return fmt.Sprintf("effLiteral{pos:%t, name:%s, parms:%v}",
		eff.pos, eff.name, eff.parms)
}

var assignOpNames = map[assignOp]string{
	opAssign: "opAssign",
	opScaleUp: "opScaleUp",
	opScaleDown: "opScaleDown",
	opIncrease: "opIncrease",
	opDecrease: "opDecrease",
}

func (o assignOp) String() string {
	return assignOpNames[o]
}

func (eff effAssign) String() string {
	return fmt.Sprintf("effAssign{op:%v, lval:%v, rval:%v}",
		eff.op, eff.lval, eff.rval)
}