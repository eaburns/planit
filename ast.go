package pddl

type Domain struct {
	Name   string
	Reqs   []string
	Types  []TypedName
	Consts []TypedName
	Preds  []Predicate
	Acts   []Action
}

type TypedName struct {
	Name string
	Typ  []string
}

type Predicate struct {
	Name  string
	Parms []TypedName
}

type Action struct {
	Name   string
	Parms  []TypedName
	Prec   Gd
	Effect Effect
}

type Literal struct {
	Positive   bool
	Name  string
	Parms []string
}

type Gd interface{}

type gdBinary struct {
	Left, Right Gd
}

type gdUnary struct {
	Expr Gd
}

type gdQuant struct {
	Variable TypedName
	gdUnary
}

type GdTrue int
type GdFalse int
type GdAnd gdBinary
type GdOr gdBinary
type GdNot gdUnary
type GdForall gdQuant
type GdExists gdQuant
type GdLiteral Literal

type Effect interface{}

type effBinary struct {
	Left, Right Effect
}

type effUnary struct {
	Effect Effect
}

type EffNone int
type EffAnd effBinary
type EffForall struct {
	Variable TypedName
	effUnary
}
type EffWhen struct {
	Condition Gd
	effUnary
}
type EffLiteral Literal

type AssignOp int

const (
	OpAssign AssignOp = iota
	OpScaleUp
	OpScaleDown
	OpIncrease
	OpDecrease
)

var assignOps = map[string]AssignOp{
	//	"assign": OpAssign,
	//	"scale-up": OpScaleUp,
	//	"scale-down": OpScaleDown,
	//	"decrease": OpDecrease,
	// Just support increase for now for :Action-costs
	"increase": OpIncrease,
}

// Just total-cost for now
type Fhead struct {
	Name string
}
type Fexp string  // Just a number for now

type EffAssign struct {
	Op   AssignOp
	Lval Fhead
	Rval Fexp
}

type problem struct {
	Name string
	Domain string
	Reqs []string
	Objs []TypedName
	Init []InitEl
	Goal Gd
	Metric Metric
}

type Metric int

const (
	MetricMakespan Metric = iota
	MetricMinCost
)

type InitEl interface {}

type InitLiteral Literal

type InitEq struct {
	Lval Fhead
	Rval string
}