package lifted

// Assign numbers to all things of type Name

import (
	"fmt"
	"os"
)

func (d *Domain) AssignNums(s *Symtab) os.Error {
	for i, _ := range d.Types {
		d.Types[i].Name.numberType(s)
	}
	for i, _ := range d.Types {
		t := d.Types[i]
		for j, _ := range t.Type {
			if found := t.Type[j].numberType(s); !found {
				return undefType(&t.Type[j])
			}
		}
	}

	s.typeObjs = make([][]int, len(s.typeNames))
	if err := numberConsts(s, d.Constants); err != nil {
		return err
	}

	for i, _ := range d.Predicates {
		p := &d.Predicates[i]
		p.Name.numberPred(s)
		for j, _ := range p.Parameters {
			parm := p.Parameters[j]
			for k, _ := range parm.Type {
				if found := parm.Type[k].numberType(s); !found {
					return undefType(&parm.Type[k])
				}
			}
		}
	}

	for _, a := range d.Actions {
		if err := a.AssignNums(s); err != nil {
			return err
		}
	}
	return nil
}

func numberConsts(s *Symtab, consts []TypedName) os.Error {
	for i, _ := range consts {
		c := &consts[i]
		first := c.Name.numberConst(s)
		cnum := c.Name.Num
		for j, _ := range c.Type {
			if found := c.Type[j].numberType(s); !found {
				return undefType(&c.Type[j])
			}
			// If this is the 1st decl of this object
			// then add it to the table of all objects
			// of the given type
			if !first {
				tnum := c.Type[j].Num
				s.typeObjs[tnum] = append(s.typeObjs[tnum], cnum)
			}
		}
	}
	return nil
}

func (a *Action) AssignNums(s *Symtab) os.Error {
	var f *numFrame
	for i, _ := range a.Parameters {
		p := &a.Parameters[i]
		f = p.Name.numberVar(s, f)
		for j, _ := range p.Type {
			if found := p.Type[j].numberType(s); !found {
				return undefType(&p.Type[j])
			}
		}
	}
	if err := a.Precondition.AssignNums(s, f); err != nil {
		return err
	}
	return a.Effect.AssignNums(s, f)
}

func (p *Problem) AssignNums(s *Symtab) os.Error {
	numberConsts(s, p.Objects)
	for _, init := range p.Init {
		if err := init.AssignNums(s, nil); err != nil {
			return err
		}
	}
	return p.Goal.AssignNums(s, nil)
}

func (l *Literal) AssignNums(s *Symtab, f *numFrame) os.Error {
	for i, t := range l.Parameters {
		name := &l.Parameters[i].Name
		switch t.Kind {
		case TermVariable:
			if fnxt := name.numberVar(s, f); fnxt == f {
				break
			}
			return fmt.Errorf("%s: Unbound variable %s\n", name.Loc, name.Str)
		case TermConstant:
			name.numberConst(s)
		}
	}
	l.Name.numberPred(s)
	return nil
}

func (e *ExprBinary) AssignNums(s *Symtab, f *numFrame) os.Error {
	if err := e.Left.AssignNums(s, f); err != nil {
		return err
	}
	return e.Right.AssignNums(s, f)
}

func (ExprTrue) AssignNums(*Symtab, *numFrame) os.Error {
	return nil
}

func (ExprFalse) AssignNums(*Symtab, *numFrame) os.Error {
	return nil
}

func (e *ExprAnd) AssignNums(s *Symtab, f *numFrame) os.Error {
	return (*ExprBinary)(e).AssignNums(s, f)
}

func (e *ExprOr) AssignNums(s *Symtab, f *numFrame) os.Error {
	return (*ExprBinary)(e).AssignNums(s, f)
}

func (e *ExprNot) AssignNums(s *Symtab, f *numFrame) os.Error {
	return e.Expr.AssignNums(s, f)
}

func (e *ExprQuant) AssignNums(s *Symtab, f *numFrame) os.Error {
	f = e.Variable.Name.numberVar(s, f)
	for i, _ := range e.Variable.Type {
		if found := e.Variable.Type[i].numberType(s); !found {
			return undefType(&e.Variable.Type[i])
		}
	}
	return e.Expr.AssignNums(s, f)
}

func (e *ExprForall) AssignNums(s *Symtab, f *numFrame) os.Error {
	return (*ExprQuant)(e).AssignNums(s, f)
}

func (e *ExprExists) AssignNums(s *Symtab, f *numFrame) os.Error {
	return (*ExprQuant)(e).AssignNums(s, f)
}

func (e *ExprLiteral) AssignNums(s *Symtab, f *numFrame) os.Error {
	return (*Literal)(e).AssignNums(s, f)
}

func (e *EffectUnary) AssignNums(s *Symtab, f *numFrame) os.Error {
	return e.Effect.AssignNums(s, f)
}

func (EffectNone) AssignNums(*Symtab, *numFrame) os.Error {
	return nil
}

func (e *EffectAnd) AssignNums(s *Symtab, f *numFrame) os.Error {
	if err := e.Left.AssignNums(s, f); err != nil {
		return err
	}
	return e.Right.AssignNums(s, f)
}

func (e *EffectForall) AssignNums(s *Symtab, f *numFrame) os.Error {
	f = e.Variable.Name.numberVar(s, f)
	for i, _ := range e.Variable.Type {
		if found := e.Variable.Type[i].numberType(s); !found {
			return undefType(&e.Variable.Type[i])
		}
	}
	return e.Effect.AssignNums(s, f)
}

func (e *EffectWhen) AssignNums(s *Symtab, f *numFrame) os.Error {
	if err := e.Condition.AssignNums(s, f); err != nil {
		return err
	}
	return e.Effect.AssignNums(s, f)
}

func (e *EffectLiteral) AssignNums(s *Symtab, f *numFrame) os.Error {
	return (*Literal)(e).AssignNums(s, f)
}

func (e *EffectAssign) AssignNums(*Symtab, *numFrame) os.Error {
	return nil
}

func (i *InitLiteral) AssignNums(s *Symtab, f *numFrame) os.Error {
	return (*Literal)(i).AssignNums(s, f)
}

func (i *InitEq) AssignNums(s *Symtab, f *numFrame) os.Error {
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

func undefType(n *Name) os.Error {
	return fmt.Errorf("%s: Undefined type %s\n", n.Loc, n.Str)
}