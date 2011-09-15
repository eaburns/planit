package lifted

// Uniquify variable names to be of the form "?x#".
// This eliminates any issues with variable shadowing.

import (
	"fmt"
	"os"
)

func (d *Domain) AssignNums(s *Symtab) os.Error {
	for i, _ := range d.Types {
		s.types.Number(&d.Types[i].Name)
		for j, _ := range d.Types[i].Type {
			s.types.Number(&d.Types[i].Type[j])
		}
	}
	for i, _ := range d.Constants {
		s.consts.Number(&d.Constants[i].Name)
	}
	for i, _ := range d.Predicates {
		s.preds.Number(&d.Predicates[i].Name)
	}
	for _, a := range d.Actions {
		if err := a.AssignNums(s); err != nil {
			return err
		}
	}
	return nil
}

func (a *Action) AssignNums(s *Symtab) os.Error {
	var f *numFrame
	for i, _ := range a.Parameters {
		f = s.VarNum(f, &a.Parameters[i].Name)
	}
	if err := a.Precondition.AssignNums(s, f); err != nil {
		return err
	}
	return a.Effect.AssignNums(s, f)
}

func (p *Problem) AssignNums(s *Symtab) os.Error {
	for i, _ := range p.Objects {
		s.consts.Number(&p.Objects[i].Name)
	}
	return p.Goal.AssignNums(s, nil)
}

func (l *Literal) AssignNums(s *Symtab, f *numFrame) os.Error {
	for i, t := range l.Parameters {
		switch t.Kind {
		case TermVariable:
			if fnxt := s.VarNum(f, &l.Parameters[i].Name); fnxt == f {
				break
			}
			name := l.Parameters[i].Name.String
			return fmt.Errorf("%s: Unbound variable %s\n", t.Loc, name)
		case TermConstant:
			s.consts.Number(&l.Parameters[i].Name)
		}
	}
	s.preds.Number(&l.Name)
	return nil
}

func (e *ExprBinary) AssignNums(s *Symtab, f *numFrame) os.Error {
	if err := e.Left.AssignNums(s, f); err != nil {
		return err
	}
	return e.Right.AssignNums(s, f)
}

func (ExprTrue) AssignNums(*Symtab, *numFrame) os.Error { return nil }

func (ExprFalse) AssignNums(*Symtab, *numFrame) os.Error { return nil }

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
	f = s.VarNum(f, &e.Variable.Name)
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

func (EffectNone) AssignNums(*Symtab, *numFrame) os.Error { return nil }

func (e *EffectAnd) AssignNums(s *Symtab, f *numFrame) os.Error {
	if err := e.Left.AssignNums(s, f); err != nil {
		return err
	}
	return e.Right.AssignNums(s, f)
}

func (e *EffectForall) AssignNums(s *Symtab, f *numFrame) os.Error {
	f = s.VarNum(f, &e.Variable.Name)
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

func (e *EffectAssign) AssignNums(*Symtab, *numFrame) os.Error { return nil }

type Symtab struct {
	consts   CoatCheck
	preds    CoatCheck
	types    CoatCheck
	varNames []string
}

func NewSymtab() *Symtab {
	return &Symtab{
		consts: MakeCoatCheck(),
		preds:  MakeCoatCheck(),
		types:  MakeCoatCheck(),
	}
}

type CoatCheck struct {
	nums    map[string]int
	strings []string
}

func MakeCoatCheck() CoatCheck {
	return CoatCheck{nums: make(map[string]int)}
}

func (n *CoatCheck) Number(name *Name) {
	if num, ok := n.nums[name.String]; ok {
		name.Number = num
	}
	num := len(n.strings)
	n.nums[name.String] = num
	n.strings = append(n.strings, name.String)
	name.Number = num
}

func (s *Symtab) VarNum(f *numFrame, name *Name) *numFrame {
	if n, ok := f.lookup(name.String); ok {
		name.Number = n
		return f
	}
	n := len(s.varNames)
	s.varNames = append(s.varNames, name.String)
	name.Number = n
	return &numFrame{name: name.String, num: n, up: f}
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
