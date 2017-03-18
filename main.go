package main

import (
	"fmt"
	"os"
)

func main() {
	expr, err := Parse(NewStream(os.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println("Parsed:", expr)
	val, err := Eval(NewScopeWithBuiltins(), expr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Evaluated:", val)
}
