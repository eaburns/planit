// © 2013 the PlanIt Authors under the MIT license. See AUTHORS for the list of authors.

package pddl

import (
	"fmt"
	"io"
)

// A Domain represents a PDDL domain definition.
type Domain struct {
	// Name is the name of the domain.
	Name

	// Requirements is the requirement definitions.
	Requirements []Name

	// Types is the type definitions.
	Types []Type

	// Constants is the constant definitions.
	Constants []TypedEntry

	// Predicates is the predicate definitions.
	Predicates []Predicate

	// Functions is the function definitions.
	Functions []Function

	// Actions is the action definitions.
	Actions []Action
}

// A Problem represents a PDDL planning problem definition
type Problem struct {
	// Name is the problem name.
	Name

	// Domain is the name of the domain for
	// which this is a problem.
	Domain Name

	// Requirements is the requirement definitions.
	Requirements []Name

	// Objects is the object definitions.
	Objects []TypedEntry

	// Init is a conjunction of initial conditions
	// for the problem.
	Init []Formula

	// Goal is the problem goal formula.
	Goal Formula

	// Metric is the metric that must be optimized.
	Metric Metric
}

// A Metric represents planning metric that must be optimized.
type Metric int

const (
	// MetricMakespan asks the planner to optimize the number of actions in the plan.
	MetricMakespan Metric = iota

	// MetricMinCost asks the planner to minimize the total-cost function.
	MetricMinCost
)

// A Name represents the name of an entity.
type Name struct {
	Str string
	Location
}

func (n Name) String() string {
	return n.Str
}

// A Type represents a type definition.
type Type struct {
	TypedEntry

	// Supers is all of the predecessor types, including this current type.
	Supers []*Type

	// Domain is a pointer to the definition of each object of this type.
	Domain []*TypedEntry
}

// A TypedEntry is the entry of a typed list.
type TypedEntry struct {
	// Name is the name of the entry.
	Name

	// Num is a number assigned to this entry.  The number is unique within
	// the class of the entry: constants, variables, function, etc.
	Num int

	// Types is the disjunctive set of types for this entry.
	Types []TypeName
}

// A TypeName represents a name that is referring to a type.
type TypeName struct {
	// Name is the name of the type.
	Name

	// Definition is a pointer to the definition of the type to which this name refers.
	Definition *Type
}

// An Action represents an action definition.
type Action struct {
	// Name is the name of the action.
	Name

	// Parameters is a typed list of the parameter names for the action.
	Parameters []TypedEntry

	// Precondition is the action precondition formula.
	Precondition Formula

	// Effect is the action effect formula.
	Effect Formula
}

// A Predicate represents a predicate definition.
type Predicate struct {
	// Name is the name of the predicate.
	Name

	// Num is a unique number assigned the predicate.
	Num int

	// Parameters is a typed list of the predicate parameters.
	Parameters []TypedEntry

	// PosEffect and NegEffect are true if the predicate appears positively or negatively
	// (respectively) in an unconditional effect or as the consequent of a conditional effect.
	PosEffect, NegEffect bool
}

// A Function represents a function definition.
type Function struct {
	// Name is the name of the function.
	Name

	// Num is a unique number assigned to the function.
	Num int

	// Types is a disjunctive list of the types for the evaluation of this function.
	Types []TypeName

	// Parameters is a typed list of the function parameters.
	Parameters []TypedEntry
}

// A Formula represents either a PDDL goal description (GD), or an expression.
type Formula interface {
	// print prints the formula as valid PDDL to an io.Writed, prefixed with a string
	// for indentation purposes.
	print(io.Writer, string)

	// check panicks an error if there is a semantic error in the formula.
	check(defs, *errors)
}

// A Node is a node in the formula tree.
type Node struct{ Location }

// A UnaryNode is a node with only a single successor.
type UnaryNode struct {
	Node
	Formula Formula
}

// A BinaryNode is a node with two successors.
type BinaryNode struct {
	Node
	Left, Right Formula
}

// A MultiNode is a node with a slice of successors.
type MultiNode struct {
	Node
	Formula []Formula
}

