package pddl

import . "planit/prob"

func (p *parser) ParseDomain() *Domain {
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

func (p *parser) parseDomainName() string {
	p.expect(tokOpen)
	p.expectId("domain")
	n := p.expect(tokId)
	p.expect(tokClose)
	return n.text
}

func (p *parser) parseReqsDef() (reqs []string) {
	if p.acceptNamedList(":requirements") {
		for t, ok := p.accept(tokCid); ok; t, ok = p.accept(tokCid) {
			reqs = append(reqs, t.text)
		}
		p.expect(tokClose)
	}
	return
}

func (p *parser) parseTypesDef() (types []TypedName) {
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

func (p *parser) parseConstsDef() (consts []TypedName) {
	if p.acceptNamedList(":constants") {
		consts = p.parseTypedListString(tokId)
		p.expect(tokClose)
	}
	return
}

func (p *parser) parsePredsDef() (Predicates []Predicate) {
	if p.acceptNamedList(":predicates") {
		Predicates = append(Predicates, p.parseAtomicFormSkele())
		for p.peek().typ == tokOpen {
			Predicates = append(Predicates, p.parseAtomicFormSkele())
		}
		p.expect(tokClose)
	}
	return
}

func (p *parser) parseAtomicFormSkele() Predicate {
	p.expect(tokOpen)
	pred := Predicate{
		Name:       p.name(p.expect(tokId).text),
		Parameters: p.parseTypedListString(tokQid),
	}
	p.expect(tokClose)
	return pred
}

func (p *parser) parseActionDef() Action {
	p.expect(tokOpen)
	p.expectId(":action")

	act := Action{Name: p.expect(tokId).text, Parameters: p.parseActParms()}

	if p.peek().text == ":precondition" {
		p.junk(1)
		if p.peek().typ == tokOpen && p.peekn(2).typ == tokClose {
			p.junk(2)
		} else {
			act.Precondition = p.parsePreGd()
		}
	}
	if p.peek().text == ":effect" {
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

func (p *parser) parseActParms() []TypedName {
	p.expectId(":parameters")
	p.expect(tokOpen)
	res := p.parseTypedListString(tokQid)
	p.expect(tokClose)
	return res
}

func (p *parser) parsePreGd() (res Formula) {
	switch {
	case p.acceptNamedList("and"):
		res = p.parseAndGd((*parser).parsePreGd)
	case p.acceptNamedList("forall"):
		res = p.parseForallGd((*parser).parsePreGd)
	default:
		res = p.parsePrefGd()
	}
	return
}

func (p *parser) parsePrefGd() Formula {
	return p.parseGd()
}

func (p *parser) parseGd() (res Formula) {
	switch {
	case p.acceptNamedList("and"):
		res = p.parseAndGd((*parser).parseGd)
	case p.acceptNamedList("or"):
		res = p.parseOrGd((*parser).parseGd)
	case p.acceptNamedList("not"):
		res = &NotNode{UnaryNode{Formula: p.parseGd()}}
		p.expect(tokClose)
	case p.acceptNamedList("imply"):
		res = &OrNode{
			BinaryNode{
				Left:  &NotNode{UnaryNode{Formula: p.parseGd()}},
				Right: p.parseGd(),
			},
		}
		p.expect(tokClose)
	case p.acceptNamedList("exists"):
		res = p.parseExistsGd((*parser).parseGd)
	case p.acceptNamedList("forall"):
		res = p.parseForallGd((*parser).parseGd)
	default:
		res = p.parseLiteral()
	}
	return
}

func (p *parser) parseLiteral() *Literal {
	pos := true
	if p.acceptNamedList("not") {
		pos = false
		defer p.expect(tokClose)
	}
	p.expect(tokOpen)
	res := &Literal{
		Predicate:  p.name(p.expect(tokId).text),
		Positive:   pos,
		Parameters: p.parseTerms(),
	}
	p.expect(tokClose)
	return res
}

func (p *parser) parseTerms() (lst []Term) {
	for {
		if t, ok := p.accept(tokId); ok {
			lst = append(lst, Term{Name: p.name(t.text)})
			continue
		}
		if t, ok := p.accept(tokQid); ok {
			lst = append(lst, Term{Name: p.name(t.text), Variable: true})
			continue
		}
		break
	}
	return
}

func (p *parser) parseAndGd(nested func(*parser) Formula) Formula {
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

func (p *parser) parseOrGd(nested func(*parser) Formula) Formula {
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

func (p *parser) parseForallGd(nested func(*parser) Formula) Formula {
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

func (p *parser) parseExistsGd(nested func(*parser) Formula) Formula {
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

func (p *parser) parseEffect() Formula {
	if p.acceptNamedList("and") {
		return p.parseAndEffect((*parser).parseCeffect)
	}
	return p.parseCeffect()
}

func (p *parser) parseAndEffect(nested func(*parser) Formula) Formula {
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

func (p *parser) parseCeffect() (res Formula) {
	switch {
	case p.acceptNamedList("forall"):
		res = p.parseForallEffect((*parser).parseEffect)
	case p.acceptNamedList("when"):
		res = p.parseWhen((*parser).parseCondEffect)
	default:
		res = p.parsePeffect()
	}
	return
}

func (p *parser) parseForallEffect(nested func(*parser) Formula) Formula {
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

func (p *parser) parseWhen(nested func(*parser) Formula) Formula {
	res := &WhenNode{
		Condition: p.parseGd(),
	}
	res.Formula = nested(p)
	p.expect(tokClose)
	return res
}

func (p *parser) parsePeffect() Formula {
	if _, ok := AssignOps[p.peekn(2).text]; ok && p.peek().typ == tokOpen {
		return p.parseAssign()
	}
	return p.parseLiteral()
}

func (p *parser) parseAssign() Formula {
	p.expect(tokOpen)
	res := &AssignNode{
		Op:   AssignOps[p.expect(tokId).text],
		Lval: p.parseFhead(),
		Rval: p.parseFexp(),
	}
	p.expect(tokClose)
	return res
}

func (p *parser) parseCondEffect() Formula {
	if p.acceptNamedList("and") {
		return p.parseAndEffect((*parser).parsePeffect)
	}
	return p.parsePeffect()
}

func (p *parser) parseFhead() string {
	if _, ok := p.accept(tokOpen); !ok {
		return p.expect(tokId).text
	}
	name := p.expect(tokId).text
	p.expect(tokClose)
	return name
}

func (p *parser) parseFexp() string {
	return p.expect(tokNum).text
}

func (p *parser) ParseProblem() *Problem {
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

func (p *parser) parseProbName() string {
	p.expect(tokOpen)
	p.expectId("problem")
	name := p.expect(tokId).text
	p.expect(tokClose)
	return name
}

func (p *parser) parseProbDomain() string {
	p.expect(tokOpen)
	p.expectId(":domain")
	name := p.expect(tokId).text
	p.expect(tokClose)
	return name
}

func (p *parser) parseObjsDecl() (objs []TypedName) {
	if p.acceptNamedList(":objects") {
		objs = p.parseTypedListString(tokId)
		p.expect(tokClose)
	}
	return
}

func (p *parser) parseInit() (els []Formula) {
	p.expect(tokOpen)
	p.expectId(":init")
	for p.peek().typ == tokOpen {
		els = append(els, p.parseInitEl())
	}
	p.expect(tokClose)
	return
}

func (p *parser) parseInitEl() Formula {
	if p.acceptNamedList("=") {
		eq := &AssignNode{
			Op:   OpAssign,
			Lval: p.parseFhead(),
			Rval: p.expect(tokNum).text,
		}
		p.expect(tokClose)
		return eq
	}
	return p.parseLiteral()
}

func (p *parser) parseGoal() Formula {
	p.expect(tokOpen)
	p.expectId(":goal")
	g := p.parsePreGd()
	p.expect(tokClose)
	return g
}

func (p *parser) parseMetric() (m Metric) {
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

func (p *parser) parseTypedListString(typ tokenType) (lst []TypedName) {
	for {
		names := p.parseStrings(typ)
		if len(names) == 0 {
			break
		}
		typ := p.parseType()
		for _, n := range names {
			name := p.name(n)
			lst = append(lst, TypedName{Name: name, Types: typ})
		}
	}
	return lst
}

func (p *parser) parseType() (typ []Name) {
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
	return []Name{p.name(t.text)}
}

func (p *parser) parseStringPlus(typ tokenType) []string {
	lst := []string{p.expect(typ).text}
	lst = append(lst, p.parseStrings(typ)...)
	return lst
}

func (p *parser) parseStrings(typ tokenType) (lst []string) {
	for t, ok := p.accept(typ); ok; t, ok = p.accept(typ) {
		lst = append(lst, t.text)
	}
	return lst
}

func (p *parser) name(text string) Name {
	return MakeName(text, p.loc())
}
