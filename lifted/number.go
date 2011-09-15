package lifted

// Assign numbers to all things of type Name

import (
	"fmt"
	"os"
)

var (
	constNums = make(map[string]int)
	constNames []string

	predNums = make(map[string]int)
	predNames []string

	typeNums = make(map[string]int)
	typeNames []string
	typeObjs [][]int

	varNames []string
)

func (d *Domain) AssignNums() os.Error {
	for i, _ := range d.Types {
		t := &d.Types[i]
		t.Name.numberType()
		for j, _ := range t.Type {
			t.Type[j].numberType()
		}
	}
	typeObjs = make([][]int, len(typeNames))
	numberConsts(d.Constants)
	for i, _ := range d.Predicates {
		d.Predicates[i].Name.numberPred()
	}
	for _, a := range d.Actions {
		if err := a.AssignNums(); err != nil {
			return err
		}
	}
	return nil
}

func numberConsts(consts []TypedName) {
	for i, _ := range consts {
		c := &consts[i]
		first := c.Name.numberConst()
		cnum := c.Name.Num
		for j, _ := range c.Type {
			c.Type[j].numberType()
			// If this is the 1st decl of this object
			// then add it to the table of all objects
			// of the given type
			if !first {
				tnum := c.Type[j].Num
				typeObjs[tnum] = append(typeObjs[tnum], cnum)
			}
		}
	}
}

func (a *Action) AssignNums() os.Error {
	var f *numFrame
	for i, _ := range a.Parameters {
		p := &a.Parameters[i]
		f = p.Name.numberVar(f)
		for j, _ := range p.Type {
			p.Type[j].numberType()
		}
	}
	if err := a.Precondition.AssignNums(f); err != nil {
		return err
	}
	return a.Effect.AssignNums(f)
}

func (p *Problem) AssignNums() os.Error {
	numberConsts(p.Objects)
	for _, init := range p.Init {
		if err := init.AssignNums(nil); err != nil {
			return err
		}
	}
	return p.Goal.AssignNums(nil)
}

func (l *Literal) AssignNums(f *numFrame) os.Error {
	for i, t := range l.Parameters {
		name := &l.Parameters[i].Name
		switch t.Kind {
		case TermVariable:
			if fnxt := name.numberVar(f); fnxt == f {
				break
			}
			return fmt.Errorf("%s: Unbound variable %s\n", t.Loc, name.Str)
		case TermConstant:
			name.numberConst()
		}
	}
	l.Name.numberPred()
	return nil
}

func (e *ExprBinary) AssignNums(f *numFrame) os.Error {
	if err := e.Left.AssignNums(f); err != nil {
		return err
	}
	return e.Right.AssignNums(f)
}

func (ExprTrue) AssignNums(*numFrame) os.Error { return nil }

func (ExprFalse) AssignNums(*numFrame) os.Error { return nil }

func (e *ExprAnd) AssignNums(f *numFrame) os.Error {
	return (*ExprBinary)(e).AssignNums(f)
}

func (e *ExprOr) AssignNums(f *numFrame) os.Error {
	return (*ExprBinary)(e).AssignNums(f)
}

func (e *ExprNot) AssignNums(f *numFrame) os.Error {
	return e.Expr.AssignNums(f)
}

func (e *ExprQuant) AssignNums(f *numFrame) os.Error {
	f = e.Variable.Name.numberVar(f)
	for i, _ := range e.Variable.Type {
		e.Variable.Type[i].numberType()
	}
	return e.Expr.AssignNums(f)
}

func (e *ExprForall) AssignNums(f *numFrame) os.Error {
	return (*ExprQuant)(e).AssignNums(f)
}

func (e *ExprExists) AssignNums(f *numFrame) os.Error {
	return (*ExprQuant)(e).AssignNums(f)
}

func (e *ExprLiteral) AssignNums(f *numFrame) os.Error {
	return (*Literal)(e).AssignNums(f)
}

func (e *EffectUnary) AssignNums(f *numFrame) os.Error {
	return e.Effect.AssignNums(f)
}

func (EffectNone) AssignNums(*numFrame) os.Error { return nil }

func (e *EffectAnd) AssignNums(f *numFrame) os.Error {
	if err := e.Left.AssignNums(f); err != nil {
		return err
	}
	return e.Right.AssignNums(f)
}

func (e *EffectForall) AssignNums(f *numFrame) os.Error {
	f = e.Variable.Name.numberVar(f)
	for i, _ := range e.Variable.Type {
		e.Variable.Type[i].numberType()
	}
	return e.Effect.AssignNums(f)
}

func (e *EffectWhen) AssignNums(f *numFrame) os.Error {
	if err := e.Condition.AssignNums(f); err != nil {
		return err
	}
	return e.Effect.AssignNums(f)
}

func (e *EffectLiteral) AssignNums(f *numFrame) os.Error {
	return (*Literal)(e).AssignNums(f)
}

func (e *EffectAssign) AssignNums(*numFrame) os.Error { return nil }

func (i *InitLiteral) AssignNums(f *numFrame) os.Error {
	return (*Literal)(i).AssignNums(f)
}

func (i *InitEq) AssignNums(f *numFrame) os.Error { return nil }

func (name *Name) numberType() {
	if n, ok := typeNums[name.Str]; ok {
		name.Num = n
	} else {
		name.Num = len(typeNames)
		typeNums[name.Str] = name.Num
		typeNames = append(typeNames, name.Str)
	}
}

func (name *Name) numberConst() bool {
	if n, ok := constNums[name.Str]; ok {
		name.Num = n
		return true
	}
	name.Num = len(constNames)
	constNums[name.Str] = name.Num
	constNames = append(constNames, name.Str)
	return false
}

func (name *Name) numberPred() {
	if n, ok := predNums[name.Str]; ok {
		name.Num = n
	} else {
		name.Num = len(predNames)
		predNums[name.Str] = name.Num
		predNames = append(predNames, name.Str)
	}
}

func (name *Name) numberVar(f *numFrame) *numFrame {
	if n, ok := f.lookup(name.Str); ok {
		name.Num = n
		return f
	}
	n := len(varNames)
	name.Num = n
	varNames = append(varNames, name.Str)
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