package lifted

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
	Type []string
}

type Predicate struct {
	Name       string
	Parameters []TypedName
}

type Action struct {
	Name         string
	Parameters   []TypedName
	Precondition Expr
	Effect       Effect
}

type Literal struct {
	Positive   bool
	Name       string
	Parameters []string
}

type Expr interface{}

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

type Effect interface{}

type EffBinary struct {
	Left, Right Effect
}

type EffUnary struct {
	Effect Effect
}

type EffNone int
type EffAnd EffBinary
type EffForall struct {
	Variable TypedName
	EffUnary
}
type EffWhen struct {
	Condition Expr
	EffUnary
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

type EffAssign struct {
	Op   AssignOp
	Lval Fhead
	Rval Fexp
}

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
