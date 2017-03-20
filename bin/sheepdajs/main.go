// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package main

import (
	"bytes"
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/jtolds/sheepda"
)

func main() {
	js.Global.Set("sheepda", map[string]interface{}{
		"eval": Eval,
	})
}

func Eval(source string, mode string) (output string, failure string) {
	prog, err := sheepda.Parse(sheepda.NewStream(
		bytes.NewReader([]byte(source))))
	if err != nil {
		return "", err.Error()
	}

	if mode == "parse" {
		return fmt.Sprint(prog.Expr), ""
	}

	var out bytes.Buffer
	val, err := sheepda.Eval(sheepda.NewScopeWithBuiltins(&out, nil), prog)
	if err != nil {
		return "", err.Error()
	}

	if mode == "output" {
		return out.String(), ""
	}

	if mode == "result" {
		return val.String(), ""
	}

	return "", "undefined mode"
}
