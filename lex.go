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
	tokErr ttype = iota
	tokOpen
	tokClose
	tokMinus
	tokId
	tokQid
	tokCid
	tokNum
	tokStr
)

var (
	ttypeNames = map[ttype]string{
		tokOpen:  "Open",
		tokClose: "Close",
		tokMinus: "Minus",
		tokId:    "Id",
		tokQid:   "Qid",
		tokCid:   "Cid",
		tokNum:   "Num",
		tokStr:   "Str",
	}

	runeToks = map[int]ttype{
		'(': tokOpen,
		')': tokClose,
		'-': tokMinus,
	}
)

func (t ttype) String() string {
	return "tok" + ttypeNames[t]
}

type token struct {
	typ ttype
	txt string
	lno int
}

func (t token) String() string {
	if len(t.txt) > 10 {
		return fmt.Sprintf("line=%d: %v [%10q...]", t.lno, t.typ, t.txt)
	}
	return fmt.Sprintf("line=%d: %v [%q]", t.lno, t.typ, t.txt)
}

type stateFn func(*lexer) stateFn

type lexer struct {
	name  string
	txt   string
	start int
	pos   int
	lno   int
	width int
	state stateFn
	toks  chan token
}

func lex(name, txt string) *lexer {
	return &lexer{
		name:  name,
		txt:   txt,
		state: lexAny,
		lno:   1,
		toks:  make(chan token, 2),
	}
}

func (l *lexer) token() token {
	for {
		select {
		case t := <-l.toks:
			return t
		default:
			l.state = l.state(l)
		}
	}
	panic("Unreachable")
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

func (l *lexer) emit(t ttype) {
	l.toks <- token{txt: l.txt[l.start:l.pos], typ: t, lno: l.lno}
	l.start = l.pos
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.toks <- token{typ: tokErr, txt: fmt.Sprintf(format, args...)}
	return nil
}

func lexAny(l *lexer) stateFn {
	r := l.next()
	if ttype, ok := runeToks[r]; ok {
		l.emit(ttype)
		return lexAny
	}
	switch {
	case r == eof:
		l.emit(tokEof)
		return nil
	case unicode.IsSpace(r):
		l.junk()
		return lexAny
	case isAlpha(r):
		return lexId
	case r == '?':
		return lexQid
	case r == ':':
		return lexQid
	case r == ';':
		return lexComment
	case isNum(r):
		return lexNum
	}
	return l.errorf("unexpected token in input: %c", r)
}

func (l *lexer) ident(t ttype) {
	for isAlphaNum(l.next()) {
	}
	l.backup()
	l.emit(t)
}

func lexId(l *lexer) stateFn {
	l.ident(tokId)
	return lexAny
}

func lexQid(l *lexer) stateFn {
	l.ident(tokQid)
	return lexAny
}

func lexCid(l *lexer) stateFn {
	l.ident(tokCid)
	return lexAny
}

func lexNum(l *lexer) stateFn {
	digits := "0123456789"
	l.acceptRun(digits)
	l.accept(".")
	l.acceptRun(digits)
	l.accept("eE")
	l.accept("-")
	l.acceptRun(digits)
	l.emit(tokNum)
	return lexAny
}

func lexComment(l *lexer) stateFn {
	for t := l.next(); t != '\n'; t = l.next() {
	}
	l.junk()
	return lexAny
}

func isAlpha(r int) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isNum(r int) bool {
	return unicode.IsDigit(r)
}

func isAlphaNum(r int) bool {
	return isAlpha(r) || isNum(r)
}
