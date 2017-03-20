// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jtolds/sheepda"
)

var (
	skipAssignments = flag.Bool("a", false,
		"if provided, skip assignments when pretty-printing in parsed mode")
)

func main() {
	flag.Parse()

	if flag.Arg(0) != "parsed" && flag.Arg(0) != "output" &&
		flag.Arg(0) != "result" || flag.NArg() < 2 {
		_, _ = fmt.Fprintf(os.Stderr,
			"Usage: %s <parsed|output|result> <file1.shp> [<file2.shp> ...]\n",
			os.Args[0])
		os.Exit(-1)
	}

	handles := make([]io.Reader, 0, flag.NArg()-1)
	for i := 1; i < flag.NArg(); i++ {
		fh, err := os.Open(flag.Arg(i))
		if err != nil {
			panic(err)
		}
		defer fh.Close()
		handles = append(handles, fh)
	}
	source := io.MultiReader(handles...)

	prog, err := sheepda.Parse(sheepda.NewStream(source))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	if flag.Arg(0) == "parsed" {
		if *skipAssignments {
			_, err = fmt.Println(prog.Expr)
			if err != nil {
				panic(err)
			}
		} else {
			_, err = fmt.Println(prog)
			if err != nil {
				panic(err)
			}
		}
		return
	}

	var out io.Writer
	if flag.Arg(0) == "output" {
		out = os.Stdout
	}

	val, err := sheepda.Eval(
		sheepda.NewContext(out, os.Stdin), sheepda.NewScopeWithBuiltins(), prog)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	if flag.Arg(0) == "result" {
		_, err = fmt.Println(val)
		if err != nil {
			panic(err)
		}
	}
}
