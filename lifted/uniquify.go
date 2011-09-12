package lifted

// Uniquify variable names to be of the form "?x#".
// This eliminates any issues with variable shadowing.

import (
	"fmt"
	"os"
)

func (d *Domain) UniquifyVars() os.Error {
	for _, a := range d.Actions {
		if err := a.UniquifyVars(); err != nil {
			return err
		}
	}
	return nil
}

func (a *Action) UniquifyVars() os.Error {
	var f *uniqFrame
	for i, _ := range a.Parameters {
		var uniq string
		f, uniq = f.push(a.Parameters[i].Name)
		a.Parameters[i].Name = uniq
	}
	if err := a.Precondition.UniquifyVars(f); err != nil {
		return err
	}
	return a.Effect.UniquifyVars(f)
}

func (p *Problem) UniquifyVars() os.Error {
	return p.Goal.UniquifyVars(nil)
}

func (l *Literal) UniquifyVars(f *uniqFrame) os.Error {
	for i, t := range l.Parameters {
		if t.Kind != TermVariable {
			continue
		}
		uniq, ok := f.lookup(t.Name)
		if !ok {
			return fmt.Errorf("%s: Unbound variable %s\n", t.Loc, t.Name)
		}
		l.Parameters[i].Name = uniq
	}
	return nil
}

func (e *ExprBinary) UniquifyVars(f *uniqFrame) os.Error {
	if err := e.Left.UniquifyVars(f); err != nil {
		return err
	}
	return e.Right.UniquifyVars(f)
}

func (ExprTrue) UniquifyVars(*uniqFrame) os.Error { return nil }

func (ExprFalse) UniquifyVars(*uniqFrame) os.Error { return nil }

func (e *ExprAnd) UniquifyVars(f *uniqFrame) os.Error {
	return (*ExprBinary)(e).UniquifyVars(f)
}

func (e *ExprOr) UniquifyVars(f *uniqFrame) os.Error {
	return (*ExprBinary)(e).UniquifyVars(f)
}

func (e *ExprNot) UniquifyVars(f *uniqFrame) os.Error {
	return e.Expr.UniquifyVars(f)
}

func (e *ExprQuant) UniquifyVars(f *uniqFrame) os.Error {
	f, uniq := f.push(e.Variable.Name)
	e.Variable.Name = uniq
	return e.Expr.UniquifyVars(f)
}

func (e *ExprForall) UniquifyVars(f *uniqFrame) os.Error {
	return (*ExprQuant)(e).UniquifyVars(f)
}

func (e *ExprExists) UniquifyVars(f *uniqFrame) os.Error {
	return (*ExprQuant)(e).UniquifyVars(f)
}

func (e *ExprLiteral) UniquifyVars(f *uniqFrame) os.Error {
	return (*Literal)(e).UniquifyVars(f)
}

func (e *EffectUnary) UniquifyVars(f *uniqFrame) os.Error {
	return e.Effect.UniquifyVars(f)
}

func (EffectNone) UniquifyVars(*uniqFrame) os.Error { return nil }

func (e *EffectAnd) UniquifyVars(f *uniqFrame) os.Error {
	if err := e.Left.UniquifyVars(f); err != nil {
		return err
	}
	return e.Right.UniquifyVars(f)
}

func (e *EffectForall) UniquifyVars(f *uniqFrame) os.Error {
	f, uniq := f.push(e.Variable.Name)
	e.Variable.Name = uniq
	return e.Effect.UniquifyVars(f)
}

func (e *EffectWhen) UniquifyVars(f *uniqFrame) os.Error {
	if err := e.Condition.UniquifyVars(f); err != nil {
		return err
	}
	return e.Effect.UniquifyVars(f)
}

func (e *EffectLiteral) UniquifyVars(f *uniqFrame) os.Error {
	return (*Literal)(e).UniquifyVars(f)
}

func (e *EffectAssign) UniquifyVars(f *uniqFrame) os.Error { return nil }

type uniqFrame struct {
	name string
	uniq string
	num  int
	up   *uniqFrame
}

func (f *uniqFrame) push(name string) (*uniqFrame, string) {
	var num int
	var uniq string
	if f == nil {
		num = 0
		uniq = "?x0"
	} else {
		num = f.num + 1
		uniq = fmt.Sprintf("?x%d", num)
	}
	return &uniqFrame{name: name, uniq: uniq, num: num, up: f}, uniq
}

func (f *uniqFrame) lookup(name string) (string, bool) {
	if f == nil {
		return "", false
	}
	if f.name == name {
		return f.uniq, true
	}
	return f.up.lookup(name)
}
