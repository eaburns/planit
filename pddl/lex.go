package pddl

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenType int

const eof = -1
const whiteSpace = " \t\n\r"

const (
	tokEof   tokenType = eof
	tokOpen  tokenType = '('
	tokClose tokenType = ')'
	tokMinus tokenType = '-'
	tokEq    tokenType = '='
	tokErr   tokenType = iota + 255
	tokId
	tokQid
	tokCid
	tokNum
)

var (
	tokenTypeNames = map[tokenType]string{
		tokErr:   "error",
		tokOpen:  "'('",
		tokClose: "')'",
		tokMinus: "'-'",
		tokEq:    "'='",
		tokId:    "identifier",
		tokQid:   "?identifier",
		tokCid:   ":identifier",
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

type token struct {
	typ tokenType
	txt string
}

func (t token) String() string {

	if _, ok := runeToks[rune(t.typ)]; ok {
		return fmt.Sprintf("%v", t.typ)
	}
	if len(t.txt) > 10 {
		return fmt.Sprintf("%v [%10q...]", t.typ, t.txt)
	}
	return fmt.Sprintf("%v [%q]", t.typ, t.txt)
}

type Lexer struct {
	name   string
	txt    string
	start  int
	pos    int
	lineno int
	width  int
}

func Lex(name, txt string) *Lexer {
	return &Lexer{
		name:   name,
		txt:    txt,
		lineno: 1,
	}
}

func (l *Lexer) next() (r rune) {
	if l.pos >= len(l.txt) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.txt[l.pos:])
	l.pos += l.width
	if r == '\n' {
		l.lineno++
	}
	return
}

func (l *Lexer) backup() {
	if strings.HasPrefix(l.txt[l.pos-l.width:l.pos], "\n") {
		l.lineno--
	}
	l.pos -= l.width
}

func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *Lexer) junk() {
	l.start = l.pos
}

func (l *Lexer) accept(s string) bool {
	if strings.IndexRune(s, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptRun(s string) (any bool) {
	for acc := l.accept(s); acc; acc = l.accept(s) {
		any = true
	}
	return
}

func (l *Lexer) makeToken(t tokenType) token {
	tok := token{txt: l.txt[l.start:l.pos], typ: t}
	l.start = l.pos
	return tok
}

func (l *Lexer) errorf(format string, args ...interface{}) token {
	return token{typ: tokErr, txt: fmt.Sprintf(format, args...)}
}

func (l *Lexer) token() token {
	for {
		r := l.next()
		if typ, ok := runeToks[r]; ok {
			return l.makeToken(typ)
		}
		switch {
		case r == eof:
			return l.makeToken(eof)
		case unicode.IsSpace(r):
			l.junk()
			continue
		case r == ';':
			l.lexComment()
			continue
		case r == '_' || unicode.IsLetter(r):
			return l.lexIdent(tokId)
		case r == '?':
			return l.lexIdent(tokQid)
		case r == ':':
			return l.lexIdent(tokCid)
		case unicode.IsDigit(r):
			return l.lexNum()
		default:
			return l.errorf("unexpected token in input: %c", r)
		}
	}
	panic("Unreachable")
}

func (l *Lexer) lexIdent(t tokenType) token {
	r := l.next()
	for !unicode.IsSpace(r) && r != ')' {
		r = l.next()
	}
	l.backup()
	return l.makeToken(t)
}

func (l *Lexer) lexNum() token {
	digits := "0123456789"
	l.acceptRun(digits)
	l.accept(".")
	l.acceptRun(digits)
	l.accept("eE")
	l.accept("-")
	l.acceptRun(digits)
	return l.makeToken(tokNum)
}

func (l *Lexer) lexComment() {
	for t := l.next(); t != '\n'; t = l.next() {
	}
	l.junk()
}
