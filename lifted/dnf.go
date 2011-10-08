package lifted

func (n *LiteralNode) dnf() Formula { return n }

func (TrueNode) dnf() Formula { return TrueNode{} }

func (FalseNode) dnf() Formula { return FalseNode{} }

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

func (n *ForallNode) dnf() Formula {
	panic("ForallNode in the tree when converting to DFN")
}

func (n *ExistsNode) dnf() Formula {
	panic("ExistsNode in the tree when converting to DFN")
}

func (n *WhenNode) dnf() Formula {
	disj := Formula(FalseNode{})
	conds := collectOrs(n.Condition.dnf())
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

func (n *AssignNode) dnf() Formula { return n }

func collectOrs(f Formula) (fs []Formula) {
	switch n := f.(type) {
	case *OrNode:
		fs = append(fs, collectOrs(n.Left)...)
		fs = append(fs, collectOrs(n.Right)...)
	default:
		fs = append(fs, n)
	}
	return
}
