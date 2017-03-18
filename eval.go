package main

import (
	"fmt"
)

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
	Transform func(Value) (Value, error)
}

func (b *Builtin) String() string {
	return fmt.Sprintf("builtin(%s)", b.Name)
}

func nextByte(v Value) (Value, error) {
	switch t := v.(type) {
	case Byte:
		return Byte(t + 1), nil
	default:
		return nil, fmt.Errorf("type %T is not a byte", v)
	}
}

func printByte(v Value) (Value, error) {
	switch t := v.(type) {
	case Byte:
		_, err := fmt.Print(string(t))
		return v, err
	default:
		return nil, fmt.Errorf("type %T is not a byte", v)
	}
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
		Set("null", Byte(0)).
		SetBuiltin("next-byte", nextByte).
		SetBuiltin("print-byte", printByte)
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

func (s *Scope) SetBuiltin(name string, t func(Value) (Value, error)) *Scope {
	return s.Set(name, &Builtin{Name: name, Transform: t})
}

func Call(fn, arg Value) (Value, error) {
	switch t := fn.(type) {
	case *Closure:
		return Eval(t.Scope.Set(t.Lambda.Arg, arg), t.Lambda.Body)
	case *Builtin:
		return t.Transform(arg)
	default:
		return nil, fmt.Errorf("not callable")
	}
}

func Eval(s *Scope, expr Expr) (Value, error) {
	switch t := expr.(type) {
	case *VariableExpr:
		val := s.Get(t.Name)
		if val == nil {
			return nil, fmt.Errorf("invalid variable lookup: %s", t.Name)
		}
		return val, nil
	case *ApplicationExpr:
		fn, err := Eval(s, t.Func)
		if err != nil {
			return nil, err
		}
		arg, err := Eval(s, t.Arg)
		if err != nil {
			return nil, err
		}
		return Call(fn, arg)
	case *LambdaExpr:
		return &Closure{Scope: s, Lambda: t}, nil
	default:
		return nil, fmt.Errorf("unknown expression")
	}
}
