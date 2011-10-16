package prob

func (a *Action) dnf() {
	a.Precondition = a.Precondition.dnf()
	a.Effect = a.Effect.dnf()
}

func (n TrueNode) dnf() Formula { return n }

func (n FalseNode) dnf() Formula { return n }

func (l *Literal) dnf() Formula { return l }

func (a *AssignNode) dnf() Formula { return a }

func (n *AndNode) dnf() Formula {
	n.Left = n.Left.dnf()
	n.Right = n.Right.dnf()

	var disj *OrNode
	var other Formula
	if l, ok := n.Left.(*OrNode); ok {
		disj = l
		other = n.Right
	} else if r, ok := n.Right.(*OrNode); ok {
		disj = r
		other = n.Left
	}

	if disj == nil {
		return n
	}

	left := Conjunct(other, disj.Left)
	right := Conjunct(other, disj.Right)
	return Disjunct(left.dnf(), right.dnf()).dnf()
}

func (n *OrNode) dnf() Formula {
	n.Left = n.Left.dnf()
	n.Right = n.Right.dnf()
	return n
}

func (n *NotNode) dnf() Formula {
	switch f := n.Formula.dnf().(type) {
	case *OrNode:
		return Conjunct(Negate(f.Left), Negate(f.Right)).dnf()
	case *AndNode:
		return Disjunct(Negate(f.Left), Negate(f.Right)).dnf()
	}
	return Negate(n.Formula)
}

func (n *QuantNode) dnf() Formula {
	panic("QuantNode in the tree when converting to DFN")
}

func (n *WhenNode) dnf() Formula {
	n.Formula = n.Formula.dnf()

	disj := Formula(FalseNode{})
	conds := gatherOrs(n.Condition.dnf())
	if len(conds) == 1 {
		return n
	}
	for _, cond := range conds {
		nd := &WhenNode{
			Condition: cond,
			UnaryNode: n.UnaryNode,
		}
		disj = Disjunct(disj, Formula(nd))
	}
	return disj
}

func (a *Action) ensureDnf() {
	a.Precondition.ensureDnf()
	a.Effect.ensureDnf()
}

func (TrueNode) ensureDnf() { return }

func (FalseNode) ensureDnf() { return }

func (*LeafNode) ensureDnf() { return }

func (n *AndNode) ensureDnf() {
	if _, ok := n.Left.(*OrNode); ok {
		panic("An OrNode follows an AndNode in a 'DNF' formula")
	}
	n.Left.ensureDnf()
	if _, ok := n.Right.(*OrNode); ok {
		panic("OrNode found beneath an AndNode in a 'DNF' formula")
	}
	n.Right.ensureDnf()
}

func (*NotNode) ensureDnf() {
	panic("NotNode found in a 'DNF' formula")
}

func (n *QuantNode) ensureDnf() {
	panic("QuantNode in the tree when converting to DFN")
}

func (n *OrNode) ensureDnf() {
	n.Left.ensureDnf()
	n.Right.ensureDnf()
}

func (n *WhenNode) ensureDnf() {
	if _, ok := n.Condition.(*OrNode); ok {
		panic("OrNode found beneath a WhenNode in a 'DNF' formula")
	}
	n.Condition.ensureDnf()
	if _, ok := n.Formula.(*OrNode); ok {
		panic("OrNode found beneath a WhenNode in a 'DNF' formula")
	}
	n.Formula.ensureDnf()
}

func gatherOrs(f Formula) (fs []Formula) {
	switch n := f.(type) {
	case *OrNode:
		fs = append(fs, gatherOrs(n.Left)...)
		fs = append(fs, gatherOrs(n.Right)...)
	default:
		fs = append(fs, f)
	}
	return
}