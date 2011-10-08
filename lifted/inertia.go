package lifted

func (d *Domain) FindInertia(s *Symtab) {
	for i := range d.Actions {
		d.Actions[i].Effect.findInertia(s)
	}
}

func (e *LiteralNode) findInertia(s *Symtab) {
	switch e.Positive {
	case e.Positive:
		s.predInertia[e.Name.Num] &^= posInertia
	case !e.Positive:
		s.predInertia[e.Name.Num] &^= negInertia
	}
}

func (TrueNode) findInertia(*Symtab) {}

func (FalseNode) findInertia(*Symtab) {}

func (f *BinaryNode) findInertia(s *Symtab) {
	f.Left.findInertia(s)
	f.Right.findInertia(s)
}

func (f *UnaryNode) findInertia(s *Symtab) {
	f.Formula.findInertia(s)
}

func (*AssignNode) findInertia(*Symtab) {}
