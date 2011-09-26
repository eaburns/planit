package lifted

// Assign numbers to all things of type Name

import "log"

func (d *Domain) AssignNums(s *Symtab) {
	d.numberTypes(s)
	s.typeObjs = make([][]int, len(s.typeNames))
	for i, _ := range d.Constants {
		d.Constants[i].numberConst(s)
	}
	for i, _ := range d.Predicates {
		d.Predicates[i].assignNums(s)
	}
	s.predInertia = make([]byte, len(s.predNames))
	for i, _ := range s.predInertia {
		s.predInertia[i] = posInertia | negInertia
	}
	for _, a := range d.Actions {
		a.assignNums(s)
	}
}

func (p *Problem) AssignNums(s *Symtab) {
	for i, _ := range p.Objects {
		p.Objects[i].numberConst(s)
	}
	for _, init := range p.Init {
		init.assignNums(s, nil)
	}
	p.Goal.assignNums(s, nil)
}

func (d *Domain) numberTypes(s *Symtab) {
	for i, _ := range d.Types {
		d.Types[i].Name.numberType(s)
	}
	for i, _ := range d.Types {
		t := d.Types[i]
		for j, _ := range t.Type {
			if found := t.Type[j].numberType(s); !found {
				undeclType(&t.Type[j])
			}
		}
	}
}

func (p *Predicate) assignNums(s *Symtab) {
	p.Name.numberPred(s)
	for i, _ := range p.Parameters {
		parm := p.Parameters[i]
		for j, _ := range parm.Type {
			if found := parm.Type[j].numberType(s); !found {
				undeclType(&parm.Type[j])
			}
		}
	}
}

func (c *TypedName) numberConst(s *Symtab) {
	first := c.Name.numberConst(s)
	cnum := c.Name.Num
	for i, _ := range c.Type {
		if found := c.Type[i].numberType(s); !found {
			undeclType(&c.Type[i])
		}
		// If this is the 1st decl of this object
		// then add it to the table of all objects
		// of the given type
		if !first {
			tnum := c.Type[i].Num
			s.typeObjs[tnum] = append(s.typeObjs[tnum], cnum)
		}
	}
}

func (a *Action) assignNums(s *Symtab) {
	var f *numFrame
	for i, _ := range a.Parameters {
		p := &a.Parameters[i]
		f = p.Name.numberVar(s, f)
		for j, _ := range p.Type {
			if found := p.Type[j].numberType(s); !found {
				undeclType(&p.Type[j])
			}
		}
	}
	a.Precondition.assignNums(s, f)
	a.Effect.assignNums(s, f)
}

func (l *LiteralNode) assignNums(s *Symtab, f *numFrame) {
	for i, t := range l.Parameters {
		name := &l.Parameters[i].Name
		switch t.Kind {
		case TermVariable:
			if fnxt := name.numberVar(s, f); fnxt != f {
				undeclVar(name)
			}
		case TermConstant:
			if found := name.numberConst(s); !found {
				undeclConst(name)
			}
		}
	}
	if found := l.Name.numberPred(s); !found {
		undeclPred(&l.Name)
	}
}

func (e *BinaryNode) assignNums(s *Symtab, f *numFrame) {
	e.Left.assignNums(s, f)
	e.Right.assignNums(s, f)
}

func (e *UnaryNode) assignNums(s *Symtab, f *numFrame) {
	e.Formula.assignNums(s, f)
}

func (TrueNode) assignNums(*Symtab, *numFrame) { }

func (FalseNode) assignNums(*Symtab, *numFrame) { }

func (e *QuantNode) assignNums(s *Symtab, f *numFrame) {
	f = e.Variable.Name.numberVar(s, f)
	for i, _ := range e.Variable.Type {
		if found := e.Variable.Type[i].numberType(s); !found {
			undeclType(&e.Variable.Type[i])
		}
	}
	e.Formula.assignNums(s, f)
}

func (e *EffectLiteralNode) assignNums(s *Symtab, f *numFrame) {
	e.LiteralNode.assignNums(s, f)
	switch e.Positive {
	case e.Positive:
		s.predInertia[e.Name.Num] &^= posInertia
	case !e.Positive:
		s.predInertia[e.Name.Num] &^= negInertia
	}
}

func (e *WhenNode) assignNums(s *Symtab, f *numFrame) {
	e.Condition.assignNums(s, f)
	e.Formula.assignNums(s, f)
}

func (e *AssignNode) assignNums(*Symtab, *numFrame) { }

func (name *Name) numberType(s *Symtab) bool {
	if n, ok := s.typeNums[name.Str]; ok {
		name.Num = n
		return true
	}
	name.Num = len(s.typeNames)
	s.typeNums[name.Str] = name.Num
	s.typeNames = append(s.typeNames, name.Str)
	return false
}

func (name *Name) numberConst(s *Symtab) bool {
	if n, ok := s.constNums[name.Str]; ok {
		name.Num = n
		return true
	}
	name.Num = len(s.constNames)
	s.constNums[name.Str] = name.Num
	s.constNames = append(s.constNames, name.Str)
	return false
}

func (name *Name) numberPred(s *Symtab) bool {
	if n, ok := s.predNums[name.Str]; ok {
		name.Num = n
		return true
	}
	name.Num = len(s.predNames)
	s.predNums[name.Str] = name.Num
	s.predNames = append(s.predNames, name.Str)
	return false
}

func (name *Name) numberVar(s *Symtab, f *numFrame) *numFrame {
	if n, ok := f.lookup(name.Str); ok {
		name.Num = n
		return f
	}
	n := len(s.varNames)
	name.Num = n
	s.varNames = append(s.varNames, name.Str)
	return &numFrame{name: name.Str, num: n, up: f}
}

type numFrame struct {
	name string
	num  int
	up   *numFrame
}

func (f *numFrame) lookup(name string) (int, bool) {
	if f == nil {
		return 0, false
	}
	if f.name == name {
		return f.num, true
	}
	return f.up.lookup(name)
}

func undeclType(n *Name) {
	log.Fatalf("%s: Undeclared type %s\n", n.Loc, n.Str)
}

func undeclConst(n *Name) {
	log.Fatalf("%s: Undeclared constant %s\n", n.Loc, n.Str)
}

func undeclPred(n *Name) {
	log.Fatalf("%s: Undeclared predicate %s\n", n.Loc, n.Str)
}

func undeclVar(n *Name) {
	log.Fatalf("%s: Unbound variable %s\n", n.Loc, n.Str)
}
