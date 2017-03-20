// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package sheepda

import (
	"fmt"
)

// Scope represents a bunch of defined variables due to argument application
// and function calling.
type Scope struct {
	Name   string
	Value  Value
	Parent *Scope
}

// NewScope returns an empty scope with nothing defined.
func NewScope() *Scope {
	return nil // deliberate
}

// Get will return the value in the current scope associated with name, or nil
// if no value is found.
func (s *Scope) Get(name string) Value {
	if s == nil {
		return nil
	}
	if s.Name == name {
		return s.Value
	}
	return s.Parent.Get(name)
}

// Set returns a new scope with all of the same values set as the previous
// scope, but will additionally have name set to value.
func (s *Scope) Set(name string, value Value) *Scope {
	return &Scope{Name: name, Value: value, Parent: s}
}

// Value represents an evaluated result. Anything can be a value! Eval only
// knows how to call *Closure or *Builtin types, so if you have something else
// as a value, don't use it like a function.
type Value interface {
	String() string
}

// A Closure represents an evaluated LambdaExpr. It has an associated Scope.
// Use NewClosure to create one.
type Closure struct {
	Scope   *Scope
	Lambda  *LambdaExpr
	memoize map[*Closure]Value
}

// NewClosure creates a Closure
func NewClosure(s *Scope, l *LambdaExpr) *Closure {
	return &Closure{Scope: s, Lambda: l, memoize: map[*Closure]Value{}}
}

func (c *Closure) String() string {
	return c.Lambda.String()
}

type resultMaps []func(Value, bool, error) (Value, bool, error)

func (r resultMaps) apply(v Value, c bool, e error) (Value, bool, error) {
	for i := len(r) - 1; i >= 0; i-- {
		v, c, e = r[i](v, c, e)
	}
	return v, c, e
}

func eval(s *Scope, expr Expr) (val Value, cacheable bool, err error) {
	var mapper resultMaps
trampoline:
	for {
		switch t := expr.(type) {
		case *ProgramExpr:
			expr = t.Expr
			continue trampoline
		case *VariableExpr:
			val := s.Get(t.Name)
			if val == nil {
				return mapper.apply(
					nil, false, fmt.Errorf("invalid variable lookup: %s", t.Name))
			}
			return mapper.apply(val, true, nil)
		case *ApplicationExpr:
			fn, fnCacheable, err := eval(s, t.Func)
			if err != nil {
				return mapper.apply(nil, false, err)
			}
			arg, argCacheable, err := eval(s, t.Arg)
			if err != nil {
				return mapper.apply(nil, false, err)
			}
			subCacheable := fnCacheable && argCacheable
			switch c := fn.(type) {
			case *Closure:
				s = c.Scope.Set(c.Lambda.Arg, arg)
				expr = c.Lambda.Body
				ca, ok := arg.(*Closure)
				if !ok {
					v, cacheable, err := eval(s, expr)
					return mapper.apply(v, cacheable && subCacheable, err)
				}
				if v, ok := c.memoize[ca]; ok {
					return mapper.apply(v, subCacheable, nil)
				}
				mapper = append(mapper,
					func(v Value, cacheable bool, err error) (Value, bool, error) {
						if err != nil {
							return nil, false, err
						}
						if cacheable {
							c.memoize[ca] = v
						}
						return v, cacheable && subCacheable, nil
					})
				continue trampoline
			case *Builtin:
				return mapper.apply(c.Transform(arg))
			default:
				return mapper.apply(nil, false, fmt.Errorf("not callable"))
			}
		case *LambdaExpr:
			return mapper.apply(NewClosure(s, t), true, nil)
		default:
			return mapper.apply(
				nil, false, fmt.Errorf("unknown expression: %T", expr))
		}
	}
}

// Eval evaluates an expression in the given scope and returns the resulting
// value.
func Eval(s *Scope, expr Expr) (val Value, err error) {
	val, _, err = eval(s, expr)
	return val, err
}
