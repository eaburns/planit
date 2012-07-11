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

// expect expects the next tokens to match
// the arguments.  If the tokens match then
// they are returned and the error is nil, otherwise
// an error is returned.
//
// The arguments can be either tokenTypes or strings.
// A tokenType matches a token on its type, and a
// string matches either a tokQid, tokCid, or tokId
// if the first character of the string is a '?', ':', or
// anything else respectively.
func (p *parser) expect(vls ...interface{}) ([]token, error) {
	var toks []token
	for i := range vls {
		t := p.peek()
		switch v := vls[i].(type) {
		case tokenType:
			if t.typ != v {
				panic("foo")
				return nil, makeError(p, "expected %v, got %v", v, t.typ)
			}
		case string:
			var typ tokenType = tokId
			switch v[0] {
			case '?':
				typ = tokQid
			case ':':
				typ = tokCid
			}
			if t.typ != typ {
				return nil, makeError(p, "expected %v, got %v", typ, t.typ)
			}
			if t.text != v {
				return nil, makeError(p, "expected %s, get %s", v, t.text)
			}
		default:
			panic(fmt.Sprintf("Unsupported type in expect: %T (%#v)", vls[i], vls[i]))
		}
		toks = append(toks, p.next())
	}
	return toks, nil
}
