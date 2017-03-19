// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

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
		return NewClosure(s, t), true, nil
	default:
		return nil, false, fmt.Errorf("unknown expression: %T", expr)
	}
}
