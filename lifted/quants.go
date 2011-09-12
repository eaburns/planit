package lifted

// Quantifier removal

func (d *Domain) ExpandQuants(objs []TypedName) {
	f := newExpandFrame(objs)
	acts := make([]Action, 0, len(d.Actions))
	for i, _ := range d.Actions {
		a := &d.Actions[i]
		a.Precondition = a.Precondition.ExpandQuants(f)
		a.Effect = a.Effect.ExpandQuants(f)
		acts = append(acts, expandParams(f, a, a.Parameters)...)
	}
	d.Actions = acts
}

func expandParams(f *expandFrame, a *Action, ps []TypedName) (acts []Action) {
	if len(ps) == 0 {
		act := Action{
			Name:         a.Name,
			Parameters:   make([]TypedName, len(a.Parameters)),
			Precondition: a.Precondition.ExpandQuants(f),
			Effect:       a.Effect.ExpandQuants(f),
		}
		copy(act.Parameters, a.Parameters)
		return []Action{act}
	}

	pnum := len(a.Parameters) - len(ps)
	saved := a.Parameters[pnum]

	objs := objsOfType(f, saved.Type)
	for _, obj := range objs {
		a.Parameters[pnum].Name = obj
		g := f.push(saved.Name, obj)
		acts = append(acts, expandParams(g, a, ps[1:])...)
	}
	a.Parameters[pnum] = saved

	return
}

func (l *Literal) ExpandQuants(f *expandFrame) *Literal {
	res := &Literal{
		Positive:   l.Positive,
		Name:       l.Name,
		Parameters: make([]Term, len(l.Parameters)),
	}
	for i, t := range l.Parameters {
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
			Loc:  t.Loc,
		}
	}
	return res
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
		g := f.push(e.Variable.Name, obj)
		seq = comb(e.Expr.ExpandQuants(g), seq)
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
	return (*ExprLiteral)((*Literal)(e).ExpandQuants(f))
}

func (EffectNone) ExpandQuants(*expandFrame) Effect {
	return EffectNone(0)
}

func (e *EffectAnd) ExpandQuants(f *expandFrame) Effect {
	return EffectConj(e.Left.ExpandQuants(f), e.Right.ExpandQuants(f))
}

func (e *EffectForall) ExpandQuants(f *expandFrame) Effect {
	objs := objsOfType(f, e.Variable.Type)
	seq := Effect(EffectNone(0))
	for _, obj := range objs {
		g := f.push(e.Variable.Name, obj)
		seq = EffectConj(e.Effect.ExpandQuants(g), seq)
	}
	return seq
}

func (e *EffectWhen) ExpandQuants(f *expandFrame) Effect {
	return &EffectWhen{
		Condition:   e.Condition.ExpandQuants(f),
		EffectUnary: EffectUnary{e.Effect.ExpandQuants(f)},
	}
}

func (e *EffectLiteral) ExpandQuants(f *expandFrame) Effect {
	return (*EffectLiteral)((*Literal)(e).ExpandQuants(f))
}

func (e *EffectAssign) ExpandQuants(f *expandFrame) Effect {
	return &EffectAssign{
		Op:   e.Op,
		Lval: e.Lval,
		Rval: e.Rval,
	}
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

	return &expandFrame{objsByType: objsByType}
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
