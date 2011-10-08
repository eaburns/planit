package pddl

import . "goplan/lifted"

func (p *Parser) ParseDomain() *Domain {
	p.expect(tokOpen)
	p.expectId("define")
	d := &Domain{
		Name:         p.parseDomainName(),
		Requirements: p.parseReqsDef(),
		Types:        p.parseTypesDef(),
		Constants:    p.parseConstsDef(),
		Predicates:   p.parsePredsDef(),
	}
	// Ignore :functions for now
	if p.acceptNamedList(":functions") {
		for nesting := 1; nesting > 0; {
			switch p.next().typ {
			case tokClose:
				nesting--
			case tokOpen:
				nesting++
			}
		}
	}
	for p.peek().typ == tokOpen {
		d.Actions = append(d.Actions, p.parseActionDef())
	}

	p.expect(tokClose)
	return d
}

func (p *Parser) parseDomainName() string {
	p.expect(tokOpen)
	p.expectId("domain")
	n := p.expect(tokId)
	p.expect(tokClose)
	return n.txt
}

func (p *Parser) parseReqsDef() (reqs []string) {
	if p.acceptNamedList(":requirements") {
		for t, ok := p.accept(tokCid); ok; t, ok = p.accept(tokCid) {
			reqs = append(reqs, t.txt)
		}
		p.expect(tokClose)
	}
	return
}

func (p *Parser) parseTypesDef() (types []TypedName) {
	if p.acceptNamedList(":types") {
		types = p.parseTypedListString(tokId)
		p.expect(tokClose)
	}
	object := false
	for i := range types {
		if types[i].Name.Str == "object" {
			object = true
		}
	}
	if !object {
		objname := MakeName("object", Loc{"<implicit>", -1})
		types = append(types, TypedName{Name: objname})
	}
	return
}

func (p *Parser) parseConstsDef() (consts []TypedName) {
	if p.acceptNamedList(":constants") {
		consts = p.parseTypedListString(tokId)
		p.expect(tokClose)
	}
	return
}

func (p *Parser) parsePredsDef() (Predicates []Predicate) {
	if p.acceptNamedList(":predicates") {
		Predicates = append(Predicates, p.parseAtomicFormSkele())
		for p.peek().typ == tokOpen {
			Predicates = append(Predicates, p.parseAtomicFormSkele())
		}
		p.expect(tokClose)
	}
	return
}

func (p *Parser) parseAtomicFormSkele() Predicate {
	p.expect(tokOpen)
	pred := Predicate{
		Name:       p.name(p.expect(tokId).txt),
		Parameters: p.parseTypedListString(tokQid),
	}
	p.expect(tokClose)
	return pred
}

func (p *Parser) parseActionDef() Action {
	p.expect(tokOpen)
	p.expectId(":action")

	act := Action{Name: p.expect(tokId).txt, Parameters: p.parseActParms()}

	if p.peek().txt == ":precondition" {
		p.junk(1)
		if p.peek().typ == tokOpen && p.peekn(2).typ == tokClose {
			p.junk(2)
		} else {
			act.Precondition = p.parsePreExpr()
		}
	}
	if p.peek().txt == ":effect" {
		p.junk(1)
		if p.peek().typ == tokOpen && p.peekn(2).typ == tokClose {
			p.junk(2)
		} else {
			act.Effect = p.parseEffect()
		}
	}

	p.expect(tokClose)
	return act
}

func (p *Parser) parseActParms() []TypedName {
	p.expectId(":parameters")
	p.expect(tokOpen)
	res := p.parseTypedListString(tokQid)
	p.expect(tokClose)
	return res
}

func (p *Parser) parsePreExpr() (res Formula) {
	parseNested := func(p *Parser) Formula { return p.parsePreExpr() }
	switch {
	case p.acceptNamedList("and"):
		res = p.parseAndExpr(parseNested)
	case p.acceptNamedList("forall"):
		res = p.parseForallExpr(parseNested)
	default:
		res = p.parsePrefExpr()
	}
	return
}

func (p *Parser) parsePrefExpr() Formula {
	return p.parseExpr()
}

func (p *Parser) parseExpr() (res Formula) {
	parseNested := func(p *Parser) Formula { return p.parseExpr() }
	switch {
	case p.acceptNamedList("and"):
		res = p.parseAndExpr(parseNested)
	case p.acceptNamedList("or"):
		res = p.parseOrExpr(parseNested)
	case p.acceptNamedList("not"):
		res = &NotNode{UnaryNode{Formula: p.parseExpr()}}
		p.expect(tokClose)
	case p.acceptNamedList("imply"):
		res = &OrNode{
			BinaryNode{
				Left:  &NotNode{UnaryNode{Formula: p.parseExpr()}},
				Right: p.parseExpr(),
			},
		}
		p.expect(tokClose)
	case p.acceptNamedList("exists"):
		res = p.parseExistsExpr(parseNested)
	case p.acceptNamedList("forall"):
		res = p.parseForallExpr(parseNested)
	default:
		res = p.parseLiteral()
	}
	return
}

