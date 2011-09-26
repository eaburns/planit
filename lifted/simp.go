package lifted

func Conjunct(l Formula, r Formula) Formula {
	switch l.(type) {
	case TrueNode:
		return r
	case FalseNode:
		return FalseNode(0)
	}
	switch r.(type) {
	case TrueNode:
		return l
	case FalseNode:
		return FalseNode(0)
	}
	return &AndNode{BinaryNode{Left: l, Right: r}}
}

func Disjunct(l Formula, r Formula) Formula {
	switch l.(type) {
	case TrueNode:
		return TrueNode(1)
	case FalseNode:
		return r
	}
	switch r.(type) {
	case TrueNode:
		return TrueNode(1)
	case FalseNode:
		return l
	}
	return &OrNode{BinaryNode{Left: l, Right: r}}
}

func Negate(e Formula) Formula {
	switch neg := e.(type) {
	case *NotNode:
		return neg.Formula
	case *LiteralNode:
		neg.Positive = !neg.Positive
		return neg
	}
	return &NotNode{UnaryNode{Formula: e}}
}