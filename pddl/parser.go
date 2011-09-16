package pddl

import (
	"fmt"
	"log"
	. "goplan/lifted"
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

func (p *Parser) loc() Loc {
	return Loc{p.lex.name, p.lex.lineno}
}

func (p *Parser) errorf(format string, args ...interface{}) {
	log.Fatalf("%s: %s", p.loc(), fmt.Sprintf(format, args...))
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
