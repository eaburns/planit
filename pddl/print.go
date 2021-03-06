// © 2013 the PlanIt Authors under the MIT license. See AUTHORS for the list of authors.

package pddl

import (
	"fmt"
	"io"
)

// PrintDomain prints the domain in valid PDDL to a writer.
func PrintDomain(w io.Writer, d *Domain) {
	fmt.Fprintf(w, "(define (domain %s)\n", d.Name)
	printReqsDef(w, d.Requirements)
	printTypesDef(w, d.Types)
	printConstsDef(w, ":constants", d.Constants)
	printPredsDef(w, d.Predicates)
	printFuncsDef(w, d.Functions)
	for _, act := range d.Actions {
		printAction(w, act)
	}
	fmt.Fprintln(w, ")")
}

func printReqsDef(w io.Writer, reqs []Name) {
	if len(reqs) == 0 {
		return
	}
	fmt.Fprintf(w, "%s(:requirements\n", indent(1))
	for i, r := range reqs {
		s := r.Str
		if i == len(reqs)-1 {
			s += ")"
		}
		fmt.Fprintln(w, indent(2), s)
	}
}

func printTypesDef(w io.Writer, ts []Type) {
	if len(ts) == 0 {
		return
	}
	fmt.Fprintf(w, "%s(:types", indent(1))
	var ids []TypedEntry
	for _, t := range ts {
		if t.Location.Line == 0 {
			// Skip undeclared implicit types like object.
			continue
		}
		ids = append(ids, t.TypedEntry)
	}
	printTypedNames(w, "\n"+indent(2), ids)
	fmt.Fprintln(w, ")")
}

// PrintConstsDef prints a constant definition with the given definition name
// (should be either :constants or :objects).
func printConstsDef(w io.Writer, def string, cs []TypedEntry) {
	if len(cs) == 0 {
		return
	}
	fmt.Fprintf(w, "%s(%s", indent(1), def)
	printTypedNames(w, "\n"+indent(2), cs)
	fmt.Fprintln(w, ")")
}

func printPredsDef(w io.Writer, ps []Predicate) {
	if len(ps) == 0 {
		return
	}
	fmt.Fprintf(w, "%s(:predicates\n", indent(1))
	for i, p := range ps {
		if p.Location.Line == 0 {
			// Skip undefined implicit predicates like =.
			continue
		}
		fmt.Fprintf(w, "%s(%s", indent(2), p.Str)
		printTypedNames(w, " ", p.Parameters)
		fmt.Fprint(w, ")")
		if i < len(ps)-1 {
			fmt.Fprint(w, "\n")
		}
	}
	fmt.Fprintln(w, ")")
}

func printFuncsDef(w io.Writer, fs []Function) {
	if len(fs) == 0 {
		return
	}
	fmt.Fprintf(w, "%s(:functions\n", indent(1))
	for i, f := range fs {
		fmt.Fprintf(w, "%s(%s", indent(2), f.Str)
		printTypedNames(w, " ", f.Parameters)
		fmt.Fprint(w, ")")
		if len(f.Types) > 0 {
			fmt.Fprint(w, " - ", typeString(f.Types))
		}
		if i < len(fs)-1 {
			fmt.Fprint(w, "\n")
		}
	}
	fmt.Fprintln(w, ")")
}

func printAction(w io.Writer, act Action) {
	fmt.Fprintf(w, "%s(:action %s\n", indent(1), act.Name)
	fmt.Fprintf(w, "%s:parameters (", indent(2))
	printTypedNames(w, "", act.Parameters)
	fmt.Fprint(w, ")")
	if act.Precondition != nil {
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, "%s:precondition\n", indent(2))
		act.Precondition.print(w, indent(3))
	}
	if act.Effect != nil {
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, "%s:effect\n", indent(2))
		act.Effect.print(w, indent(3))
	}
	fmt.Fprintln(w, ")")
}

