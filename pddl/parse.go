package pddl

import (
	"fmt"
)

type Parser struct {
	lex    *Lexer
	peeks  [2]token
	npeeks int
}

func (p *Parser) next() token {
	if p.npeeks == 0 {
		return p.lex.token()
	}
	t := p.peeks[0]
	for i := 1; i < p.npeeks; i++ {
		p.peeks[i-1] = p.peeks[i]
	}
	p.npeeks--
	return t
}

func Parse(lex *Lexer) *Parser {
	return &Parser{
		lex: lex,
	}
}

// peek at the nth token
func (p *Parser) peekn(n int) token {
	if n > len(p.peeks) {
		panic("Too much peeking in the Parser")
	}
	for ; p.npeeks < n; p.npeeks++ {
		p.peeks[p.npeeks] = p.lex.token()
	}
	return p.peeks[n-1]
}

func (p *Parser) peek() token {
	return p.peekn(1)
}

func (p *Parser) junk(n int) {
	for i := 0; i < n; i++ {
		p.next()
	}
}

func (p *Parser) accept(typ tokenType) (t token, ok bool) {
	if p.peek().typ == typ {
		t = p.next()
		ok = true
	}
	return
}

func (p *Parser) acceptNamedList(name string) bool {
	if p.peek().typ != tokOpen || p.peekn(2).txt != name {
		return false
	}
	p.junk(2)
	return true
}

func (p *Parser) errorf(format string, args ...interface{}) {
	pre := fmt.Sprintf("%s:%d", p.lex.name, p.lex.lineno)
	suf := fmt.Sprintf(format, args...)
	panic(fmt.Errorf("%s: %s", pre, suf))
}

func (p *Parser) expect(typ tokenType) token {
	t := p.peek()
	if t.typ != typ {
		p.errorf("expected %v, got %v", typ, t)
	}
	return p.next()
}

func (p *Parser) expectId(s string) token {
	t := p.peek()
	typ := tokId
	if s[0] == ':' {
		typ = tokCid
	} else if s[0] == '?' {
		typ = tokQid
	}
	if t.typ != typ || t.txt != s {
		p.errorf("expected identifier [\"%s\"], got %v", s, t)
	}
	return p.next()
}

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
	} else {
		fmt.Printf("No predicates\n")
	}
	return
}

