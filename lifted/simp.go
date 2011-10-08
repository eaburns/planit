package lifted

func Conjunct(l Formula, r Formula) Formula {
	switch l.(type) {
	case TrueNode:
		return r
	case FalseNode:
		return FalseNode{}
	}
	switch r.(type) {
	case TrueNode:
		return l
	case FalseNode:
		return FalseNode{}
	}
	return &AndNode{BinaryNode{Left: l, Right: r}}
}

func Disjunct(l Formula, r Formula) Formula {
	switch l.(type) {
	case TrueNode:
		return TrueNode{}
	case FalseNode:
		return r
	}
	switch r.(type) {
	case TrueNode:
		return TrueNode{}
	case FalseNode:
		return l
	}
	return &OrNode{BinaryNode{Left: l, Right: r}}
}

func Negate(e Formula) Formula {
	switch n := e.(type) {
	case TrueNode:
		return FalseNode{}
	case FalseNode:
		return TrueNode{}
	case *NotNode:
		return n.Formula
	case *LiteralNode:
		n.Positive = !n.Positive
		return n
	}
	return &NotNode{UnaryNode{Formula: e}}
}