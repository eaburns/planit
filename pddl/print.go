package pddl

import (
	"fmt"
	"io"
)

const indent = "\t"

// PrintDomain prints the domain in valid PDDL to the given writer.
func PrintDomain(w io.Writer, d *Domain) {
	fmt.Fprintf(w, "(define (domain %s)\n", d.Name)
	printRequirements(w, d.Requirements)

	if len(d.Types) > 0 {
		fmt.Fprintf(w, "%s(:types", indent)
		printTypedNames(w, "\n"+indent+indent, d.Types)
		fmt.Fprintln(w, ")")
	}

	if len(d.Constants) > 0 {
		fmt.Fprintf(w, "%s(:constants", indent)
		printTypedNames(w, "\n"+indent+indent, d.Constants)
		fmt.Fprintln(w, ")")
	}

	if len(d.Predicates) > 0 {
		fmt.Fprintf(w, "%s(:predicates\n", indent)
		for i, p := range d.Predicates {
			fmt.Fprintf(w, "%s(%s", indent+indent, p.Name.Str)
			printTypedNames(w, " ", p.Parameters)
			fmt.Fprint(w, ")")
			if i < len(d.Predicates)-1 {
				fmt.Fprint(w, "\n")
			}
		}
		fmt.Fprintln(w, ")")
	}

	for _, act := range d.Actions {
		fmt.Fprintf(w, "%s(:action %s\n", indent, act.Name)
		fmt.Fprintf(w, "%s:parameters (", indent+indent+indent)
		printTypedNames(w, "", act.Parameters)
		fmt.Fprint(w, ")")
		if act.Precondition != nil {
			fmt.Fprint(w, "\n")
			fmt.Fprintf(w, "%s:precondition\n", indent+indent+indent)
			act.Precondition.print(w, indent+indent+indent+indent)
		}
		if act.Effect != nil {
			fmt.Fprint(w, "\n")
			fmt.Fprintf(w, "%s:effect\n", indent+indent+indent)
			act.Effect.print(w, indent+indent+indent+indent)
		}
		fmt.Fprintln(w, ")")
	}

	fmt.Fprintln(w, ")")
}

// PrintProblem prints the problem in valid PDDL to the given writer.
func PrintProblem(w io.Writer, p *Problem) {
	fmt.Fprintf(w, "(define (problem %s)\n%s(:domain %s)\n", p.Name, indent, p.Domain)
	printRequirements(w, p.Requirements)

	if len(p.Objects) > 0 {
		fmt.Fprintf(w, "%s(:objects", indent)
		printTypedNames(w, "\n"+indent+indent, p.Objects)
		fmt.Fprintln(w, ")")
	}

	fmt.Fprintf(w, "%s(:init", indent)
	for _, f := range p.Init {
		fmt.Fprint(w, "\n")
		f.print(w, indent+indent)
	}
	fmt.Fprint(w, ")\n")

	fmt.Fprintf(w, "%s(:goal\n", indent)
	p.Goal.print(w, indent+indent)

	fmt.Fprintln(w, ")\n)")
}

func printRequirements(w io.Writer, reqs []Name) {
	if len(reqs) > 0 {
		fmt.Fprintf(w, "%s(:requirements\n", indent)
		for i, r := range reqs {
			s := r.Str
			if i == len(reqs)-1 {
				s += ")"
			}
			fmt.Fprintln(w, indent+indent, s)
		}
	}
}

// declGroup is a group of declarators along
// with their type.
type declGroup struct {
	typ string
	ents []string
}

// declGroups implements sort.Interface, sorting
// the list of typed declarations by their type name.
type declGroups []declGroup

func (t declGroups) Len() int {
	return len(t)
}

func (t declGroups) Less(i, j int) bool {
	return t[i].typ < t[j].typ
}

func (t declGroups) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// printTypedNames prints a slice of TypedNames
// in their given order. The prefix is printed before
// each group.
func printTypedNames(w io.Writer, prefix string, ns []TypedName) {
	if len(ns) == 0 {
		return
	}
	tprev := typeString(ns[0].Types)
	sep := prefix
	for _, n := range ns {
		tcur := typeString(n.Types)
		if tcur != tprev {
			if tprev == "" {
				// Should be impossible.
				panic(n.Loc.String() + ": untyped declarations in the middle of a typed list")
			}
			fmt.Fprintf(w, " - %s", tprev)
			tprev = tcur
			sep = prefix
			if sep == "" {
				sep = " "
			}
		}
		fmt.Fprintf(w, "%s%s", sep, n.Str)
		sep = " "
	}
	if tprev != "" {
		fmt.Fprintf(w, " - %s", tprev)
	}
}

// typeString returns the string representation of a type.
func typeString(t []Name) (str string) {
	switch len(t) {
	case 0:
		break
	case 1:
		str = t[0].Str
	default:
		str = "(either"
		for _, n := range t {
			str += " " + n.Str
		}
		str += ")"
	}
	return
}

func (l *PropositionNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(", prefix)
	fmt.Fprint(w, l.Predicate.Str)
	for _, t := range l.Parameters {
		fmt.Fprintf(w, " %s", t.Name.Str)
	}
	fmt.Fprint(w, ")")
}

func (n *AndNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(and", prefix)
	for _, f := range n.Formula {
		fmt.Fprint(w, "\n")
		f.print(w, prefix+indent)
	}
	fmt.Fprint(w, ")")
}

func (n *OrNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(or", prefix)
	for _, f := range n.Formula {
		fmt.Fprint(w, "\n")
		f.print(w, prefix+indent)
	}
	fmt.Fprint(w, ")")
}

func (n *NotNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(not\n", prefix)
	n.Formula.print(w, prefix+indent)
	fmt.Fprint(w, ")")
}

func (n *ImplyNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(imply\n", prefix)
	n.Left.print(w, prefix+indent)
	fmt.Fprint(w, "\n")
	n.Right.print(w, prefix+indent)
	fmt.Fprint(w, ")")
}

func (n *ForallNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(forall (", prefix)
	printTypedNames(w, "", n.Variables)
	fmt.Fprint(w, ")\n")
	n.Formula.print(w, prefix+indent)
	fmt.Fprint(w, ")")
}

func (n *ExistsNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(exists (", prefix)
	printTypedNames(w, "", n.Variables)
	fmt.Fprint(w, ")\n")
	n.Formula.print(w, prefix+indent)
	fmt.Fprint(w, ")")
}

func (n *WhenNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(when\n", prefix)
	n.Condition.print(w, prefix+indent)
	fmt.Fprint(w, "\n")
	n.Formula.print(w, prefix+indent)
	fmt.Fprint(w, ")")
}

func (n *AssignNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(%s %s %s)", prefix, n.Op, n.Lval, n.Rval)
}
