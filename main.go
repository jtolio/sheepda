// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	skipAssignments = flag.Bool("a", false,
		"if provided, skip assignments when pretty-printing in parsed mode")
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	default:
		fmt.Fprintf(os.Stderr, "Usage: %s <parsed|output|result>\n", os.Args[0])
		return
	case "parsed", "output", "result":
	}
	prog, err := Parse(NewStream(os.Stdin))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	if flag.Arg(0) == "parsed" {
		if *skipAssignments {
			fmt.Println(prog.Expr)
		} else {
			fmt.Println(prog)
		}
		return
	}
	out := ioutil.Discard
	if flag.Arg(0) == "output" {
		out = os.Stdout
	}
	val, _, err := Eval(NewContext(out), NewScopeWithBuiltins(), prog)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	if flag.Arg(0) == "result" {
		fmt.Println(val)
	}
}