func (p *Parser) parseAtomicFormSkele() Predicate {
	p.expect(tokOpen)
	pred := Predicate{
		Name:       p.expect(tokId).txt,
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
			act.Precondition = p.parsePreGd()
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

func (p *Parser) parsePreGd() (res Gd) {
	parseNested := func(p *Parser) Gd { return p.parsePreGd() }
	switch {
	case p.acceptNamedList("and"):
		res = p.parseAndGd(parseNested)
	case p.acceptNamedList("forall"):
		res = p.parseForallGd(parseNested)
	default:
		res = p.parsePrefGd()
	}
	return
}

func (p *Parser) parsePrefGd() Gd {
	return p.parseGd()
}

func (p *Parser) parseGd() (res Gd) {
	parseNested := func(p *Parser) Gd { return p.parseGd() }
	switch {
	case p.acceptNamedList("and"):
		res = p.parseAndGd(parseNested)
	case p.acceptNamedList("or"):
		res = p.parseOrGd(parseNested)
	case p.acceptNamedList("not"):
		res = GdNot{Expr: p.parseGd()}
		p.expect(tokClose)
	case p.acceptNamedList("imply"):
		res = GdOr{Left: GdNot{Expr: p.parseGd()}, Right: p.parseGd()}
		p.expect(tokClose)
	case p.acceptNamedList("exists"):
		res = p.parseExistsGd(parseNested)
	case p.acceptNamedList("forall"):
		res = p.parseForallGd(parseNested)
	default:
		res = GdLiteral(p.parseLiteral())
	}
	return
}

func (p *Parser) parseLiteral() Literal {
	pos := true
	if p.acceptNamedList("not") {
		pos = false
	}
	p.expect(tokOpen)
	res := Literal{
		Positive:   pos,
		Name:       p.expect(tokId).txt,
		Parameters: p.parseTerms(),
	}
	if !pos {
		p.expect(tokClose)
	}
	p.expect(tokClose)
	return res
}

func (p *Parser) parseTerms() (lst []string) {
	for {
		if t, ok := p.accept(tokId); ok {
			lst = append(lst, t.txt)
			continue
		}
		if t, ok := p.accept(tokQid); ok {
			lst = append(lst, t.txt)
			continue
		}
		break
	}
	return
}

func (p *Parser) parseAndGd(nested func(*Parser) Gd) Gd {
	conj := make([]Gd, 0)
	for p.peek().typ == tokOpen {
		conj = append(conj, nested(p))
	}
	res := seqAndGd(conj)
	p.expect(tokClose)
	return res
}

func seqAndGd(conj []Gd) (res Gd) {
	switch len(conj) {
	case 0:
		res = GdTrue(1)
	case 1:
		res = conj[0]
	default:
		res = GdAnd{Left: conj[0], Right: seqAndGd(conj[1:])}
	}
	return
}

func (p *Parser) parseOrGd(nested func(*Parser) Gd) Gd {
	disj := make([]Gd, 0)
	for p.peek().typ == tokOpen {
		disj = append(disj, nested(p))
	}
	res := seqOrGd(disj)

	p.expect(tokClose)
	return res
}

func seqOrGd(disj []Gd) (res Gd) {
	switch len(disj) {
	case 0:
		res = GdFalse(0)
	case 1:
		res = disj[0]
	default:
		res = GdOr{Left: disj[0], Right: seqOrGd(disj[1:])}
	}
	return
}

func (p *Parser) parseForallGd(nested func(*Parser) Gd) Gd {
	p.expect(tokOpen)
	vrs := p.parseTypedListString(tokQid)
	p.expect(tokClose)

	res := GdForall{}
	bottom := res
	for i, vr := range vrs {
		bottom.Variable = vr
		if i < len(vrs)-1 {
			bottom.Expr = GdForall{}
			bottom = bottom.Expr.(GdForall)
		}
	}

	bottom.Expr = nested(p)
	p.expect(tokClose)
	return res
}

func (p *Parser) parseExistsGd(nested func(*Parser) Gd) Gd {
	p.expect(tokOpen)
	vrs := p.parseTypedListString(tokQid)
	p.expect(tokClose)

	res := GdExists{}
	bottom := res
	for i, vr := range vrs {
		bottom.Variable = vr
		if i < len(vrs)-1 {
			bottom.Expr = GdExists{}
			bottom = bottom.Expr.(GdExists)
		}
	}

	bottom.Expr = nested(p)
	p.expect(tokClose)
	return res
}

func (p *Parser) parseEffect() Effect {
	if p.acceptNamedList("and") {
		parseNested := func(p *Parser) Effect {
			return p.parseCeffect()
		}
		return p.parseAndEffect(parseNested)
	}
	return p.parseCeffect()
}

func (p *Parser) parseAndEffect(nested func(*Parser) Effect) Effect {
	conj := make([]Effect, 0)
	for p.peek().typ == tokOpen {
		conj = append(conj, nested(p))
	}
	res := seqAndEffect(conj)
	p.expect(tokClose)
	return res
}

func seqAndEffect(conj []Effect) (res Effect) {
	switch len(conj) {
	case 0:
		res = EffNone(0)
	case 1:
		res = conj[0]
	default:
		res = EffAnd{Left: conj[0], Right: seqAndEffect(conj[1:])}
	}
	return
}

func (p *Parser) parseCeffect() (res Effect) {
	switch {
	case p.acceptNamedList("forall"):
		parseNested := func(p *Parser) Effect {
			return p.parseEffect()
		}
		res = p.parseForallEffect(parseNested)
	case p.acceptNamedList("when"):
		parseNested := func(p *Parser) Effect {
			return p.parseCondEffect()
		}
		res = p.parseWhen(parseNested)
	default:
		res = p.parsePeffect()
	}
	return
}

func (p *Parser) parseForallEffect(nested func(*Parser) Effect) Effect {
	p.expect(tokOpen)
	vrs := p.parseTypedListString(tokQid)
	p.expect(tokClose)

	res := EffForall{}
	bottom := res
	for i, vr := range vrs {
		bottom.Variable = vr
		if i < len(vrs)-1 {
			bottom.Effect = EffForall{}
			bottom = bottom.Effect.(EffForall)
		}
	}

	bottom.Effect = nested(p)
	p.expect(tokClose)
	return res
}

func (p *Parser) parseWhen(nested func(*Parser) Effect) Effect {
	res := EffWhen{
		Condition: p.parseGd(),
	}
	res.Effect = nested(p)
	p.expect(tokClose)
	return res
}

func (p *Parser) parsePeffect() Effect {
	if _, ok := assignOps[p.peekn(2).txt]; ok && p.peek().typ == tokOpen {
		return p.parseAssign()
	}
	return EffLiteral(p.parseLiteral())
}

func (p *Parser) parseAssign() Effect {
	p.expect(tokOpen)
	res := EffAssign{
		Op:   assignOps[p.expect(tokId).txt],
		Lval: p.parseFhead(),
		Rval: p.parseFexp(),
	}
	p.expect(tokClose)
	return res
}

func (p *Parser) parseCondEffect() Effect {
	if p.acceptNamedList("and") {
		parseNested := func(p *Parser) Effect {
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

func (p *Parser) ParseProblem() *problem {
	p.expect(tokOpen)
	p.expectId("define")
	prob := &problem{
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

func (p *Parser) parseInit() (els []InitEl) {
	p.expect(tokOpen)
	p.expectId(":init")
	for p.peek().typ == tokOpen {
		els = append(els, p.parseInitEl())
	}
	p.expect(tokClose)
	return
}

func (p *Parser) parseInitEl() InitEl {
	if p.acceptNamedList("=") {
		eq := InitEq{
			Lval: p.parseFhead(),
			Rval: p.expect(tokNum).txt,
		}
		p.expect(tokClose)
		return eq
	}
	return InitLiteral(p.parseLiteral())
}

func (p *Parser) parseGoal() Gd {
	p.expect(tokOpen)
	p.expectId(":goal")
	g := p.parsePreGd()
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
			lst = append(lst, TypedName{Name: n, Type: typ})
		}
	}
	return lst
}

func (p *Parser) parseType() []string {
	if _, ok := p.accept(tokMinus); !ok {
		return []string{"object"}
	}
	if _, ok := p.accept(tokOpen); ok {
		p.expectId("either")
		lst := p.parseStringPlus(tokId)
		p.expect(tokClose)
		return lst
	}
	t := p.expect(tokId)
	return []string{t.txt}
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
