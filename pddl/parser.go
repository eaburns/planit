package pddl

import (
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

func (p *parser) Loc() Location {
	return Location{p.lex.name, p.lex.lineno}
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

// acceptToken returns a token and true
// if the next token has the given type,
// otherwise it returns false.
func (p *parser) acceptToken(typ tokenType) (token, bool) {
	if p.peek().typ != typ {
		return token{}, false
	}
	return p.next(), true
}

// accept returns true if each upcoming token,
// matches the text of the corresponding
// parameter, in sequence.  Otherwise, accept
// returns false.
func (p *parser) accept(texts ...string) bool {
	if len(texts) > cap(p.peeks) {
		panic("too many peeks in accept")
	}
	for i := range texts {
		if p.peekn(i + 1).text != texts[i] {
			return false
		}
	}
	p.junk(len(texts))
	return true
}

// expectType returns the next token and no error
// if the next token has the specified type, otherwise
// an error is returned.
func (p *parser) expectType(typ tokenType) (token, error) {
	t := p.next()
	if t.typ != typ {
		return token{}, makeError(p, "expected %s, got %s", typ, t.typ)
	}
	return t, nil
}

// expectText returns the next token and no error
// if the next token has the specified text, otherwise
// an error is returned.
func (p *parser) expectText(text string) (token, error) {
	t := p.next()
	if t.text != text {
		return token{}, makeError(p, "expected %s, got %s", text, t.text)
	}
	return t, nil
}

// except is just like expectTokens except that it
// only accepts strings and it only returns the error
// value.
func (p *parser) expect(vls ...string) error {
	for i := range vls {
		t := p.next()
		if t.text != vls[i] {
			return makeError(p, "expected %s, got %s", vls[i], t)
		}
	}
	return nil
}
