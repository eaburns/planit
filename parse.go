package main

import (
	"fmt"
)

type parser struct {
	lex    *lexer
	peeks  [2]token
	npeeks int
}

func (p *parser) next() token {
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

func parse(lex *lexer) *parser {
	return &parser{
		lex: lex,
	}
}

// peek at the nth token
func (p *parser) peekn(n int) token {
	if n > len(p.peeks) {
		panic("Too much peeking in the parser")
	}
	for ; p.npeeks < n; p.npeeks++ {
		p.peeks[p.npeeks] = p.lex.token()
	}
	return p.peeks[n-1]
}

func (p *parser) peek() token {
	return p.peekn(1)
}

func (p *parser) junk(n int) {
	for i := 0; i < n; i++ {
		p.next()
	}
}

func (p *parser) accept(typ tokenType) (t token, ok bool) {
	if p.peek().typ == typ {
		t = p.next()
		ok = true
	}
	return
}

func (p *parser) acceptNamedList(name string) bool {
	if p.peek().typ != tokOpen || p.peekn(2).txt != name {
		return false
	}
	p.junk(2)
	return true
}

func (p *parser) errorf(format string, args ...interface{}) {
	pre := fmt.Sprintf("%s:%d", p.lex.name, p.lex.lineno)
	suf := fmt.Sprintf(format, args...)
	panic(fmt.Errorf("%s: %s", pre, suf))
}

func (p *parser) expect(typ tokenType) token {
	t := p.peek()
	if t.typ != typ {
		p.errorf("expected %v, got %v", typ, t)
	}
	return p.next()
}

func (p *parser) expectId(s string) token {
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

func (p *parser) parseDomain() *domain {
	p.expect(tokOpen)
	p.expectId("define")
	d := &domain{
		name:   p.parseDomainName(),
		reqs:   p.parseReqsDef(),
		types:  p.parseTypesDef(),
		consts: p.parseConstsDef(),
		preds:  p.parsePredsDef(),
		acts:   make([]action, 0),
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
		d.acts = append(d.acts, p.parseActionDef())
	}

	p.expect(tokClose)
	return d
}

func (p *parser) parseDomainName() string {
	p.expect(tokOpen)
	p.expectId("domain")
	n := p.expect(tokId)
	p.expect(tokClose)
	return n.txt
}

func (p *parser) parseReqsDef() (reqs []string) {
	reqs = make([]string, 0)
	if !p.acceptNamedList(":requirements") {
		return
	}
	for t, ok := p.accept(tokCid); ok; t, ok = p.accept(tokCid) {
		reqs = append(reqs, t.txt)
	}
	p.expect(tokClose)
	return
}

func (p *parser) parseTypesDef() (types []typedName) {
	types = make([]typedName, 0)
	if !p.acceptNamedList(":types") {
		return
	}
	types = p.parseTypedListString(tokId)
	p.expect(tokClose)
	return
}

func (p *parser) parseConstsDef() (consts []typedName) {
	consts = make([]typedName, 0)
	if !p.acceptNamedList(":constants") {
		return
	}
	consts = p.parseTypedListString(tokId)
	p.expect(tokClose)
	return
}

func (p *parser) parsePredsDef() (preds []pred) {
	preds = make([]pred, 0)
	if !p.acceptNamedList(":predicates") {
		return
	}
	preds = append(preds, p.parseAtomicFormSkele())
	for p.peek().typ == tokOpen {
		preds = append(preds, p.parseAtomicFormSkele())
	}
	p.expect(tokClose)
	return
}

func (p *parser) parseAtomicFormSkele() pred {
	p.expect(tokOpen)
	pred := pred{
		name:  p.expect(tokId).txt,
		parms: p.parseTypedListString(tokQid),
	}
	p.expect(tokClose)
	return pred
}

func (p *parser) parseActionDef() action {
	p.expect(tokOpen)
	p.expectId(":action")

	act := action{name: p.expect(tokId).txt, parms: p.parseActParms()}

	if p.peek().txt == ":precondition" {
		p.junk(1)
		if p.peek().typ == tokOpen && p.peekn(2).typ == tokClose {
			p.junk(2)
		} else {
			act.prec = p.parsePreGd()
		}
	}
	if p.peek().txt == ":effect" {
		p.junk(1)
		if p.peek().typ == tokOpen && p.peekn(2).typ == tokClose {
			p.junk(2)
		} else {
			act.effect = p.parseEffect()
		}
	}

	p.expect(tokClose)
	return act
}

func (p *parser) parseActParms() []typedName {
	p.expectId(":parameters")
	p.expect(tokOpen)
	res := p.parseTypedListString(tokQid)
	p.expect(tokClose)
	return res
}

func (p *parser) parsePreGd() (res gd) {
	parseNested := func(p *parser) gd { return p.parsePreGd() }
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

func (p *parser) parsePrefGd() gd {
	return p.parseGd()
}

func (p *parser) parseGd() (res gd) {
	parseNested := func(p *parser) gd { return p.parseGd() }
	switch {
	case p.acceptNamedList("and"):
		res = p.parseAndGd(parseNested)
	case p.acceptNamedList("or"):
		res = p.parseOrGd(parseNested)
	case p.acceptNamedList("not"):
		res = gdNot{expr: p.parseGd()}
		p.expect(tokClose)
	case p.acceptNamedList("imply"):
		res = gdOr{left: gdNot{expr: p.parseGd()}, right: p.parseGd()}
		p.expect(tokClose)
	case p.acceptNamedList("exists"):
		res = p.parseExistsGd(parseNested)
	case p.acceptNamedList("forall"):
		res = p.parseForallGd(parseNested)
	default:
		p.expect(tokOpen)
		res = gdLiteral{
			pos:   true,
			name:  p.expect(tokId).txt,
			parms: p.parseTerms(),
		}
		p.expect(tokClose)
	}
	return
}

func (p *parser) parseAndGd(nested func(*parser) gd) gd {
	conj := make([]gd, 0)
	for p.peek().typ == tokOpen {
		conj = append(conj, nested(p))
	}
	res := seqAnd(conj)
	p.expect(tokClose)
	return res
}

func seqAnd(conj []gd) (res gd) {
	switch len(conj) {
	case 0:
		res = gdTrue(1)
	case 1:
		res = conj[0]
	default:
		res = gdAnd{left: conj[0], right: seqAnd(conj[1:])}
	}
	return
}

func (p *parser) parseOrGd(nested func(*parser) gd) gd {
	disj := make([]gd, 0)
	for p.peek().typ == tokOpen {
		disj = append(disj, nested(p))
	}
	res := seqOr(disj)

	p.expect(tokClose)
	return res
}

func seqOr(disj []gd) (res gd) {
	switch len(disj) {
	case 0:
		res = gdFalse(0)
	case 1:
		res = disj[0]
	default:
		res = gdOr{left: disj[0], right: seqOr(disj[1:])}
	}
	return
}

func (p *parser) parseForallGd(nested func(*parser) gd) gd {
	p.expect(tokOpen)
	vrs := p.parseTypedListString(tokQid)
	p.expect(tokClose)

	res := gdForall{}
	bottom := res
	for i, vr := range vrs {
		bottom.vr = vr
		if i < len(vrs)-1 {
			bottom.expr = gdForall{}
			bottom = bottom.expr.(gdForall)
		}
	}

	bottom.expr = nested(p)
	p.expect(tokClose)
	return res
}

func (p *parser) parseExistsGd(nested func(*parser) gd) gd {
	p.expect(tokOpen)
	vrs := p.parseTypedListString(tokQid)
	p.expect(tokClose)

	res := gdExists{}
	bottom := res
	for i, vr := range vrs {
		bottom.vr = vr
		if i < len(vrs)-1 {
			bottom.expr = gdExists{}
			bottom = bottom.expr.(gdExists)
		}
	}

	bottom.expr = nested(p)
	p.expect(tokClose)
	return res
}

func (p *parser) parseTerms() []string {
	lst := make([]string, 0)
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
	return lst

}

func (p *parser) parseEffect() *effect {
	// ignore
	p.expect(tokOpen)
	for nesting := 1; nesting > 0; {
		switch p.next().typ {
		case tokClose:
			nesting--
		case tokOpen:
			nesting++
		}
	}
	return nil
}

func (p *parser) parseTypedListString(typ tokenType) []typedName {
	lst := make([]typedName, 0)

	for {
		names := p.parseStrings(typ)
		if len(names) == 0 {
			break
		}
		typ := p.parseType()
		for _, n := range names {
			lst = append(lst, typedName{name: n, typ: typ})
		}
	}
	return lst
}

func (p *parser) parseType() []string {
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

func (p *parser) parseStringPlus(typ tokenType) []string {
	lst := []string{p.expect(typ).txt}
	lst = append(lst, p.parseStrings(typ)...)
	return lst
}

func (p *parser) parseStrings(typ tokenType) []string {
	lst := make([]string, 0)
	for t, ok := p.accept(typ); ok; t, ok = p.accept(typ) {
		lst = append(lst, t.txt)
	}
	return lst
}