func (p *Parser) parseLiteral() *LiteralNode {
	pos := true
	if p.acceptNamedList("not") {
		pos = false
	}
	p.expect(tokOpen)
	res := &LiteralNode{
		Positive:   pos,
		Name:       p.name(p.expect(tokId).txt),
		Parameters: p.parseTerms(),
	}
	if !pos {
		p.expect(tokClose)
	}
	p.expect(tokClose)
	return res
}

func (p *Parser) parseTerms() (lst []Term) {
	for {
		if t, ok := p.accept(tokId); ok {
			lst = append(lst, Term{
				Kind: TermConstant,
				Name: p.name(t.txt),
			})
			continue
		}
		if t, ok := p.accept(tokQid); ok {
			lst = append(lst, Term{
				Kind: TermVariable,
				Name: p.name(t.txt),
			})
			continue
		}
		break
	}
	return
}

func (p *Parser) parseAndExpr(nested func(*Parser) Formula) Formula {
	conj := make([]Formula, 0)
	for p.peek().typ == tokOpen {
		conj = append(conj, nested(p))
	}
	e := Formula(TrueNode{})
	for i := len(conj) - 1; i >= 0; i-- {
		e = Conjunct(conj[i], e)
	}
	p.expect(tokClose)
	return e
}

func (p *Parser) parseOrExpr(nested func(*Parser) Formula) Formula {
	disj := make([]Formula, 0)
	for p.peek().typ == tokOpen {
		disj = append(disj, nested(p))
	}
	e := Formula(FalseNode{})
	for i := len(disj) - 1; i >= 0; i-- {
		e = Disjunct(disj[i], e)
	}
	p.expect(tokClose)
	return e
}

func (p *Parser) parseForallExpr(nested func(*Parser) Formula) Formula {
	p.expect(tokOpen)
	vrs := p.parseTypedListString(tokQid)
	p.expect(tokClose)

	res := &ForallNode{}
	bottom := res
	for i, vr := range vrs {
		bottom.Variable = vr
		if i < len(vrs)-1 {
			bottom.Formula = &ForallNode{}
			bottom = bottom.Formula.(*ForallNode)
		}
	}

	bottom.Formula = nested(p)
	p.expect(tokClose)
	return res
}

func (p *Parser) parseExistsExpr(nested func(*Parser) Formula) Formula {
	p.expect(tokOpen)
	vrs := p.parseTypedListString(tokQid)
	p.expect(tokClose)

	res := &ExistsNode{}
	bottom := res
	for i, vr := range vrs {
		bottom.Variable = vr
		if i < len(vrs)-1 {
			bottom.Formula = &ExistsNode{}
			bottom = bottom.Formula.(*ExistsNode)
		}
	}

	bottom.Formula = nested(p)
	p.expect(tokClose)
	return res
}

func (p *Parser) parseEffect() Formula {
	if p.acceptNamedList("and") {
		parseNested := func(p *Parser) Formula {
			return p.parseCeffect()
		}
		return p.parseAndEffect(parseNested)
	}
	return p.parseCeffect()
}

func (p *Parser) parseAndEffect(nested func(*Parser) Formula) Formula {
	conj := make([]Formula, 0)
	for p.peek().typ == tokOpen {
		conj = append(conj, nested(p))
	}
	e := Formula(TrueNode{})
	for i := len(conj) - 1; i >= 0; i-- {
		e = Conjunct(conj[i], e)
	}
	p.expect(tokClose)
	return e
}

func (p *Parser) parseCeffect() (res Formula) {
	switch {
	case p.acceptNamedList("forall"):
		parseNested := func(p *Parser) Formula {
			return p.parseEffect()
		}
		res = p.parseForallEffect(parseNested)
	case p.acceptNamedList("when"):
		parseNested := func(p *Parser) Formula {
			return p.parseCondEffect()
		}
		res = p.parseWhen(parseNested)
	default:
		res = p.parsePeffect()
	}
	return
}

func (p *Parser) parseForallEffect(nested func(*Parser) Formula) Formula {
	p.expect(tokOpen)
	vrs := p.parseTypedListString(tokQid)
	p.expect(tokClose)

	res := &ForallNode{}
	bottom := res
	for i, vr := range vrs {
		bottom.Variable = vr
		if i < len(vrs)-1 {
			bottom.Formula = &ForallNode{}
			bottom = bottom.Formula.(*ForallNode)
		}
	}

	bottom.Formula = nested(p)
	p.expect(tokClose)
	return res
}

