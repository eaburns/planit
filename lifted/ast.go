package lifted

import "os"

type Domain struct {
	Name         string
	Requirements []string
	Types        []TypedName
	Constants    []TypedName
	Predicates   []Predicate
	Actions      []Action
}

type TypedName struct {
	Name string
	Num int
	Type []string
}

type Predicate struct {
	Name       string
	Num int
	Parameters []TypedName
}

type Action struct {
	Name         string
	Parameters   []TypedName
	Precondition Expr
	Effect       Effect
}

type TermKind int

const (
	TermVariable TermKind = iota
	TermConstant
)

type Term struct {
	Kind TermKind
	Name string
	Num int
	Loc  string
}

type Literal struct {
	Positive   bool
	Name       string
	Num	int
	Parameters []Term
}

type Expr interface {
	UniquifyVars(*uniqFrame) os.Error
	ExpandQuants(*expandFrame) Expr
}

type ExprBinary struct {
	Left, Right Expr
}

type ExprUnary struct {
	Expr Expr
}

type ExprQuant struct {
	Variable TypedName
	ExprUnary
}

type ExprTrue int
type ExprFalse int
type ExprAnd ExprBinary
type ExprOr ExprBinary
type ExprNot ExprUnary
type ExprForall ExprQuant
type ExprExists ExprQuant
type ExprLiteral Literal

type Effect interface {
	UniquifyVars(*uniqFrame) os.Error
	ExpandQuants(*expandFrame) Effect
}

type EffectBinary struct {
	Left, Right Effect
}

type EffectUnary struct {
	Effect Effect
}

type EffectNone int
type EffectAnd EffectBinary
type EffectForall struct {
	Variable TypedName
	EffectUnary
}
type EffectWhen struct {
	Condition Expr
	EffectUnary
}
type EffectLiteral Literal
type EffectAssign struct {
	Op   AssignOp
	Lval Fhead
	Rval Fexp
}

type AssignOp int

const (
	OpAssign AssignOp = iota
	OpScaleUp
	OpScaleDown
	OpIncrease
	OpDecrease
)

var AssignOps = map[string]AssignOp{
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
type Fexp string // Just a number for now

type Problem struct {
	Name         string
	Domain       string
	Requirements []string
	Objects      []TypedName
	Init         []InitEl
	Goal         Expr
	Metric       Metric
}

type Metric int

const (
	MetricMakespan Metric = iota
	MetricMinCost
)

type InitEl interface{}

type InitLiteral Literal

type InitEq struct {
	Lval Fhead
	Rval string
}
