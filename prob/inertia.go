package prob

func (d *Domain) findInertia(s *symtab) {
	s.inertia = make([]byte, len(s.predNames))
	for i := range s.inertia {
		s.inertia[i] = posInertia | negInertia
	}
	for i := range d.Actions {
		d.Actions[i].Effect.findInertia(s)
	}
	return
}

func (TrueNode) findInertia(*symtab) {}

func (FalseNode) findInertia(*symtab) {}

func (*LeafNode) findInertia(*symtab) {}

func (f *UnaryNode) findInertia(s *symtab) {
	f.Formula.findInertia(s)
}

func (f *BinaryNode) findInertia(s *symtab) {
	f.Left.findInertia(s)
	f.Right.findInertia(s)
}

func (e *Literal) findInertia(s *symtab) {
	switch e.Positive {
	case e.Positive:
		s.inertia[e.Num] &^= posInertia
	case !e.Positive:
		s.inertia[e.Num] &^= negInertia
	}
}
