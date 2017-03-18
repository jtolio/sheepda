package main

import (
	"fmt"
	"io"
)

type Context struct {
	out io.Writer
}

func NewContext(out io.Writer) *Context {
	return &Context{out: out}
}

type Value interface {
	String() string
}

type Closure struct {
	Scope  *Scope
	Lambda *LambdaExpr
}

func (c *Closure) String() string {
	return c.Lambda.String()
}

type Byte byte

func (b Byte) String() string {
	return fmt.Sprintf("byte(%x)", string(b))
}

type Builtin struct {
	Name      string
	Transform func(*Context, Value) (Value, error)
}

func (b *Builtin) String() string {
	return fmt.Sprintf("builtin(%s)", b.Name)
}

func nextByte(ctx *Context, v Value) (Value, error) {
	if t, ok := v.(Byte); ok {
		return Byte(t + 1), nil
	}
	return nil, fmt.Errorf("type %T is not a byte", v)
}

func printByte(ctx *Context, v Value) (Value, error) {
	if t, ok := v.(Byte); ok {
		_, err := fmt.Fprint(ctx.out, string(t))
		return v, err
	}
	return nil, fmt.Errorf("type %T is not a byte", v)
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
	return NewScope().
		Set("BYTE_NULL", Byte(0)).
		SetBuiltin("BYTE_NEXT", nextByte).
		SetBuiltin("BYTE_PRINT", printByte)
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
	t func(*Context, Value) (Value, error)) *Scope {
	return s.Set(name, &Builtin{Name: name, Transform: t})
}

func Eval(ctx *Context, s *Scope, expr Expr) (Value, error) {
trampoline:
	for {
		switch t := expr.(type) {
		case *ProgramExpr:
			expr = t.Expr
			continue trampoline
		case *VariableExpr:
			val := s.Get(t.Name)
			if val == nil {
				return nil, fmt.Errorf("invalid variable lookup: %s", t.Name)
			}
			return val, nil
		case *ApplicationExpr:
			fn, err := Eval(ctx, s, t.Func)
			if err != nil {
				return nil, err
			}
			arg, err := Eval(ctx, s, t.Arg)
			if err != nil {
				return nil, err
			}
			switch t := fn.(type) {
			case *Closure:
				s = t.Scope.Set(t.Lambda.Arg, arg)
				expr = t.Lambda.Body
				continue trampoline
			case *Builtin:
				return t.Transform(ctx, arg)
			default:
				return nil, fmt.Errorf("not callable")
			}
		case *LambdaExpr:
			return &Closure{Scope: s, Lambda: t}, nil
		default:
			return nil, fmt.Errorf("unknown expression")
		}
	}
}
