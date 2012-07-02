package pddl

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	. "planit/prob"
)

// A parser parses PDDL.
type parser struct {
	lex    *lexer
	peeks  [2]token
	npeeks int
}

// next returns the next lexical token from the parser.
func (p *parser) next() token {
	if p.npeeks == 0 {
		return p.lex.token()
	}
	t := p.peeks[0]
	copy(p.peeks[:], p.peeks[1:])
	p.npeeks--
	return t
}

// newParserFile returns a new parser that parses
// a PDDL file.
func newParserFile(path string) (*parser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	text, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return &parser{lex: newLexer(path, string(text))}, nil
}

func (p *parser) loc() Loc {
	return Loc{p.lex.name, p.lex.lineno}
}

func (p *parser) errorf(format string, args ...interface{}) {
	log.Panicf("%s: %s", p.loc(), fmt.Sprintf(format, args...))
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
	if p.peek().typ != tokOpen || p.peekn(2).text != name {
		return false
	}
	p.junk(2)
	return true
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
	if t.typ != typ || t.text != s {
		p.errorf("expected identifier [\"%s\"], got %v", s, t)
	}
	return p.next()
}