// PrintProblem prints the problem in valid PDDL to the given writer.
func PrintProblem(w io.Writer, p *Problem) {
	fmt.Fprintf(w, "(define (problem %s)\n%s(:domain %s)\n",
		p.Name, indent(1), p.Domain)
	printReqsDef(w, p.Requirements)
	printConstsDef(w, ":objects", p.Objects)

	fmt.Fprintf(w, "%s(:init", indent(1))
	for _, f := range p.Init {
		fmt.Fprint(w, "\n")
		f.print(w, indent(2))
	}
	fmt.Fprint(w, ")\n")

	fmt.Fprintf(w, "%s(:goal\n", indent(1))
	p.Goal.print(w, indent(2))

	fmt.Fprintln(w, ")\n)")
}

// DeclGroup is a group of declarators along with their type.
type declGroup struct {
	typ  string
	ents []string
}

// DeclGroups implements sort.Interface, sorting the list of typed declarations by their type name.
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

// PrintTypedNames prints a slice of TypedNames. Adjacent items with the same type are
// all printed in a group.  Each group is preceeded by the prefix.
func printTypedNames(w io.Writer, prefix string, ns []TypedEntry) {
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
				panic(n.Location.String() + ": untyped declarations in the middle of a typed list")
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

// TypeString returns the string representation of a type.
func typeString(t []TypeName) (str string) {
	switch len(t) {
	case 0:
		break
	case 1:
		if t[0].Location.Line == 0 {
			// Use the empty string for undeclared
			// implicit types (such as object).
			break
		}
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

func (lit *LiteralNode) print(w io.Writer, prefix string) {
	if lit.Negative {
		fmt.Fprintf(w, "%s(not ", prefix)
		prefix = ""
	}
	fmt.Fprintf(w, "%s(", prefix)
	fmt.Fprint(w, lit.Predicate)
	for _, t := range lit.Arguments {
		fmt.Fprintf(w, " %s", t.Name)
	}
	fmt.Fprint(w, ")")
	if lit.Negative {
		fmt.Fprint(w, ")")
	}
}

func (n *AndNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(and", prefix)
	for _, f := range n.Formula {
		fmt.Fprint(w, "\n")
		f.print(w, prefix+indent(1))
	}
	fmt.Fprint(w, ")")
}

func (n *OrNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(or", prefix)
	for _, f := range n.Formula {
		fmt.Fprint(w, "\n")
		f.print(w, prefix+indent(1))
	}
	fmt.Fprint(w, ")")
}

func (n *NotNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(not\n", prefix)
	n.Formula.print(w, prefix+indent(1))
	fmt.Fprint(w, ")")
}

func (n *ImplyNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(imply\n", prefix)
	n.Left.print(w, prefix+indent(1))
	fmt.Fprint(w, "\n")
	n.Right.print(w, prefix+indent(1))
	fmt.Fprint(w, ")")
}

func (n *ForallNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(forall (", prefix)
	printTypedNames(w, "", n.Variables)
	fmt.Fprint(w, ")\n")
	n.Formula.print(w, prefix+indent(1))
	fmt.Fprint(w, ")")
}

func (n *ExistsNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(exists (", prefix)
	printTypedNames(w, "", n.Variables)
	fmt.Fprint(w, ")\n")
	n.Formula.print(w, prefix+indent(1))
	fmt.Fprint(w, ")")
}

func (n *WhenNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(when\n", prefix)
	n.Condition.print(w, prefix+indent(1))
	fmt.Fprint(w, "\n")
	n.Formula.print(w, prefix+indent(1))
	fmt.Fprint(w, ")")
}

func (n *AssignNode) print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s(%s ", prefix, n.Op)
	n.Lval.print(w)
	if n.IsNumber {
		fmt.Fprintf(w, " %s", n.Number)
	} else {
		fmt.Fprint(w, " ")
		n.Fhead.print(w)
	}
	fmt.Fprintf(w, ")")
}

func (h *Fhead) print(w io.Writer) {
	if len(h.Arguments) == 0 {
		fmt.Fprintf(w, "(%s)", h.Name)
		return
	}
	fmt.Fprintf(w, "(%s", h.Name)
	for _, t := range h.Arguments {
		fmt.Fprintf(w, " %s", t.Name)
	}
	fmt.Fprint(w, ")")
}

// Indent returns a string containing a given number of indentations.
func indent(n int) (s string) {
	for i := 0; i < n; i++ {
		s += "\t"
	}
	return
}
