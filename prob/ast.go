package prob

import "fmt"

type Domain struct {
	Name         string
	Requirements []string
	Types        []TypedName
	Constants    []TypedName
	Predicates   []Predicate
	Actions      []Action
}

type Problem struct {
	Name         string
	Domain       string
	Requirements []string
	Objects      []TypedName
	Init         []Formula
	Goal         Formula
	Metric       Metric
}

type Metric int

const (
	MetricMakespan Metric = iota
	MetricMinCost
)

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

func (l Loc) String() string {
	if l.Line < 0 {
		return l.File
	}
	return fmt.Sprintf("%s:%d", l.File, l.Line)
}

type Action struct {
	Name         string
	Parameters   []TypedName
	Precondition Formula
	Effect       Formula
}

type TypedName struct {
	Name
	Types []Name
}

type Predicate struct {
	Name
	Parameters []TypedName
}

type Term interface{}
type Variable struct{ Name }
type Constant struct{ Name }

type Formula interface {
	assignNums(*Symtab, *numFrame)
	findInertia(*Symtab)
	expandQuants(*Symtab, *expFrame) Formula
	dnf() Formula
	ensureDnf() // Panic if not in DNF
}

type LeafNode struct{}

type BinaryNode struct {
	Left, Right Formula
}

type UnaryNode struct {
	Formula Formula
}

type QuantNode struct {
	Variable TypedName
	UnaryNode
}

var (
	theTrueNode trueNode
	theFalseNode falseNode
)

type trueNode struct{ LeafNode }
func MakeTrue() Formula { return &theTrueNode }

type falseNode struct{ LeafNode }
func MakeFalse() Formula { return &theFalseNode }

type LiteralNode struct {
	Name
	Positive   bool
	Parameters []Term
	LeafNode
}

type AndNode struct{ BinaryNode }
type OrNode struct{ BinaryNode }
type NotNode struct{ UnaryNode }
type ForallNode struct{ QuantNode }
type ExistsNode struct{ QuantNode }

type WhenNode struct {
	Condition Formula
	UnaryNode
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
	"assign": OpAssign,
	//	"scale-up": OpScaleUp,
	//	"scale-down": OpScaleDown,
	//	"decrease": OpDecrease,
	"increase": OpIncrease,
}

type AssignNode struct {
	Op   AssignOp
	Lval string // Just total-cost for now.
	Rval string // Just a number
	LeafNode
}
