package main

import "fmt"

func (gd gdBinary) String() string {
	return fmt.Sprintf("left: %v, right: %v", gd.left, gd.right)
}

func (gd gdUnary) String() string {
	return fmt.Sprintf("expr: %v", gd.expr)
}

func (gd gdQuant) String() string {
	return fmt.Sprintf("vr: %v, %v", gd.vr, gd.gdUnary)
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

func (gd gdPred) String() string {
	return fmt.Sprintf("gdPred{name: %v, parms: %v}",
		gd.name, gd.parms)
}