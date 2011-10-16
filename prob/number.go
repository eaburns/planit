package prob

// Assign numbers to all things of type Name

import "log"

func (d *Domain) assignNums(s *symtab) {
	d.numberTypes(s)
	s.typeObjs = make([][]int, len(s.typeNames))
	for i := range d.Constants {
		d.Constants[i].numberConst(s)
	}
	for i := range d.Predicates {
		d.Predicates[i].assignNums(s)
	}
	for _, a := range d.Actions {
		a.assignNums(s)
	}
}

func (p *Problem) assignNums(s *symtab) {
	for i := range p.Objects {
		p.Objects[i].numberConst(s)
	}
	for _, init := range p.Init {
		init.assignNums(s, nil)
	}
	p.Goal.assignNums(s, nil)
}

func (d *Domain) numberTypes(s *symtab) {
	for i := range d.Types {
		d.Types[i].numberType(s)
	}
	for i := range d.Types {
		t := d.Types[i]
		for j := range t.Types {
			if found := t.Types[j].numberType(s); !found {
				undeclType(&t.Types[j])
			}
		}
	}
}

func (p *Predicate) assignNums(s *symtab) {
	p.numberPred(s)
	for i := range p.Parameters {
		parm := p.Parameters[i]
		for j := range parm.Types {
			if found := parm.Types[j].numberType(s); !found {
				undeclType(&parm.Types[j])
			}
		}
	}
}

func (c *TypedName) numberConst(s *symtab) {
	seen := c.Name.numberConst(s)
	cnum := c.Num
	for i := range c.Types {
		if found := c.Types[i].numberType(s); !found {
			undeclType(&c.Types[i])
		}
		// If this is the 1st decl of this object
		// then add it to the table of all objects
		// of the given type
		if !seen {
			tnum := c.Types[i].Num
			s.typeObjs[tnum] = append(s.typeObjs[tnum], cnum)
		}
	}
}

func (a *Action) assignNums(s *symtab) {
	var f *numFrame
	for i := range a.Parameters {
		p := &a.Parameters[i]
		f = p.numberVar(s, f)
		for j := range p.Types {
			if found := p.Types[j].numberType(s); !found {
				undeclType(&p.Types[j])
			}
		}
	}
	a.Precondition.assignNums(s, f)
	a.Effect.assignNums(s, f)
}

func (l *Literal) assignNums(s *symtab, f *numFrame) {
	for i := range l.Parameters {
		switch term := l.Parameters[i].(type) {
		case *Variable:
			if fnxt := term.numberVar(s, f); fnxt != f {
				undeclVar(&term.Name)
			}
		case *Constant:
			if found := term.numberConst(s); !found {
				undeclConst(&term.Name)
			}
		}
	}
	if found := l.Predicate.numberPred(s); !found {
		undeclPred(&l.Predicate)
	}
}

func (TrueNode) assignNums(*symtab, *numFrame) {}

func (FalseNode) assignNums(*symtab, *numFrame) {}

func (*LeafNode) assignNums(*symtab, *numFrame) {}

func (e *UnaryNode) assignNums(s *symtab, f *numFrame) {
	e.Formula.assignNums(s, f)
}

func (e *BinaryNode) assignNums(s *symtab, f *numFrame) {
	e.Left.assignNums(s, f)
	e.Right.assignNums(s, f)
}

func (e *QuantNode) assignNums(s *symtab, f *numFrame) {
	f = e.Variable.numberVar(s, f)
	for i := range e.Variable.Types {
		if found := e.Variable.Types[i].numberType(s); !found {
			undeclType(&e.Variable.Types[i])
		}
	}
	e.Formula.assignNums(s, f)
}

func (e *WhenNode) assignNums(s *symtab, f *numFrame) {
	e.Condition.assignNums(s, f)
	e.Formula.assignNums(s, f)
}

func (name *Name) numberType(s *symtab) bool {
	if n, ok := s.typeNums[name.Str]; ok {
		name.Num = n
		return true
	}
	name.Num = len(s.typeNames)
	s.typeNums[name.Str] = name.Num
	s.typeNames = append(s.typeNames, name.Str)
	return false
}

func (name *Name) numberConst(s *symtab) bool {
	if n, ok := s.constNums[name.Str]; ok {
		name.Num = n
		return true
	}
	name.Num = len(s.constNames)
	s.constNums[name.Str] = name.Num
	s.constNames = append(s.constNames, name.Str)
	return false
}

func (name *Name) numberPred(s *symtab) bool {
	if n, ok := s.predNums[name.Str]; ok {
		name.Num = n
		return true
	}
	name.Num = len(s.predNames)
	s.predNums[name.Str] = name.Num
	s.predNames = append(s.predNames, name.Str)
	return false
}

func (name *Name) numberVar(s *symtab, f *numFrame) *numFrame {
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
