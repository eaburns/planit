package main

import (
	"strings"
	"fmt"
	"utf8"
	"unicode"
)

type ttype int

const eof = -1
const whiteSpace = " \t\n\r"

const (
	tokEof       = eof
	tokOpen ttype = '('
	tokClose ttype = ')'
	tokMinus ttype = '-'
	tokErr ttype = iota + 255
	tokId
	tokQid
	tokCid
	tokNum
)

var (
	ttypeNames = map[ttype]string{
		tokErr: "error",
		tokOpen: "'('",
		tokClose: "')'",
		tokMinus: "'-'",
		tokId:    "identifier",
		tokQid:   "?identifier",
		tokCid:   ":identifier",
		tokNum:   "number",
	}

	runeToks = map[int]ttype{
		'(': tokOpen,
		')': tokClose,
		'-': tokMinus,
	}
)

func (t ttype) String() string {
	return ttypeNames[t]
}

type token struct {
	typ ttype
	txt string
	lno int
}

func (t token) String() string {
	
	if _, ok := runeToks[int(t.typ)]; ok  {
		return fmt.Sprintf("%v", t.typ)
	}
	if len(t.txt) > 10 {
		return fmt.Sprintf("%v [%10q...]", t.typ, t.txt)
	}
	return fmt.Sprintf("%v [%q]", t.typ, t.txt)
}

type lexer struct {
	name  string
	txt   string
	start int
	pos   int
	lno   int
	width int
}

func lex(name, txt string) *lexer {
	return &lexer{
		name:  name,
		txt:   txt,
		lno:   1,
	}
}

func (l *lexer) next() (rune int) {
	if l.pos >= len(l.txt) {
		l.width = 0
		return eof
	}
	rune, l.width = utf8.DecodeRuneInString(l.txt[l.pos:])
	l.pos += l.width
	if rune == '\n' {
		l.lno++
	}
	return rune
}

func (l *lexer) backup() {
	if strings.HasPrefix(l.txt[l.pos-l.width:l.pos], "\n") {
		l.lno--
	}
	l.pos -= l.width
}

func (l *lexer) peek() int {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) junk() {
	l.start = l.pos
}

func (l *lexer) accept(s string) bool {
	if strings.IndexRune(s, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(s string) (any bool) {
	for acc := l.accept(s); acc; acc = l.accept(s) {
		any = true
	}
	return
}

func (l *lexer) makeToken(t ttype) token {
	tok := token{txt: l.txt[l.start:l.pos], typ: t, lno: l.lno}
	l.start = l.pos
	return tok
}

func (l *lexer) errorf(format string, args ...interface{}) token {
	return token{typ: tokErr, txt: fmt.Sprintf(format, args...)}
}

func (l *lexer) token() token {
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
		case isAlpha(r):
			return l.lexIdent(tokId)
		case r == '?':
			return l.lexIdent(tokQid)
		case r == ':':
			return l.lexIdent(tokCid)
		case isNum(r):
			return l.lexNum()
		default:
			return l.errorf("unexpected token in input: %c", r)
		}
	}
	panic("Unreachable")
}

func (l *lexer) lexIdent(t ttype) token {
	for isIdRune(l.next()) {
	}
	l.backup()
	return l.makeToken(t)
}

func (l *lexer) lexNum() token {
	digits := "0123456789"
	l.acceptRun(digits)
	l.accept(".")
	l.acceptRun(digits)
	l.accept("eE")
	l.accept("-")
	l.acceptRun(digits)
	return l.makeToken(tokNum)
}

func (l *lexer) lexComment() {
	for t := l.next(); t != '\n'; t = l.next() {
	}
	l.junk()
}

func isAlpha(r int) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isNum(r int) bool {
	return unicode.IsDigit(r)
}

func isIdRune(r int) bool {
	return !unicode.IsSpace(r) && r != ')'
}
