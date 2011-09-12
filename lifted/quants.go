package lifted

// Quantifier removal

import ("fmt"
	"os")

func (d *Domain) ExpandQuants(objs []TypedName) {
	f := newExpandFrame(objs)
	for i, _ := range d.Actions {
		a := &d.Actions[i]
		a.Precondition = a.Precondition.ExpandQuants(f)
	}
}

func (e ExprTrue) ExpandQuants(*expandFrame) Expr {
	return e
}

func (e ExprFalse) ExpandQuants(*expandFrame) Expr {
	return e
}

func (e *ExprAnd) ExpandQuants(f *expandFrame) Expr {
	return ExprConj(e.Left.ExpandQuants(f), e.Right.ExpandQuants(f))
}

func (e *ExprOr) ExpandQuants(f *expandFrame) Expr {
	return ExprDisj(e.Left.ExpandQuants(f), e.Right.ExpandQuants(f))
}

func (e *ExprNot) ExpandQuants(f *expandFrame) Expr {
	return ExprNeg(e.Expr.ExpandQuants(f))
}

func (e *ExprQuant) ExpandQuants(f *expandFrame, base Expr,
comb func(Expr, Expr) Expr) Expr {
	objs := objsOfType(f, e.Variable.Type)
	seq := base
	for _, obj := range objs {
		fmt.Fprintf(os.Stderr, "binding %s to %s\n", e.Variable.Name, obj)
		g := f.push(e.Variable.Name, obj)
		seq = comb(e.Expr.ExpandQuants(g), seq)
		fmt.Fprintf(os.Stderr, "unbinding %s to %s\n", e.Variable.Name, obj)
	}
	return seq
}

func (e *ExprForall) ExpandQuants(f *expandFrame) Expr {
	return (*ExprQuant)(e).ExpandQuants(f, ExprTrue(1), ExprConj)
}

func (e *ExprExists) ExpandQuants(f *expandFrame) Expr {
	return (*ExprQuant)(e).ExpandQuants(f, ExprFalse(0), ExprDisj)
}

func (e *ExprLiteral) ExpandQuants(f *expandFrame) Expr {
	res := &ExprLiteral{
		Positive:   e.Positive,
		Name:       e.Name,
		Parameters: make([]Term, len(e.Parameters)),
	}
	for i, t := range e.Parameters {
		if t.Kind == TermConstant {
			res.Parameters[i] = t
			continue
		}
		o, ok := f.lookup(t.Name)
		if !ok {
			res.Parameters[i] = t
			continue
		}
		res.Parameters[i] = Term{
			Kind: TermConstant,
			Name: o,
			Loc: t.Loc,
		}
	}
	return res
}

type expandFrame struct {
	objsByType map[string][]string
	variable   string
	object     string
	up         *expandFrame
}

func newExpandFrame(objs []TypedName) *expandFrame {
	objsByType := make(map[string][]string)

	for _, obj := range objs {
		for _, t := range obj.Type {
			lst, _ := objsByType[t]
			objsByType[t] = append(lst, obj.Name)
		}
	}

	return &expandFrame{ objsByType: objsByType }
}

func (f *expandFrame) push(vr string, obj string) *expandFrame {
	return &expandFrame{
		objsByType: f.objsByType,
		variable:   vr,
		object:     obj,
		up:         f,
	}
}

func (f *expandFrame) lookup(vr string) (string, bool) {
	if f == nil {
		return "", false
	}
	if f.variable == vr {
		return f.object, true
	}
	return f.up.lookup(vr)
}

func objsOfType(f *expandFrame, typ []string) (objs []string) {
	seen := make(map[string]bool)
	for _, t := range typ {
		for _, o := range f.objsByType[t] {
			if _, ok := seen[o]; ok {
				continue
			}
			seen[o] = true
			objs = append(objs, o)
		}
	}
	return
}
