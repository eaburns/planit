package pddl

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eof = -1
const whiteSpace = " \t\n\r"

// tokenType is the type tag for a scanned token.
type tokenType int

const (
	tokEof   tokenType = eof
	tokOpen  tokenType = '('
	tokClose tokenType = ')'
	tokErr   tokenType = iota + 255
	tokAtom
	tokQAtom
	tokCAtom
	tokNum
)

var (
	tokenTypeNames = map[tokenType]string{
		tokErr:   "error",
		tokOpen:  "'('",
		tokClose: "')'",
		tokAtom:    "atom",
		tokQAtom:   "?atom",
		tokCAtom:   ":atom",
		tokNum:   "number",
	}
)

func (t tokenType) String() string {
	return tokenTypeNames[t]
}

// token is a scanned token from a PDDL input.
type token struct {
	typ  tokenType
	text string
}

func (t token) String() string {
	if len(t.text) > 10 {
		return fmt.Sprintf("%v [%10q...]", t.typ, t.text)
	}
	return fmt.Sprintf("%v [%q]", t.typ, t.text)
}

// A lexer holds information and performs lexical
// analysis of a PDDL input.
type lexer struct {
	name   string
	text   string
	start  int
	pos    int
	lineno int
	width  int
}

// newLexer returns a new lexer that returns tokens
// for the given PDDL string.
func newLexer(name, text string) *lexer {
	return &lexer{
		name:   name,
		text:   text,
		lineno: 1,
	}
}

// next consumes and returns the next rune.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.text) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.text[l.pos:])
	l.pos += l.width
	if r == '\n' {
		l.lineno++
	}
	return
}

// backup returns the last rune that was scanned so that
// it will be returned by the next call to next().  backup
// can only be called once per call to next.
func (l *lexer) backup() {
	if strings.HasPrefix(l.text[l.pos-l.width:l.pos], "\n") {
		l.lineno--
	}
	l.pos -= l.width
}

// peek returns the next rune without consuming it.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// junk consumes the next rune.
func (l *lexer) junk() {
	l.start = l.pos
}

// accept returns true if the next rune is any of the
// runes in the given string.  If accept returns true
// then the rune is consumed.
func (l *lexer) accept(s string) bool {
	if strings.IndexRune(s, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun returns true if the next rune is any of
// the runse in the given string.  If acceptRun returns
// true then all of the next consecutive runes in the
// input that match a rune in the string are consumed.
func (l *lexer) acceptRun(s string) (any bool) {
	for acc := l.accept(s); acc; acc = l.accept(s) {
		any = true
	}
	return
}

// makeToken returns a token with the given type
// where the text is that between the start and current
// positions of the lexer.
func (l *lexer) makeToken(t tokenType) token {
	tok := token{text: l.text[l.start:l.pos], typ: t}
	l.start = l.pos
	return tok
}

// errorf returns a token of type tokErr with the text
// given by the format.
func (l *lexer) errorf(format string, args ...interface{}) token {
	return token{typ: tokErr, text: fmt.Sprintf(format, args...)}
}

// token returns the next token scanned from the PDDL.
func (l *lexer) token() token {
	for {
		r := l.next()
		switch {
		case r == eof:
			return l.makeToken(eof)
		case unicode.IsSpace(r):
			l.junk()
			continue
		case r == '(':
			return l.makeToken(tokOpen)
		case r == ')':
			return l.makeToken(tokClose)
		case r == ';':
			l.lexComment()
			continue
		default:
			return l.lexAtom()
		}
	}
	panic("Unreachable")
}

func (l *lexer) lexAtom() token {
	r := l.next()
	for !unicode.IsSpace(r) && r != ')' {
		r = l.next()
	}
	l.backup()
	tok := l.makeToken(tokAtom)
	if tok.text[0] == '?' {
		tok.typ = tokQAtom
	} else if tok.text[0] == ':' {
		tok.typ = tokCAtom
	} else if isNumber(tok.text) {
		tok.typ = tokNum
	}
	return tok
}

func isNumber(s string) bool {
	i := 0
	for ; i < len(s) && s[i] == '-'; i++ {
	}
	for ; i < len(s) && (s[i] == '.' || unicode.IsDigit(rune(s[i]))); i++ {
	}
	if i == 'e' || i == 'E' {
		i++
		for ; i < len(s) && s[i] == '-'; i++ {
		}
		for ; i < len(s) && unicode.IsDigit(rune(s[i])); i++ {
		}
	}
	return i == len(s)
}

func (l *lexer) lexComment() {
	for t := l.next(); t != '\n'; t = l.next() {
	}
	l.junk()
}
