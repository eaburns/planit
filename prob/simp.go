package prob

func Conjunct(l Formula, r Formula) Formula {
	switch l.(type) {
	case *trueNode:
		return r
	case *falseNode:
		return MakeFalse()
	}
	switch r.(type) {
	case *trueNode:
		return l
	case *falseNode:
		return MakeFalse()
	}
	return &AndNode{BinaryNode{Left: l, Right: r}}
}

func Disjunct(l Formula, r Formula) Formula {
	switch l.(type) {
	case *trueNode:
		return MakeTrue()
	case *falseNode:
		return r
	}
	switch r.(type) {
	case *trueNode:
		return MakeTrue()
	case *falseNode:
		return l
	}
	return &OrNode{BinaryNode{Left: l, Right: r}}
}

func Negate(e Formula) Formula {
	switch n := e.(type) {
	case *trueNode:
		return MakeFalse()
	case *falseNode:
		return MakeTrue()
	case *NotNode:
		return n.Formula
	case *LiteralNode:
		n.Positive = !n.Positive
		return n
	}
	return &NotNode{UnaryNode{Formula: e}}
}
