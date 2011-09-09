package main

type domain struct {
	name   string
	reqs   []string
	types  []typedName
	consts []typedName
	preds  []pred
	acts   []action
}

type typedName struct {
	name string
	typ  []string
}

type pred struct {
	name  string
	parms []typedName
}

type action struct {
	name   string
	parms  []typedName
	prec   gd
	effect effect
}

type gd interface{}

type gdBinary struct {
	left, right gd
}

type gdUnary struct {
	expr gd
}

type gdQuant struct {
	vr typedName
	gdUnary
}

type gdTrue int
type gdFalse int
type gdAnd gdBinary
type gdOr gdBinary
type gdNot gdUnary
type gdForall gdQuant
type gdExists gdQuant
type gdLiteral struct {
	pos   bool
	name  string
	parms []string
}

type effect interface{}

type effBinary struct {
	left, right effect
}

type effUnary struct {
	eff effect
}

type effNone int
type effAnd effBinary
type effForall struct {
	vr typedName
	effUnary
}
type effWhen struct {
	gd gd
	effUnary
}
type effLiteral struct {
	pos bool
	name string
	parms []string
}

type assignOp int

const (
	opAssign assignOp = iota
	opScaleUp
	opScaleDown
	opIncrease
	opDecrease
)

type fhead string	// Just total-cost for now
type fexp string	// Just a number for now

type effAssign struct {
	op assignOp
	lval fhead
	rval fexp
}