package lifted

import "fmt"

type Domain struct {
	Name         string
	Requirements []string
	Types        []TypedName
	Constants    []TypedName
	Predicates   []Predicate
	Actions      []Action
}

type Name struct {
	Str string
	Num int
	Loc Loc
}

func MakeName(s string, l Loc) Name {
	return Name{Str: s, Num: -1, Loc: l}
}

type Loc struct {
	File string
	Line int
}

func (l *Loc) String() string {
	if l.Line < 0 {
		return l.File
	}
	return fmt.Sprintf("%s:%d", l.File, l.Line)
}

type Action struct {
	Name         string
	Parameters   []TypedName
	Precondition Expr
	Effect       Effect
}

type TypedName struct {
	Name Name
	Type []Name
}

type Predicate struct {
	Name       Name
	Parameters []TypedName
}

type TermKind int

const (
	TermVariable TermKind = iota
	TermConstant
)

type Term struct {
	Kind TermKind
	Name Name
}

type Literal struct {
	Positive   bool
	Name       Name
	Parameters []Term
}

type Expr interface {
	assignNums(*Symtab, *numFrame)
	expandQuants(*Symtab, *expFrame) Expr
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
	assignNums(*Symtab, *numFrame)
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

type InitEl interface {
	assignNums(*Symtab, *numFrame)
}

type InitLiteral Literal

type InitEq struct {
	Lval Fhead
	Rval string
}
