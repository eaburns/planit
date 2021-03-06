// © 2013 the PlanIt Authors under the MIT license. See AUTHORS for the list of authors.

package pddl

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eof = -1
const whiteSpace = " \t\n\r"

// TokenType is the type tag for a scanned token.
type tokenType int

const (
	tokEof   tokenType = eof
	tokOpen  tokenType = '('
	tokClose tokenType = ')'
	tokMinus tokenType = '-'
	tokEq    tokenType = '='
	tokErr   tokenType = iota + 255
	tokName
	tokQname
	tokCname
	tokNum
)

var (
	tokenTypeNames = map[tokenType]string{
		tokErr:   "error",
		tokOpen:  "'('",
		tokClose: "')'",
		tokMinus: "'-'",
		tokEq:    "'='",
		tokName:  "name",
		tokQname: "?name",
		tokCname: ":name",
		tokNum:   "number",
	}

	runeToks = map[rune]tokenType{
		'(': tokOpen,
		')': tokClose,
		'-': tokMinus,
		'=': tokEq,
	}
)

func (t tokenType) String() string {
	return tokenTypeNames[t]
}

// Token is a scanned token from a PDDL input.
type token struct {
	typ  tokenType
	text string
}

func (t token) String() string {
	if _, ok := runeToks[rune(t.typ)]; ok {
		return fmt.Sprintf("%v", t.typ)
	}
	if len(t.text) > 10 {
		return fmt.Sprintf("%v [%10q...]", t.typ, t.text)
	}
	return fmt.Sprintf("%v [%q]", t.typ, t.text)
}

// A lexer holds information and performs lexical analysis of a PDDL input.
type lexer struct {
	name   string
	text   string
	start  int
	pos    int
	lineno int
	width  int
}

// NewLexer returns a new lexer that returns tokens for the given PDDL string.
func newLexer(name, text string) *lexer {
	return &lexer{
		name:   name,
		text:   text,
		lineno: 1,
	}
}

// Next consumes and returns the next rune.
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

// Backup puts the last rune that was scanned back in the scanner buffer so that it will
// be returned by the next call to next(). Backup can only be called once per call to next.
func (l *lexer) backup() {
	if strings.HasPrefix(l.text[l.pos-l.width:l.pos], "\n") {
		l.lineno--
	}
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) junk() {
	l.start = l.pos
}

// Accept returns true if the next rune is any of the runes in the given string.
// If accept returns true then the rune is consumed.
func (l *lexer) accept(s string) bool {
	if strings.IndexRune(s, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// AcceptRun returns true if the next rune is any of the runse in the given string.
// If acceptRun returns true then all of the next consecutive runes in the input
// that match a rune in the string are consumed.
func (l *lexer) acceptRun(s string) (any bool) {
	for acc := l.accept(s); acc; acc = l.accept(s) {
		any = true
	}
	return
}

// MakeToken returns a token with the given type where the text is that between the start
// and current positions of the lexer.
func (l *lexer) makeToken(t tokenType) token {
	tok := token{text: l.text[l.start:l.pos], typ: t}
	l.start = l.pos
	return tok
}

// Errorf returns a token of type tokErr with the text given by the format.
func (l *lexer) errorf(format string, args ...interface{}) token {
	return token{typ: tokErr, text: fmt.Sprintf(format, args...)}
}

// Token returns the next token scanned from the PDDL.
func (l *lexer) token() token {
	for {
		r := l.next()
		// '-' may be a number, not a minus beginning a type
		if r == '-' && (unicode.IsDigit(l.peek()) || l.peek() == '-') {
			return l.lexNum()
		}
		if typ, ok := runeToks[r]; ok {
			return l.makeToken(typ)
		}
		switch {
		case r == eof:
			return l.makeToken(eof)
		case unicode.IsSpace(r):
			l.junk()
		case r == ';':
			l.lexComment()
		case r == '?':
			return l.lexName(tokQname)
		case r == ':':
			return l.lexName(tokCname)
		case unicode.IsLetter(r):
			return l.lexName(tokName)
		case unicode.IsDigit(r):
			return l.lexNum()
		default:
			return l.errorf("unexpected token in input: %c", r)
		}
	}
}

// According to the PDDL 1.2 paper:
//	Names of domains, like other occurrences of syntactic category <name>, are strings of
//	characters beginning with a letter and containing letters, digits, hyphens (``-"), and
//	underscores (``_"). Case is not significant.
func (l *lexer) lexName(t tokenType) token {
	r := l.next()
	for unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
		r = l.next()
	}
	l.backup()
	return l.makeToken(t)
}

func (l *lexer) lexNum() token {
	l.acceptRun("-")
	digits := "0123456789"
	l.acceptRun(digits)
	l.accept(".")
	l.acceptRun(digits)
	if l.accept("eE") {
		l.accept("-")
		l.acceptRun(digits)
	}
	return l.makeToken(tokNum)
}

func (l *lexer) lexComment() {
	for t := l.next(); t != '\n'; t = l.next() {
	}
	l.junk()
}
