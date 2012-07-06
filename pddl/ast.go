package pddl

import (
	"fmt"
	"io"
)

type Domain struct {
	Name         string
	Requirements []Name
	Types        []TypedName
	Constants    []TypedName
	Predicates   []Predicate
	Functions    []Function
	Actions      []Action
}

type Problem struct {
	Name         string
	Domain       string
	Requirements []Name
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
	Loc Loc
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
	Num int
	// A disjunction of types.
	Types []Type
}

type Type struct {
	Name
	Definition *TypedName
}

type Predicate struct {
	Name
	Num        int
	Parameters []TypedName
}

type Function struct {
	Name
	Num        int
	Types      []Type
	Parameters []TypedName
}

type Formula interface {
	// print prints the formula as valid PDDL
	// to an io.Writed, prefixed with a string
	// for indentation purposes.
	print(io.Writer, string)

	// check returns an error if there is a semantic
	// error in the formula, otherwise it returns
	// nil.
	check(*defs) error
}

type Node struct{ Loc Loc }

type UnaryNode struct {
	Node
	Formula Formula
}

type BinaryNode struct {
	Node
	Left, Right Formula
}

type MultiNode struct {
	Node
	Formula []Formula
}

type QuantNode struct {
	Variables []TypedName
	UnaryNode
}

type PropositionNode struct {
	Predicate  string
	Definition *Predicate
	Arguments  []Term
	Node
}

type Term struct {
	Name
	// Definition points to the variable
	// or constant definition for this term.
	Definition *TypedName
	Variable   bool
}

type AndNode struct{ MultiNode }

type OrNode struct{ MultiNode }

type NotNode struct{ UnaryNode }

type ImplyNode struct{ BinaryNode }

type ForallNode struct {
	QuantNode

	// Effect is true if this node is an effect.
	// This is used to distinguish between
	// the need to require :universal-preconditions
	// and :conditional-effects.
	Effect bool
}

type ExistsNode struct{ QuantNode }

type WhenNode struct {
	Condition Formula
	UnaryNode
}

var (
	// AssignOps is the set of valid assignment operators.
	AssignOps = map[string]bool{
		"=":        true,
		"assign":   true,
		"increase": true,
	}
)

type AssignNode struct {
	Op   string
	Lval Name   // Just total-cost for now.
	Rval string // Just a number
	Node
}