func (p *Parser) parseWhen(nested func(*Parser) Formula) Formula {
	res := &WhenNode{
		Condition: p.parseExpr(),
	}
	res.Formula = nested(p)
	p.expect(tokClose)
	return res
}

func (p *Parser) parsePeffect() Formula {
	if _, ok := AssignOps[p.peekn(2).txt]; ok && p.peek().typ == tokOpen {
		return p.parseAssign()
	}
	return p.parseLiteral()
}

func (p *Parser) parseAssign() Formula {
	p.expect(tokOpen)
	res := &AssignNode{
		Op:   AssignOps[p.expect(tokId).txt],
		Lval: p.parseFhead(),
		Rval: p.parseFexp(),
	}
	p.expect(tokClose)
	return res
}

func (p *Parser) parseCondEffect() Formula {
	if p.acceptNamedList("and") {
		parseNested := func(p *Parser) Formula {
			return p.parsePeffect()
		}
		return p.parseAndEffect(parseNested)
	}
	return p.parsePeffect()
}

func (p *Parser) parseFhead() Fhead {
	if _, ok := p.accept(tokOpen); !ok {
		return Fhead{Name: p.expect(tokId).txt}
	}
	name := p.expect(tokId).txt
	p.expect(tokClose)
	return Fhead{Name: name}
}

func (p *Parser) parseFexp() Fexp {
	return Fexp(p.expect(tokNum).txt)
}

func (p *Parser) ParseProblem() *Problem {
	p.expect(tokOpen)
	p.expectId("define")
	prob := &Problem{
		Name:         p.parseProbName(),
		Domain:       p.parseProbDomain(),
		Requirements: p.parseReqsDef(),
		Objects:      p.parseObjsDecl(),
		Init:         p.parseInit(),
		Goal:         p.parseGoal(),
		Metric:       p.parseMetric(),
	}
	p.expect(tokClose)
	return prob
}

func (p *Parser) parseProbName() string {
	p.expect(tokOpen)
	p.expectId("problem")
	name := p.expect(tokId).txt
	p.expect(tokClose)
	return name
}

func (p *Parser) parseProbDomain() string {
	p.expect(tokOpen)
	p.expectId(":domain")
	name := p.expect(tokId).txt
	p.expect(tokClose)
	return name
}

func (p *Parser) parseObjsDecl() (objs []TypedName) {
	if p.acceptNamedList(":objects") {
		objs = p.parseTypedListString(tokId)
		p.expect(tokClose)
	}
	return
}

func (p *Parser) parseInit() (els []Formula) {
	p.expect(tokOpen)
	p.expectId(":init")
	for p.peek().typ == tokOpen {
		els = append(els, p.parseInitEl())
	}
	p.expect(tokClose)
	return
}

func (p *Parser) parseInitEl() Formula {
	if p.acceptNamedList("=") {
		eq := &AssignNode{
			Op: OpAssign,
			Lval: p.parseFhead(),
			Rval: Fexp(p.expect(tokNum).txt),
		}
		p.expect(tokClose)
		return eq
	}
	return p.parseLiteral()
}

func (p *Parser) parseGoal() Formula {
	p.expect(tokOpen)
	p.expectId(":goal")
	g := p.parsePreExpr()
	p.expect(tokClose)
	return g
}

func (p *Parser) parseMetric() (m Metric) {
	if p.acceptNamedList(":metric") {
		m = MetricMinCost
		p.expectId("minimize")
		p.expect(tokOpen)
		p.expectId("total-cost")
		p.expect(tokClose)
		p.expect(tokClose)
	}
	return
}

func (p *Parser) parseTypedListString(typ tokenType) (lst []TypedName) {
	for {
		names := p.parseStrings(typ)
		if len(names) == 0 {
			break
		}
		typ := p.parseType()
		for _, n := range names {
			name := p.name(n)
			lst = append(lst, TypedName{Name: name, Type: typ})
		}
	}
	return lst
}

func (p *Parser) parseType() (typ []Name) {
	if _, ok := p.accept(tokMinus); !ok {
		return []Name{p.name("object")}
	}
	if _, ok := p.accept(tokOpen); ok {
		p.expectId("either")
		for _, s := range p.parseStringPlus(tokId) {
			typ = append(typ, p.name(s))
		}
		p.expect(tokClose)
		return typ
	}
	t := p.expect(tokId)
	return []Name{p.name(t.txt)}
}

func (p *Parser) parseStringPlus(typ tokenType) []string {
	lst := []string{p.expect(typ).txt}
	lst = append(lst, p.parseStrings(typ)...)
	return lst
}

func (p *Parser) parseStrings(typ tokenType) (lst []string) {
	for t, ok := p.accept(typ); ok; t, ok = p.accept(typ) {
		lst = append(lst, t.txt)
	}
	return lst
}

func (p *Parser) name(txt string) Name {
	return MakeName(txt, p.loc())
}
