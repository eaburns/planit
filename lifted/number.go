package lifted

// Assign numbers to all things of type Name

import (
	"fmt"
	"os"
)

func (d *Domain) AssignNums(s *Symtab) os.Error {
	if err := d.numberTypes(s); err != nil {
		return err
	}
	s.typeObjs = make([][]int, len(s.typeNames))
	if err := numberConsts(s, d.Constants); err != nil {
		return err
	}
	if err := d.numberPreds(s); err != nil {
		return err
	}
	for _, a := range d.Actions {
		if err := a.assignNums(s); err != nil {
			return err
		}
	}
	return nil
}

func (p *Problem) AssignNums(s *Symtab) os.Error {
	numberConsts(s, p.Objects)
	for _, init := range p.Init {
		if err := init.assignNums(s, nil); err != nil {
			return err
		}
	}
	return p.Goal.assignNums(s, nil)
}

func (d *Domain) numberTypes(s *Symtab) os.Error {
	for i, _ := range d.Types {
		d.Types[i].Name.numberType(s)
	}
	for i, _ := range d.Types {
		t := d.Types[i]
		for j, _ := range t.Type {
			if found := t.Type[j].numberType(s); !found {
				return undeclType(&t.Type[j])
			}
		}
	}
	return nil
}

func (d *Domain) numberPreds(s *Symtab) os.Error {
	for i, _ := range d.Predicates {
		if err := d.Predicates[i].assignNums(s); err != nil {
			return err
		}
	}
	return nil
}

func (p *Predicate) assignNums(s *Symtab) os.Error {
	p.Name.numberPred(s)
	for i, _ := range p.Parameters {
		parm := p.Parameters[i]
		for j, _ := range parm.Type {
			if found := parm.Type[j].numberType(s); !found {
				return undeclType(&parm.Type[j])
			}
		}
	}
	return nil
}

func numberConsts(s *Symtab, consts []TypedName) os.Error {
	for i, _ := range consts {
		if err := consts[i].numberConst(s); err != nil {
			return err
		}
	}
	return nil
}

func (c *TypedName) numberConst(s *Symtab) os.Error {
	first := c.Name.numberConst(s)
	cnum := c.Name.Num
	for i, _ := range c.Type {
		if found := c.Type[i].numberType(s); !found {
			return undeclType(&c.Type[i])
		}
		// If this is the 1st decl of this object
		// then add it to the table of all objects
		// of the given type
		if !first {
			tnum := c.Type[i].Num
			s.typeObjs[tnum] = append(s.typeObjs[tnum], cnum)
		}
	}
	return nil
}

func (a *Action) assignNums(s *Symtab) os.Error {
	var f *numFrame
	for i, _ := range a.Parameters {
		p := &a.Parameters[i]
		f = p.Name.numberVar(s, f)
		for j, _ := range p.Type {
			if found := p.Type[j].numberType(s); !found {
				return undeclType(&p.Type[j])
			}
		}
	}
	if err := a.Precondition.assignNums(s, f); err != nil {
		return err
	}
	return a.Effect.assignNums(s, f)
}

func (l *Literal) assignNums(s *Symtab, f *numFrame) os.Error {
	for i, t := range l.Parameters {
		name := &l.Parameters[i].Name
		switch t.Kind {
		case TermVariable:
			if fnxt := name.numberVar(s, f); fnxt != f {
				return undeclVar(name)
			}
		case TermConstant:
			if found := name.numberConst(s); !found {
				return undeclConst(name)
			}
		}
	}
	if found := l.Name.numberPred(s); !found {
		return undeclPred(&l.Name)
	}
	return nil
}

func (e *ExprBinary) assignNums(s *Symtab, f *numFrame) os.Error {
	if err := e.Left.assignNums(s, f); err != nil {
		return err
	}
	return e.Right.assignNums(s, f)
}

func (ExprTrue) assignNums(*Symtab, *numFrame) os.Error {
	return nil
}

func (ExprFalse) assignNums(*Symtab, *numFrame) os.Error {
	return nil
}

func (e *ExprAnd) assignNums(s *Symtab, f *numFrame) os.Error {
	return (*ExprBinary)(e).assignNums(s, f)
}

func (e *ExprOr) assignNums(s *Symtab, f *numFrame) os.Error {
	return (*ExprBinary)(e).assignNums(s, f)
}

func (e *ExprNot) assignNums(s *Symtab, f *numFrame) os.Error {
	return e.Expr.assignNums(s, f)
}

func (e *ExprQuant) assignNums(s *Symtab, f *numFrame) os.Error {
	f = e.Variable.Name.numberVar(s, f)
	for i, _ := range e.Variable.Type {
		if found := e.Variable.Type[i].numberType(s); !found {
			return undeclType(&e.Variable.Type[i])
		}
	}
	return e.Expr.assignNums(s, f)
}

func (e *ExprForall) assignNums(s *Symtab, f *numFrame) os.Error {
	return (*ExprQuant)(e).assignNums(s, f)
}

func (e *ExprExists) assignNums(s *Symtab, f *numFrame) os.Error {
	return (*ExprQuant)(e).assignNums(s, f)
}

func (e *ExprLiteral) assignNums(s *Symtab, f *numFrame) os.Error {
	return (*Literal)(e).assignNums(s, f)
}

func (e *EffectUnary) assignNums(s *Symtab, f *numFrame) os.Error {
	return e.Effect.assignNums(s, f)
}

func (EffectNone) assignNums(*Symtab, *numFrame) os.Error {
	return nil
}

func (e *EffectAnd) assignNums(s *Symtab, f *numFrame) os.Error {
	if err := e.Left.assignNums(s, f); err != nil {
		return err
	}
	return e.Right.assignNums(s, f)
}

func (e *EffectForall) assignNums(s *Symtab, f *numFrame) os.Error {
	f = e.Variable.Name.numberVar(s, f)
	for i, _ := range e.Variable.Type {
		if found := e.Variable.Type[i].numberType(s); !found {
			return undeclType(&e.Variable.Type[i])
		}
	}
	return e.Effect.assignNums(s, f)
}

func (e *EffectWhen) assignNums(s *Symtab, f *numFrame) os.Error {
	if err := e.Condition.assignNums(s, f); err != nil {
		return err
	}
	return e.Effect.assignNums(s, f)
}

func (e *EffectLiteral) assignNums(s *Symtab, f *numFrame) os.Error {
	return (*Literal)(e).assignNums(s, f)
}

func (e *EffectAssign) assignNums(*Symtab, *numFrame) os.Error {
	return nil
}

func (i *InitLiteral) assignNums(s *Symtab, f *numFrame) os.Error {
	return (*Literal)(i).assignNums(s, f)
}

func (i *InitEq) assignNums(s *Symtab, f *numFrame) os.Error {
	return nil
}

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

func undeclType(n *Name) os.Error {
	return fmt.Errorf("%s: Undeclared type %s\n", n.Loc, n.Str)
}

func undeclConst(n *Name) os.Error {
	return fmt.Errorf("%s: Undeclared constant %s\n", n.Loc, n.Str)
}

func undeclPred(n *Name) os.Error {
	return fmt.Errorf("%s: Undeclared predicate %s\n", n.Loc, n.Str)
}

func undeclVar(n *Name) os.Error {
	return fmt.Errorf("%s: Unbound variable %s\n", n.Loc, n.Str)
}