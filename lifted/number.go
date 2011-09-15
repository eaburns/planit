package lifted

// Uniquify variable names to be of the form "?x#".
// This eliminates any issues with variable shadowing.

import (
	"fmt"
	"os"
)

func (d *Domain) AssignNums(s *Symtab) os.Error {
	for i, _ := range d.Constants {
		d.Constants[i].Num = s.ConstNum(d.Constants[i].Name)
	}
	for i, _ := range d.Predicates {
		d.Predicates[i].Num = s.PredNum(d.Predicates[i].Name)
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
		fnxt, n := s.VarNum(f, a.Parameters[i].Name)
		f = fnxt
		a.Parameters[i].Num = n
	}
	if err := a.Precondition.AssignNums(s, f); err != nil {
		return err
	}
	return a.Effect.AssignNums(s, f)
}

func (p *Problem) AssignNums(s *Symtab) os.Error {
	for i, _ := range p.Objects {
		p.Objects[i].Num = s.ConstNum(p.Objects[i].Name)
	}
	return p.Goal.AssignNums(s, nil)
}

func (l *Literal) AssignNums(s *Symtab, f *numFrame) os.Error {
	for i, t := range l.Parameters {
		var n int
		switch t.Kind {
		case TermVariable:
			if fnxt, num := s.VarNum(f, t.Name); fnxt != f {
				return fmt.Errorf("%s: Unbound variable %s\n", t.Loc, t.Name)
			} else {
	 			n = num
			}
		case TermConstant:
			n = s.ConstNum(t.Name)
		}
		l.Parameters[i].Num = n
	}
	l.Num = s.PredNum(l.Name)
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
	f, n := s.VarNum(f, e.Variable.Name)
	e.Variable.Num = n
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
	f, n := s.VarNum(f, e.Variable.Name)
	e.Variable.Num = n
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
	constNums map[string]int
	constNames []string
	predNums map[string]int
	predNames []string
	varNames[]string
}

func NewSymtab() *Symtab {
	return &Symtab{
		constNums: make(map[string]int),
		predNums: make(map[string]int),
	}
}

func (s *Symtab) ConstNum(name string) int {
	if n, ok := s.constNums[name]; ok {
		return n
	}
	n := len(s.constNames)
	s.constNums[name] = n
	s.constNames = append(s.constNames, name)
	return n
}

func (s *Symtab) ConstName(n int) (string, bool) {
	if n < 0 || n > len(s.constNames) {
		return "", false
	}	
	return s.constNames[n], true
}

func (s *Symtab) PredNum(name string) int {
	if n, ok := s.predNums[name]; ok {
		return n
	}
	n := len(s.predNums)
	s.predNums[name] = n
	s.predNames = append(s.predNames, name)
	return n
}

func (s *Symtab) PredName(n int) (string, bool) {
	if n < 0 || n > len(s.predNames) {
		return "", false
	}	
	return s.predNames[n], true
}

func (s *Symtab) VarNum(f *numFrame, name string) (*numFrame, int) {
	if n, ok := f.lookup(name); ok {
		return f, n
	}
	n := len(s.varNames)
	s.varNames = append(s.varNames, name)
	return &numFrame{name: name, num: n, up: f}, n
}

func (s *Symtab) VarName(n int) (string, bool) {
	if n < 0 || n > len(s.varNames) {
		return "", false
	}	
	return s.varNames[n], true
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