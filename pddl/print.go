package pddl

import (
	"fmt"
	"io"
	"sort"
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

// printTypedNames prints a slice of TypedNames.
// Names are grouped and printed with all other names
// of the same type, in alphebetical order on the type
// names. Untyped entries are printed in a group after all
// typed entries.  The elements of each group are printed
// in alphebetical order.  The prefix is printed before each
// group.
func printTypedNames(w io.Writer, prefix string, ns []TypedName) {
	byType := make(map[string][]string)
	for _, n := range ns {
		tstr := typeString(n.Types)
		byType[tstr] = append(byType[tstr], n.Str)
	}
	noType := byType[""]
	delete(byType, "")

	var decls declGroups
	for typ, names := range byType {
		decls = append(decls, declGroup{ typ, names })
	}
	sort.Sort(decls)

	for _, d := range decls {
		sort.Strings(d.ents)
		for i, e := range d.ents {
			if i == 0 {
				fmt.Fprint(w, prefix+e)
			} else {
				fmt.Fprint(w, " "+e)
			}
		}
		fmt.Fprint(w, " - ", d.typ)
		// The first entry can be prefixed by the empty
		// string, but subsequent entries need at least
		// a space.
		if prefix == "" {
			prefix = " "
		}
	}

	sort.Strings(noType)
	for i, n := range noType {
		if i == 0 {
			fmt.Fprint(w, prefix+n)
		} else {
			fmt.Fprint(w, " "+n)
		}
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
