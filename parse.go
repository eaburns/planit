package main

import (
	"fmt"
)

type parser struct {
	lex *lexer
	peeks [2]token
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

func (p *parser) accept(typ ttype) (t token, ok bool) {
	if p.peek().typ == typ {
		t = p.next()
		ok = true
	}
	return
}

func (p *parser) junk(n int) {
	for i := 0; i < n; i++ {
		p.next()
	}
}
func (p *parser) errorf(name string, line int, format string, args ...interface{}) {
	pre := fmt.Sprintf("%s:%d", name, line)
	suf := fmt.Sprintf(format, args...)
	panic(fmt.Errorf("%s: %s", pre, suf))
}

func (p *parser) expect(typ ttype) token {
	t := p.peek()
	if t.typ != typ {
		p.errorf(p.lex.name, t.lno, "expected %v, got %v", typ, t)
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
	if t.typ != typ || t.txt != s{
		p.errorf(p.lex.name, t.lno, "expected identifier [\"%s\"], got %v", s, t)
	}
	return p.next()
}

func (p *parser) parseDomain() *domain {
	p.expect(tokOpen)
	p.expectId("define")
	d := &domain{
		name: p.parseDomainName(),
	}
	d.reqs = p.parseReqsDef()
	d.types = p.parseTypesDef()
	d.consts = p.parseConstsDef()
	d.preds = p.parsePredsDef()

	//p.expect(tokClose)
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

	if p.peek().typ != tokOpen || p.peekn(2).txt != ":requirements" {
		return
	}
	p.junk(2)
	for t, ok := p.accept(tokCid); ok; t, ok = p.accept(tokCid) {
		reqs = append(reqs, t.txt)
	}
	p.expect(tokClose)
	return
}

func (p *parser) parseTypesDef() (types []tname) {
	types = make([]tname, 0)

	if p.peek().typ != tokOpen || p.peekn(2).txt != ":types" {
		return
	}
	p.junk(2)
	types = p.parseTypedListString(tokId)
	p.expect(tokClose)
	return
}

func (p *parser) parseConstsDef() (consts []tname) {
	consts = make([]tname, 0)

	if p.peek().typ != tokOpen || p.peekn(2).txt != ":constants" {
		return
	}
	p.junk(2)
	consts = p.parseTypedListString(tokId)
	p.expect(tokClose)
	return
}

func (p *parser) parsePredsDef() (preds []pred) {
	preds = make([]pred, 0)

	if p.peek().typ != tokOpen || p.peekn(2).txt != ":predicates" {
		return
	}
	p.junk(2)
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
		name: p.expect(tokId).txt,
		parms: p.parseTypedListString(tokQid),
	}
	p.expect(tokClose)
	return pred
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

func (p *parser) parseTypedListString(typ ttype) []tname {
	lst := make([]tname, 0)

	for {
		names := p.parseStrings(typ)
		if len(names) == 0 {
			break
		}
		typ := p.parseType()
		for _, n := range names {
			lst = append(lst, tname{name: n, typ: typ})
		}
	}
	return lst
}

func (p *parser) parseStringPlus(typ ttype) []string {
	lst := []string{p.expect(typ).txt}
	lst = append(lst, p.parseStrings(typ)...)
	return lst
}

func (p *parser) parseStrings(typ ttype) []string {
	lst := make([]string, 0)
	for t, ok := p.accept(typ); ok; t, ok = p.accept(typ) {
		lst = append(lst, t.txt)
	}
	return lst
}