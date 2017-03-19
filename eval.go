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
	Scope   *Scope
	Lambda  *LambdaExpr
	memoize map[*Closure]Value
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
	Transform func(*Context, Value) (Value, bool, error)
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
	t func(*Context, Value) (Value, bool, error)) *Scope {
	return s.Set(name, &Builtin{Name: name, Transform: t})
}

func Eval(ctx *Context, s *Scope, expr Expr) (Value, bool, error) {
	switch t := expr.(type) {
	case *ProgramExpr:
		return Eval(ctx, s, t.Expr)
	case *VariableExpr:
		val := s.Get(t.Name)
		if val == nil {
			return nil, false, fmt.Errorf("invalid variable lookup: %s", t.Name)
		}
		return val, true, nil
	case *ApplicationExpr:
		fn, fnCacheable, err := Eval(ctx, s, t.Func)
		if err != nil {
			return nil, false, err
		}
		arg, argCacheable, err := Eval(ctx, s, t.Arg)
		if err != nil {
			return nil, false, err
		}
		switch c := fn.(type) {
		case *Closure:
			s = c.Scope.Set(c.Lambda.Arg, arg)
			expr = c.Lambda.Body
			ca, ok := arg.(*Closure)
			if !ok || !fnCacheable || !argCacheable {
				v, cacheable, err := Eval(ctx, s, expr)
				return v, cacheable && argCacheable && fnCacheable, err
			}
			if v, ok := c.memoize[ca]; ok {
				return v, true, nil
			}
			v, cacheable, err := Eval(ctx, s, expr)
			if err != nil {
				return nil, false, err
			}
			if cacheable {
				c.memoize[ca] = v
			}
			return v, cacheable, nil
		case *Builtin:
			return c.Transform(ctx, arg)
		default:
			return nil, false, fmt.Errorf("not callable")
		}
	case *LambdaExpr:
		return &Closure{Scope: s, Lambda: t, memoize: map[*Closure]Value{}},
			true, nil
	default:
		return nil, false, fmt.Errorf("unknown expression")
	}
}
