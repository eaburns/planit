package main

import "fmt"

type domain struct {
	name string
	reqs []string
	types []tname
	consts []tname
	preds []pred
	acts []action
}

type tname struct {
	name string
	typ []string
}

type pred struct {
	name string
	parms []tname
}

type action struct {
	name string
	parms []tname
	prec *gd
	effect *effect
}

type gdtype int

const (
	gdTrue gdtype = iota
	gdFalse
	gdAnd
	gdOr
	gdNot
	gdForall
	gdExists
	gdPred
)

var (
	gdNames = map[gdtype]string{
	gdTrue: "gdTrue",
	gdFalse: "gdFalse",
	gdAnd: "gdAnd",
	gdOr: "gdOr",
	gdNot: "gdNot",
	gdForall: "gdForall",
	gdExists: "gdExists",
	gdPred: "gdPred",
	}
)

func (t gdtype) String() string {
	return gdNames[t]
}

type gd struct {
	typ gdtype
	left *gd
	right *gd
	vr tname	// gdForall and gdExists
	name string	// gdPred
	parms []string	// gdPred
}

func (g *gd) String() string {
	s := fmt.Sprintf("{typ: %v", g.typ)
	switch g.typ {
	case gdAnd:
		s = fmt.Sprintf("%s, left: %v, right: %v}", s, g.left, g.right)
	case gdOr:
		s = fmt.Sprintf("%s, left: %v, right: %v}", s, g.left, g.right)
	case gdNot:
		s = fmt.Sprintf("%s, left: %v}", s, g.left)
	case gdForall:
		s = fmt.Sprintf("%s, vr: %v, left: %v}", s, g.vr, g.left)
	case gdExists:
		s = fmt.Sprintf("%s, vr: %v, left: %v}", s, g.vr, g.left)
	case gdPred:
		s = fmt.Sprintf("%s, name: %s, parms: %v}", s, g.name, g.parms)
	default:
		s += "}"
	}
	return s
}

type effect int