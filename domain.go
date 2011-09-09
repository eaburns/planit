package main

import "fmt"

type domain struct {
	name string
	reqs []string
	types []typedName
	consts []typedName
	preds []pred
	acts []action
}

type typedName struct {
	name string
	typ []string
}

type pred struct {
	name string
	parms []typedName
}

type action struct {
	name string
	parms []typedName
	prec gd
	effect *effect
}

type gd interface {
}

type gdBinary struct {
	left, right gd
}

func (gd gdBinary) String() string {
	return fmt.Sprintf("left: %v, right: %v", gd.left, gd.right)
}

type gdUnary struct {
	expr gd
}

func (gd gdUnary) String() string {
	return fmt.Sprintf("expr: %v", gd.expr)
}

type gdQuant struct {
	vr typedName
	gdUnary
}

func (gd gdQuant) String() string {
	return fmt.Sprintf("vr: %v, %v", gd.vr, gd.gdUnary)
}

type gdTrue int

func (gdTrue) String() string { return "gdTrue" }

type gdFalse int

func (gdFalse) String() string { return "gdFalse" }

type gdAnd gdBinary

func (gd gdAnd) String() string {
	return fmt.Sprintf("gdAnd{%v}", gdBinary(gd))
}

type gdOr gdBinary

func (gd gdOr) String() string {
	return fmt.Sprintf("gdOr{%v}", gdBinary(gd))
}

type gdNot gdUnary

func (gd gdNot) String() string {
	return fmt.Sprintf("gdNot{%v}", gdUnary(gd))
}

type gdForall gdQuant

func (gd gdForall) String() string {
	return fmt.Sprintf("gdForall{%v}", gdQuant(gd))
}

type gdExists gdQuant

func (gd gdExists) String() string {
	return fmt.Sprintf("gdExists[{%v}", gdQuant(gd))
}

type gdPred struct {
	name string
	parms []string
}

func (gd gdPred) String() string {
	return fmt.Sprintf("gdPred{name: %v, parms: %v}",
		gd.name, gd.parms)
}

type effect int