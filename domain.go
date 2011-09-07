package main

type domain struct {
	name string
	reqs []string
	types []tname
	consts []tname
	preds []pred
}

type tname struct {
	name string
	typ []string
}

type pred struct {
	name string
	parms []tname
}