// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package sheepda

import (
	"fmt"
)

type Scope struct {
	Name   string
	Value  Value
	Parent *Scope
}

func NewScope() *Scope {
	return nil // deliberate
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

type Value interface {
	String() string
}

type Closure struct {
	Scope   *Scope
	Lambda  *LambdaExpr
	memoize map[*Closure]Value
}

func NewClosure(s *Scope, l *LambdaExpr) *Closure {
	return &Closure{Scope: s, Lambda: l, memoize: map[*Closure]Value{}}
}

func (c *Closure) String() string {
	return c.Lambda.String()
}

func Eval(ctx *Context, s *Scope, expr Expr) (
	val Value, cacheable bool, err error) {
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
		subCacheable := fnCacheable && argCacheable
		switch c := fn.(type) {
		case *Closure:
			s = c.Scope.Set(c.Lambda.Arg, arg)
			expr = c.Lambda.Body
			ca, ok := arg.(*Closure)
			if !ok {
				v, cacheable, err := Eval(ctx, s, expr)
				return v, cacheable && subCacheable, err
			}
			if v, ok := c.memoize[ca]; ok {
				return v, subCacheable, nil
			}
			v, cacheable, err := Eval(ctx, s, expr)
			if err != nil {
				return nil, false, err
			}
			if cacheable {
				c.memoize[ca] = v
			}
			return v, cacheable && subCacheable, nil
		case *Builtin:
			return c.Transform(ctx, arg)
		default:
			return nil, false, fmt.Errorf("not callable")
		}
	case *LambdaExpr:
		return NewClosure(s, t), true, nil
	default:
		return nil, false, fmt.Errorf("unknown expression: %T", expr)
	}
}
