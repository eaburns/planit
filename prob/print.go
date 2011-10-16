package prob

import (
	"fmt"
	"bytes"
)

func (d *Domain) String() string {
	var buf bytes.Buffer

	buf.WriteString("&Domain{")
	fmt.Fprintf(&buf, "Name:%s,\n", d.Name)
	fmt.Fprintf(&buf, "Requirements:%+v\n", d.Requirements)
	fmt.Fprintf(&buf, "Types:%v\n", d.Types)
	fmt.Fprintf(&buf, "Constants:%v\n", d.Constants)
	fmt.Fprintf(&buf, "Predicates:%+v\n", d.Predicates)

	buf.WriteString("Actions:[\n")
	for i, a := range d.Actions {
		buf.WriteString(a.String())
		if i < len(d.Actions)-1 {
			buf.WriteString("\n\n")
		}
	}
	buf.WriteString("],\n}")

	return buf.String()
}

func (p Predicate) String() string {
	return fmt.Sprintf("Predicate{Name:%v, Parameters:%v}",
		p.Name, p.Parameters)
}

func (a Action) String() string {
	var buf bytes.Buffer

	buf.WriteString("Action{")
	fmt.Fprintf(&buf, "Name:%s", a.Name)
	fmt.Fprintf(&buf, "\nParameters:%+v", a.Parameters)
	fmt.Fprintf(&buf, "\nPrecondition:%+v", a.Precondition)
	fmt.Fprintf(&buf, "\nEffect:%+v", a.Effect)
	buf.WriteString("}")

	return buf.String()
}

func (p *Problem) String() string {
	var buf bytes.Buffer

	buf.WriteString("Problem{")
	fmt.Fprintf(&buf, "Name:%s\n", p.Name)
	fmt.Fprintf(&buf, "Domain:%s\n", p.Domain)
	fmt.Fprintf(&buf, "Requirements:%v\n", p.Requirements)
	fmt.Fprintf(&buf, "Objects:%v\n", p.Objects)
	fmt.Fprintf(&buf, "Init:%+v\n", p.Init)
	fmt.Fprintf(&buf, "Goal:%+v\n", p.Goal)
	fmt.Fprintf(&buf, "Metric:%+v\n", p.Metric)
	buf.WriteByte('}')

	return buf.String()
}

func (n Name) String() string {
	if n.Num < 0 {
		return fmt.Sprintf("{%s}", n.Str)
	}
	return fmt.Sprintf("{%s, %d}", n.Str, n.Num)
}

func (lit *Literal) String() string {
	return fmt.Sprintf("Literal{Predicate:%v, Positive:%t, Parameters:%v}",
		lit.Predicate, lit.Positive, lit.Parameters)
}

func (e *BinaryNode) String() string {
	return fmt.Sprintf("Left:%v, Right:%v", e.Left, e.Right)
}

func (e *UnaryNode) String() string {
	return fmt.Sprintf("Formula:%v", e.Formula)
}

func (e *QuantNode) String() string {
	return fmt.Sprintf("Variable:%v, %v", e.Variable, e.UnaryNode)
}

func (TrueNode) String() string {
	return "TrueNode"
}

func (FalseNode) String() string {
	return "FalseNode"
}

func (e *AndNode) String() string {
	return fmt.Sprintf("AndNode{%v}", e.BinaryNode)
}

func (e *OrNode) String() string {
	return fmt.Sprintf("OrNode{%v}", e.BinaryNode)
}

func (e *NotNode) String() string {
	return fmt.Sprintf("NotNode{%v}", e.UnaryNode)
}

func (e *ForallNode) String() string {
	return fmt.Sprintf("ForallNode{%v}", e.QuantNode)
}

func (e *ExistsNode) String() string {
	return fmt.Sprintf("ExistsNode[{%v}", e.QuantNode)
}

func (eff *WhenNode) String() string {
	return fmt.Sprintf("WhenNode{Condition:%v, }", eff.Condition,
		eff.UnaryNode)
}

func (eff *AssignNode) String() string {
	return fmt.Sprintf("AssignNode{Op:%v, Lval:%v, Rval:%v}",
		eff.Op, eff.Lval, eff.Rval)
}

var assignOpNames = map[AssignOp]string{
	OpAssign:    "OpAssign",
	OpScaleUp:   "OpScaleUp",
	OpScaleDown: "OpScaleDown",
	OpIncrease:  "OpIncrease",
	OpDecrease:  "OpDecrease",
}