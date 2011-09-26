package lifted

func (n *LiteralNode) dnf() Formula { return n }

func (TrueNode) dnf() Formula { return TrueNode(1) }

func (FalseNode) dnf() Formula { return FalseNode(1) }

func (n *AndNode) dnf() Formula {
	return nil
}

func (n *OrNode) dnf() Formula {
	return nil
}

func (n *NotNode) dnf() Formula {
	return Negate(n.Formula.dnf())
}

func (n *ForallNode) dnf() Formula {
	panic("ForallNode in the tree when converting to DFN")
}

func (n *ExistsNode) dnf() Formula {
	panic("ExistsNode in the tree when converting to DFN")
}

func (n *WhenNode) dnf() Formula {
	disj := Formula(FalseNode(0))
	conds := collectDisj(n.Condition.dnf())
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

func collectDisj(f Formula) (fs []Formula) {
	switch n := f.(type) {
	case *OrNode:
		fs = append(fs, collectDisj(n.Left)...)
		fs = append(fs, collectDisj(n.Right)...)
	default:
		fs = append(fs, n)
	}
	return
}