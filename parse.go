package main

import (
	"fmt"
	"io"
	"unicode"
)

func isVariableRune(ch rune) bool {
	return !unicode.IsSpace(ch) &&
		ch != '\\' && ch != '(' && ch != ')' && ch != '.'
}

func ParseVariable(s *Stream) (name string, err error) {
	for {
		ch, err := s.Peek()
		if err != nil {
			if err == io.EOF && name != "" {
				break
			}
			return "", err
		}
		if !isVariableRune(ch) {
			break
		}
		name += string(ch)
		s.Next()
	}
	if name == "" {
		return "", fmt.Errorf("variable expected, not found")
	}
	return name, s.SwallowWhitespace()
}

type Expr interface {
	String() string
}

type LambdaExpr struct {
	Arg  string
	Body Expr
}

func (e *LambdaExpr) String() string {
	return fmt.Sprintf("\\%s.%s", e.Arg, e.Body)
}

func ParseLambda(s *Stream) (Expr, error) {
	err := s.AssertMatch('\\')
	if err != nil {
		return nil, err
	}
	arg, err := ParseVariable(s)
	if err != nil {
		return nil, err
	}
	err = s.AssertMatch('.')
	if err != nil {
		return nil, err
	}
	body, err := ParseExpr(s)
	if err != nil {
		return nil, err
	}
	return &LambdaExpr{Arg: arg, Body: body}, nil
}

type ApplicationExpr struct {
	Func Expr
	Arg  Expr
}

func (e *ApplicationExpr) String() string {
	return fmt.Sprintf("(%s %s)", e.Func, e.Arg)
}

func ParseApplication(s *Stream) (Expr, error) {
	err := s.AssertMatch('(')
	if err != nil {
		return nil, err
	}
	fn, err := ParseExpr(s)
	if err != nil {
		return nil, err
	}
	arg, err := ParseExpr(s)
	if err != nil {
		return nil, err
	}
	err = s.AssertMatch(')')
	if err != nil {
		return nil, err
	}
	return &ApplicationExpr{Func: fn, Arg: arg}, nil
}

type VariableExpr struct {
	Name string
}

func (e *VariableExpr) String() string {
	return e.Name
}

func ParseExpr(s *Stream) (Expr, error) {
	r, err := s.Peek()
	if err != nil {
		return nil, err
	}

	if r == '\\' {
		return ParseLambda(s)
	}
	if r == '(' {
		return ParseApplication(s)
	}
	if isVariableRune(r) {
		name, err := ParseVariable(s)
		return &VariableExpr{Name: name}, err
	}

	return nil, fmt.Errorf("expression not found")
}

type assignment struct {
	LHS string
	RHS Expr
}

func Parse(s *Stream) (Expr, error) {
	err := s.SwallowWhitespace()
	if err != nil {
		return nil, err
	}
	var assignments []assignment
	for {
		expr, err := ParseExpr(s)
		if err != nil {
			return nil, err
		}
		if s.EOF() {
			for i := len(assignments) - 1; i >= 0; i-- {
				expr = &ApplicationExpr{
					Func: &LambdaExpr{Arg: assignments[i].LHS, Body: expr},
					Arg:  assignments[i].RHS,
				}
			}
			return expr, nil
		}
		switch t := expr.(type) {
		default:
			return nil, fmt.Errorf("unparsed code remaining")
		case *VariableExpr:
			err = s.AssertMatch('=')
			if err != nil {
				return nil, err
			}
			rhs, err := ParseExpr(s)
			if err != nil {
				return nil, err
			}
			assignments = append(assignments, assignment{LHS: t.Name, RHS: rhs})
		}
	}
}
