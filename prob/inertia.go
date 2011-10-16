package prob

func (*LeafNode) findInertia(*Symtab) {}

func (f *UnaryNode) findInertia(s *Symtab) {
	f.Formula.findInertia(s)
}

func (f *BinaryNode) findInertia(s *Symtab) {
	f.Left.findInertia(s)
	f.Right.findInertia(s)
}

func (d *Domain) FindInertia(s *Symtab) {
	s.predInertia = make([]byte, len(s.predNames))
	for i := range s.predInertia {
		s.predInertia[i] = posInertia | negInertia
	}
	for i := range d.Actions {
		d.Actions[i].Effect.findInertia(s)
	}
}

func (e *LiteralNode) findInertia(s *Symtab) {
	switch e.Positive {
	case e.Positive:
		s.predInertia[e.Num] &^= posInertia
	case !e.Positive:
		s.predInertia[e.Num] &^= negInertia
	}
}
