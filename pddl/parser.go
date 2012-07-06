package pddl

import (
	"fmt"
	"io"
	"io/ioutil"
)

// A parser parses PDDL.
type parser struct {
	lex    *lexer
	peeks  [4]token
	npeeks int
}

// newParser returns a new parser that parses
// from the given io.Reader.
func newParser(file string, r io.Reader) (*parser, error) {
	text, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &parser{lex: newLexer(file, string(text))}, nil
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

func (p *parser) loc() Loc {
	return Loc{p.lex.name, p.lex.lineno}
}

// parseErrors are panicked by the parser
// when a parser error is encountered.
type parseError string

func (e parseError) Error() string {
	return string(e)
}

func (p *parser) errorf(format string, args ...interface{}) {
	panic(parseError(fmt.Sprintf("%s: %s", p.loc(), fmt.Sprintf(format, args...))))
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
