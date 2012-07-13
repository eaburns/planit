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

// acceptTokens consumes and returns a slice of
// tokens and true if the next tokens in the stream
// match the pattern elements (see token.match).
// Otherwise, nothing is consumed and false is
// returned.
func (p *parser) acceptTokens(vls ...interface{}) ([]token, bool) {
	if len(vls) > cap(p.peeks) {
		panic("too many peeks in accept")
	}
	var toks []token
	for i := range vls {
		if !p.peekn(i+1).matches(vls[i]) {
			return nil, false
		}
		toks = append(toks, p.peekn(i+1))
	}
	p.junk(len(vls))
	return toks, true
}

// accept is just like acceptTokens except that it
// only returns the boolean value.
func (p *parser) accept(vls ...interface{}) bool {
	_, ok := p.acceptTokens(vls...)
	return ok
}

// expectTokens expects the next tokens to match
// the arguments.  If the tokens match then
// they are returned and the error is nil, otherwise
// an error is returned.
//
// The arguments can be either tokenTypes or strings.
// A tokenType matches a token on its type, and a
// string matches on its text.
func (p *parser) expectTokens(vls ...interface{}) ([]token, error) {
	var toks []token
	for i := range vls {
		t := p.peek()
		if !p.peek().matches(vls[i]) {
			return nil, makeError(p, "expected %s, got %s", vls[i], t)
		}
		toks = append(toks, p.next())
	}
	return toks, nil
}

// except is just like expectTokens except that it
// only returns the error value.
func (p *parser) expect(vls ...interface{}) error {
	_, err := p.expectTokens(vls...)
	return err
}