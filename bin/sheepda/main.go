// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jtolds/sheepda"
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
	prog, err := sheepda.Parse(sheepda.NewStream(os.Stdin))
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
	val, _, err := sheepda.Eval(
		sheepda.NewContext(out), sheepda.NewScopeWithBuiltins(), prog)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	if flag.Arg(0) == "result" {
		fmt.Println(val)
	}
}
