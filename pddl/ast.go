package pddl

import (
	"fmt"
	"io"
)

type Domain struct {
	Identifier
	Requirements []Identifier
	Types        []Type
	Constants    []TypedIdentifier
	Predicates   []Predicate
	Functions    []Function
	Actions      []Action
}

type Problem struct {
	Identifier
	Domain       Identifier
	Requirements []Identifier
	Objects      []TypedIdentifier
	Init         []Formula
	Goal         Formula
	Metric       Metric
}

type Metric int

const (
	MetricMakespan Metric = iota
	MetricMinCost
)

type Identifier struct {
	Str string
	Location
}

func (n Identifier) String() string {
	return n.Str
}

type Type struct {
	TypedIdentifier
	// Objects is a pointer to the definition
	// of each object of this type.
	Objects []*TypedIdentifier
}

type Action struct {
	Identifier
	Parameters   []TypedIdentifier
	Precondition Formula
	Effect       Formula
}

type TypedIdentifier struct {
	Identifier
	Num int
	Types []TypeName
}

type TypeName struct {
	Identifier
	Definition *Type
}

type Predicate struct {
	Identifier
	Num        int
	Parameters []TypedIdentifier
}

type Function struct {
	Identifier
	Num        int
	Types      []TypeName
	Parameters []TypedIdentifier
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

type Node struct{ Location }

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
	Variables []TypedIdentifier
	UnaryNode
}

type PropositionNode struct {
	Predicate Identifier
	Definition *Predicate
	Arguments  []Term
	Node
}

type Term struct {
	Identifier
	// Definition points to the variable
	// or constant definition for this term.
	Definition *TypedIdentifier
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
	Op   Identifier
	Lval Identifier   // Just total-cost for now.
	Rval string // Just a number
	Node
}

// Locer wraps the Loc method.
type Locer interface {
	Loc() Location
}

// A Location is a location in a PDDL input file.
type Location struct {
	File string
	Line int
}

func (l Location) Loc() Location {
	return l
}

func (l Location) String() string {
	if l.Line < 0 {
		return l.File
	}
	return fmt.Sprintf("%s:%d", l.File, l.Line)
}

// An Error holds information about errors
// assocated with locations in a PDDL file.
type Error struct {
	Location
	msg string
}

func (e Error) Error() string {
	return e.Location.String() + ": " + e.msg
}
 
// makeError returns an error at a location
// in a PDDL file with the message set  by a
// format string.
func makeError(l Locer, f string, vls ...interface{}) Error {
	return Error{ l.Loc(), fmt.Sprintf(f, vls...) }
}