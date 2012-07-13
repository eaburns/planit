package pddl

import (
	"fmt"
	"io"
)

type Domain struct {
	Name
	Requirements []Name
	Types        []Type
	Constants    []TypedEntry
	Predicates   []Predicate
	Functions    []Function
	Actions      []Action
}

type Problem struct {
	Name
	Domain       Name
	Requirements []Name
	Objects      []TypedEntry
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
	Location
}

func (n Name) String() string {
	return n.Str
}

type Type struct {
	TypedEntry

	// Supers is all of the predecessor types,
	// including this current type.
	Supers []*Type

	// Objects is a pointer to the definition
	// of each object of this type.
	Objects []*TypedEntry
}

type Action struct {
	Name
	Parameters   []TypedEntry
	Precondition Formula
	Effect       Formula
}

type TypedEntry struct {
	Name
	Num   int
	Types []TypeName
}

type TypeName struct {
	Name

	Definition *Type
}

type Predicate struct {
	Name
	Num        int
	Parameters []TypedEntry

	// PosEffect and NegEffect are true if the predicate
	// appears positively or negatively (respectively)
	// in an unconditional effect or as the consequent
	// of a conditional effect.
	PosEffect, NegEffect bool
}

type Function struct {
	Name
	Num        int
	Types      []TypeName
	Parameters []TypedEntry
}

type Formula interface {
	// print prints the formula as valid PDDL
	// to an io.Writed, prefixed with a string
	// for indentation purposes.
	print(io.Writer, string)

	// check returns an error if there is a semantic
	// error in the formula, otherwise it returns
	// nil.
	check(defs) error
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
	Variables []TypedEntry
	UnaryNode
}

type LiteralNode struct {
	Predicate  Name
	Negative   bool
	Definition *Predicate
	Arguments  []Term
	Node

	// IsEffect is true if the literal is appearing
	// in an unconditional effect or as a
	// consequent of a conditional effect.
	// This is used to determine inertia for
	// the literal's predicate.
	IsEffect bool
}

type Term struct {
	Name
	Variable bool

	// Definition points to the variable
	// or constant definition for this term.
	Definition *TypedEntry
}

type AndNode struct{ MultiNode }

type OrNode struct{ MultiNode }

type NotNode struct{ UnaryNode }

type ImplyNode struct{ BinaryNode }

type ForallNode struct {
	QuantNode

	// IsEffect is true if the literal is appearing
	// in an unconditional effect or as a
	// consequent of a conditional effect.
	// This is used to distinguish between
	// the need to require :universal-preconditions
	// and :conditional-effects.
	IsEffect bool
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
	Op   Name
	Lval Fhead

	// IsNumber is true if the right-hand-side
	// is a number, in which case the Number
	// field is valid and the Fhead field is not.
	// If IsNumber is false, then the opposite
	// is the case.
	IsNumber bool
	Number   string
	Fhead    Fhead

	Node
}

type Fhead struct {
	Name
	Definition *Function
	Arguments  []Term
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
	return Error{l.Loc(), fmt.Sprintf(f, vls...)}
}
