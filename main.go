package main

import (
	"fmt"
	"os"
	"io/ioutil"
)

func main() {
	s, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic("Error reading standard input")
	}

	l := lex("stdin", string(s))
	var t token
	for t = l.token(); t.typ != tokEof && t.typ != tokErr; t = l.token(){
		fmt.Printf("%v\n", t)
	}
	fmt.Printf("%v\n", t)
}