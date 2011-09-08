package main

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

type gdtype int

type gd struct {
	typ gdtype
	left *gd
	right *gd
	vr tname	// gdForall and gdExists
	name string	// gdPred
	parms []string	// gdPred
}

type effect int