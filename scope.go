// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package main

import (
	"bytes"
	"fmt"
)

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
		_, err := fmt.Fprint(ctx.out, string(t))
		return v, false, err
	}
	return nil, false, fmt.Errorf("type %T is not a byte", v)
}

type Scope struct {
	Name   string
	Value  Value
	Parent *Scope
}

func NewScope() *Scope {
	return nil // deliberate
}

func NewScopeWithBuiltins() *Scope {
	expr, err := Parse(NewStream(bytes.NewReader([]byte(`
    \n.(\a.\b.b (print (n next null)) n)
  `))))
	if err != nil {
		// programmer screwed up the builtin above
		panic(err)
	}
	return NewScope().Set("PRINT_BYTE", NewClosure(
		NewScope().
			Set("null", Byte(0)).
			SetBuiltin("print", printByte).
			SetBuiltin("next", nextByte),
		expr.Expr.(*LambdaExpr)))
}

func (s *Scope) Get(name string) Value {
	if s == nil {
		return nil
	}
	if s.Name == name {
		return s.Value
	}
	return s.Parent.Get(name)
}

func (s *Scope) Set(name string, value Value) *Scope {
	return &Scope{Name: name, Value: value, Parent: s}
}

func (s *Scope) SetBuiltin(name string,
	t func(*Context, Value) (Value, bool, error)) *Scope {
	return s.Set(name, &Builtin{Name: name, Transform: t})
}