// A QuantNode is a node with a single successor that also declares a typed list of variables.
type QuantNode struct {
	Variables []TypedEntry
	UnaryNode
}

// A LiteralNode represents the instantiation of a predicate.
type LiteralNode struct {
	Node

	// Predicate is the name of the predicate.
	Predicate Name

	// Negative is true if this literal is negative, or it is false if the predicate is positive.
	Negative bool

	// Arguments are the terms that are passed as the arguments to this instantiation.
	Arguments []Term

	// IsEffect is true if the literal is appearing in an unconditional effect or as a
	// consequent of a conditional effect. This is used to determine inertia for
	// the literal's predicate.
	IsEffect bool

	// Definition is a pointer to the definition of the predicate to which this literal refers.
	Definition *Predicate
}

// A Term represents either a constant or a variable.
type Term struct {
	// Name is the name of the term.
	Name

	// Variable is true if this term is referring to a variable and it is false if this term
	// is referring to a constant.
	Variable bool

	// Definition points to the variable or constant definition for this term.
	Definition *TypedEntry
}

// An AndNode represents a conjunction of it successors.
type AndNode struct{ MultiNode }

// An OrNode represents a disjunction of its successors.
type OrNode struct{ MultiNode }

// A NotNode represents the negation of its successor.
type NotNode struct{ UnaryNode }

// An ImplyNode represents an antecedent and its consequent.
type ImplyNode struct{ BinaryNode }

// A ForallNode represents a universal quantifier.
type ForallNode struct {
	QuantNode

	// IsEffect is true if the literal is appearing in an unconditional effect or as a
	// consequent of a conditional effect. This is used to distinguish between
	// the need to require :universal-preconditions and :conditional-effects.
	IsEffect bool
}

// An ExistsNode represents an existential quantifier.
type ExistsNode struct{ QuantNode }

// A WhenNode represents a conditional effect.
type WhenNode struct {
	// Condition is the condition of the conditional effect.
	Condition Formula

	// The Formula of the UnaryNode is the consequent of the conditional effect.
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

// An AssignNode represents the assingment of a value to a function.
type AssignNode struct {
	Node

	// Op is the assignment operation.
	Op Name

	// Lval is the function to which a value is being assigned.
	Lval Fhead

	// IsNumber is true if the right-hand-side is a number, in which case the Number field
	// is valid and the Fhead field is not. If IsNumber is false, then the opposite is the case.
	IsNumber bool

	// Number is valid if IsNumber is true; it is a string representing the assigned number.
	Number string

	// Fhead is valid if IsNumber is false; it is the assigned function instantiation.
	Fhead Fhead

	// IsInit is true if the assignment is appearing in the :init section of a problem.
	IsInit bool
}

// Fhead represents a function instantiation.
type Fhead struct {
	// Name is the name of the function.
	Name

	// Arguments is the slice of terms used as the arguments to the function's parameters.
	Arguments []Term

	// Definition is a pointer to the definition of the function to which this Fhead refers.
	Definition *Function
}

// Locer wraps the Loc method.
type Locer interface {
	Loc() Location
}

// A Location is a location in a PDDL input file.
type Location struct {
	// File is the file name.
	File string
	// Line is the line number.
	Line int
}

// Loc returns the Location, implementing the Locer interface.
func (l Location) Loc() Location {
	return l
}

// String returns a human-readable string representation of the location.
func (l Location) String() string {
	if l.Line < 0 {
		return l.File
	}
	return fmt.Sprintf("%s:%d", l.File, l.Line)
}

// An Error holds information about errors assocated with locations in a PDDL file.
type Error struct {
	// Location is the location of the cause of the error.
	Location

	// msg is the error's message.
	msg string
}

func (e Error) Error() string {
	return e.Location.String() + ": " + e.msg
}

// Errorf panicks with an error at a location in a PDDL file.  The message is set by a format string.
func errorf(l Locer, f string, vls ...interface{}) Error {
	panic(Error{l.Loc(), fmt.Sprintf(f, vls...)})
}
