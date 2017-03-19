// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package sheepda

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type Context struct {
	out io.Writer

	in  *bufio.Reader
	err error
}

func NewContext(out io.Writer, in io.Reader) *Context {
	rv := &Context{out: out}
	if in != nil {
		rv.in = bufio.NewReader(in)
	}
	return rv
}

func (c *Context) readByte() (byte, error) {
	if c.err != nil {
		return 0, c.err
	}
	if c.in == nil {
		c.err = io.EOF
		return 0, io.EOF
	}
	b, err := c.in.ReadByte()
	c.err = err
	return b, err
}

type Byte byte

func (b Byte) String() string {
	return fmt.Sprintf("byte(%x)", string(b))
}

type Builtin struct {
	Name      string
	Transform func(*Context, Value) (val Value, cacheable bool, err error)
}

func (b *Builtin) String() string {
	return fmt.Sprintf("builtin(%s)", b.Name)
}

func nextByte(ctx *Context, v Value) (Value, bool, error) {
	if t, ok := v.(Byte); ok {
		return Byte(t + 1), true, nil
	}
	return nil, false, fmt.Errorf("type %T is not a byte", v)
}

func printByte(ctx *Context, v Value) (Value, bool, error) {
	if t, ok := v.(Byte); ok {
		if ctx.out != nil {
			_, err := fmt.Fprint(ctx.out, string(t))
			return v, false, err
		}
		return v, true, nil
	}
	return nil, false, fmt.Errorf("type %T is not a byte", v)
}

var (
	trueClosure = NewClosure(NewScope(), &LambdaExpr{
		Arg: "t",
		Body: &LambdaExpr{
			Arg:  "f",
			Body: &VariableExpr{Name: "t"}}})
	falseClosure = NewClosure(NewScope(), &LambdaExpr{
		Arg: "t",
		Body: &LambdaExpr{
			Arg:  "f",
			Body: &VariableExpr{Name: "f"}}})
)

func churchNumeral(val int) Value {
	var base Expr = &VariableExpr{Name: "x"}
	for i := 0; i < val; i++ {
		base = &ApplicationExpr{
			Func: &VariableExpr{Name: "f"},
			Arg:  base}
	}
	return NewClosure(NewScope(), &LambdaExpr{
		Arg: "f",
		Body: &LambdaExpr{
			Arg:  "x",
			Body: base}})
}

func churchPair(first, second Value) Value {
	return NewClosure(NewScope().
		Set("first", first).
		Set("second", second),
		&LambdaExpr{
			Arg: "p",
			Body: &ApplicationExpr{
				Func: &ApplicationExpr{
					Func: &VariableExpr{Name: "p"},
					Arg:  &VariableExpr{Name: "first"}},
				Arg: &VariableExpr{Name: "second"}}})
}

func readByte(ctx *Context, v Value) (Value, bool, error) {
	b, err := ctx.readByte()
	if err != nil {
		if err == io.EOF {
			return churchPair(falseClosure, churchNumeral(0)), true, nil
		}
		return nil, false, err
	}
	return churchPair(trueClosure, churchNumeral(int(b))), false, nil
}

func NewScopeWithBuiltins() *Scope {
	printExpr, err := Parse(NewStream(bytes.NewReader([]byte(
		`\n.(\_.\v.v (print (n next null)) n)`))))
	if err != nil {
		panic(err)
	}
	printVal := NewClosure(
		NewScope().
			Set("null", Byte(0)).
			SetBuiltin("print", printByte).
			SetBuiltin("next", nextByte),
		printExpr.Expr.(*LambdaExpr))

	readExpr, err := Parse(NewStream(bytes.NewReader([]byte(
		`\x.(read x)`))))
	if err != nil {
		panic(err)
	}
	readVal := NewClosure(
		NewScope().
			SetBuiltin("read", readByte),
		readExpr.Expr.(*LambdaExpr))

	return NewScope().
		Set("PRINT_BYTE", printVal).
		Set("READ_BYTE", readVal)
}

func (s *Scope) SetBuiltin(name string,
	t func(*Context, Value) (Value, bool, error)) *Scope {
	return s.Set(name, &Builtin{Name: name, Transform: t})
}
