package lifted

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
	var f *frame
	for i, p := range a.Parameters {
		var uniq string
		f, uniq = f.push(p.Name)
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

func (l *Literal) UniquifyVars(f *frame) os.Error {
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

func (e *ExprBinary) UniquifyVars(f *frame) os.Error {
	if err := e.Left.UniquifyVars(f); err != nil {
		return err
	}
	return e.Right.UniquifyVars(f)
}

func (ExprTrue) UniquifyVars(*frame) os.Error { return nil }

func (ExprFalse) UniquifyVars(*frame) os.Error { return nil }

func (e *ExprAnd) UniquifyVars(f *frame) os.Error {
	return (*ExprBinary)(e).UniquifyVars(f)
}

func (e *ExprOr) UniquifyVars(f *frame) os.Error {
	return (*ExprBinary)(e).UniquifyVars(f)
}

func (e *ExprNot) UniquifyVars(f *frame) os.Error {
	return e.Expr.UniquifyVars(f)
}

func (e *ExprQuant) UniquifyVars(f *frame) os.Error {
	f, uniq := f.push(e.Variable.Name)
	e.Variable.Name = uniq
	return e.Expr.UniquifyVars(f)
}

func (e *ExprForall) UniquifyVars(f *frame) os.Error {
	return (*ExprQuant)(e).UniquifyVars(f)
}

func (e *ExprExists) UniquifyVars(f *frame) os.Error {
	return (*ExprQuant)(e).UniquifyVars(f)
}

func (e *ExprLiteral) UniquifyVars(f *frame) os.Error {
	return (*Literal)(e).UniquifyVars(f)
}

func (e *EffectUnary) UniquifyVars(f *frame) os.Error {
	return e.Effect.UniquifyVars(f)
}

func (EffectNone) UniquifyVars(*frame) os.Error { return nil }

func (e *EffectAnd) UniquifyVars(f *frame) os.Error {
	if err := e.Left.UniquifyVars(f); err != nil {
		return err
	}
	return e.Right.UniquifyVars(f)
}

func (e *EffectForall) UniquifyVars(f *frame) os.Error {
	f, uniq := f.push(e.Variable.Name)
	e.Variable.Name = uniq
	return e.Effect.UniquifyVars(f)
}

func (e *EffectWhen) UniquifyVars(f *frame) os.Error {
	if err := e.Condition.UniquifyVars(f); err != nil {
		return err
	}
	return e.Effect.UniquifyVars(f)
}

func (e *EffectLiteral) UniquifyVars(f *frame) os.Error {
	return (*Literal)(e).UniquifyVars(f)
}

func (e *EffectAssign) UniquifyVars(f *frame) os.Error { return nil }

type frame struct {
	name string
	uniq string
	num  int
	up   *frame
}

func (f *frame) push(name string) (*frame, string) {
	var num int
	var uniq string
	if f == nil {
		num = 0
		uniq = "?x0"
	} else {
		num = f.num + 1
		uniq = fmt.Sprintf("?x%d", num)
	}
	return &frame{name: name, uniq: uniq, num: num, up: f}, uniq
}

func (f *frame) lookup(name string) (string, bool) {
	if f == nil {
		return "", false
	}
	if f.name == name {
		return f.uniq, true
	}
	return f.up.lookup(name)
}