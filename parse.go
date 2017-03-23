// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package sheepda

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

var (
	Lambdas = map[rune]bool{
		'Λ': true, 'λ': true, 'ᴧ': true, 'Ⲗ': true, 'ⲗ': true, '𝚲': true,
		'𝛌': true, '𝛬': true, '𝜆': true, '𝜦': true, '𝝀': true, '𝝠': true,
		'𝝺': true, '𝞚': true, '𝞴': true, '\\': true,
	}
)

// IsVariableRune will return if the rune could be part of a variable name.
func IsVariableRune(ch rune) bool {
	return !unicode.IsSpace(ch) && ch != '(' && ch != ')' && ch != '.' &&
		ch != '=' && !Lambdas[ch]
}

// ParseVariable will parse a variable out of a stream. It assumes the stream
// has been advanced to the beginning of the variable.
func ParseVariable(s *Stream) (name string, err error) {
	for {
		ch, err := s.Peek()
		if err != nil {
			if err == io.EOF && name != "" {
				break
			}
			return "", err
		}
		if !IsVariableRune(ch) {
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

// Expr represents a parsed expression. Eval only knows how to deal with
// LambdaExprs, ApplicationExprs, VariableExprs, and ProgramExprs.
type Expr interface {
	String() string
}

// LambdaExpr represents a function definition.
type LambdaExpr struct {
	Arg  string
	Body Expr
}

func (e *LambdaExpr) String() string {
	return fmt.Sprintf("λ%s.%s", e.Arg, e.Body)
}

// ParseLambda parses a LambdaExpr out of a stream. It assumes the stream
// has been advanced to the beginning of the expression.
func ParseLambda(s *Stream) (*LambdaExpr, error) {
	err := s.AssertMatch(Lambdas)
	if err != nil {
		return nil, err
	}
	arg, err := ParseVariable(s)
	if err != nil {
		return nil, err
	}
	err = s.AssertMatch(map[rune]bool{'.': true})
	if err != nil {
		return nil, err
	}
	body, err := ParseExpr(s)
	if err != nil {
		return nil, err
	}
	return &LambdaExpr{Arg: arg, Body: body}, nil
}

// ApplicationExpr represents a function application
type ApplicationExpr struct {
	Func Expr
	Arg  Expr
}

func (e *ApplicationExpr) String() string {
	if l, ok := e.Func.(*LambdaExpr); ok {
		return fmt.Sprintf("((%s) %s)", l, e.Arg)
	}
	return fmt.Sprintf("(%s %s)", e.Func, e.Arg)
}

// ParseSubexpression will parse a parenthetical expression. If only one value
// is found in parentheses, it is simply an informational subexpression. If
// multiple values are found, a function application is assumed.
func ParseSubexpression(s *Stream) (Expr, error) {
	err := s.AssertMatch(map[rune]bool{'(': true})
	if err != nil {
		return nil, err
	}
	fn, err := ParseExpr(s)
	if err != nil {
		return nil, err
	}
	result := fn
	for {
		r, err := s.Peek()
		if err != nil {
			return nil, err
		}
		if r == ')' {
			s.Next()
			return result, s.SwallowWhitespace()
		}
		next, err := ParseExpr(s)
		if err != nil {
			return nil, err
		}
		result = &ApplicationExpr{Func: result, Arg: next}
	}
}

// VariableExpr represents a variable reference.
type VariableExpr struct {
	Name string
}

func (e *VariableExpr) String() string {
	return e.Name
}

// ParseExpr will parse the next full expression. It does not know how to
// handle assignment syntax sugar, nor does it make sure the stream has been
// completely processed.
func ParseExpr(s *Stream) (Expr, error) {
	r, err := s.Peek()
	if err != nil {
		return nil, err
	}

	if Lambdas[r] {
		return ParseLambda(s)
	}
	if r == '(' {
		return ParseSubexpression(s)
	}
	if IsVariableRune(r) {
		name, err := ParseVariable(s)
		return &VariableExpr{Name: name}, err
	}

	return nil, fmt.Errorf("expression not found")
}

type assignment struct {
	LHS string
	RHS Expr
}

// ProgramExpr represents a full program.
type ProgramExpr struct {
	Expr
}

// String will regenerate a list of newline-delimited assignments to place at
// the beginning, unlike ProgramExpr.Expr.String() which is otherwise
// equivalent.
func (e *ProgramExpr) String() string {
	var out bytes.Buffer
	expr := e.Expr
	applications := false
	for {
		if t, ok := expr.(*ApplicationExpr); ok {
			if fn, ok := t.Func.(*LambdaExpr); ok {
				fmt.Fprintf(&out, "%s = %s\n", fn.Arg, t.Arg)
				expr = fn.Body
				applications = true
				continue
			}
		}
		if applications {
			fmt.Fprintln(&out)
		}
		fmt.Fprint(&out, expr)
		return out.String()
	}
}

// Parse will parse a full lambda calculus program out of the stream. It
// understands assignment syntax sugar, such as
//
//   var = \x.\y.x
//
//   (do-something var)
//
func Parse(s *Stream) (*ProgramExpr, error) {
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
			return &ProgramExpr{Expr: expr}, nil
		}
		t, ok := expr.(*VariableExpr)
		if !ok {
			return nil, fmt.Errorf("unparsed code remaining")
		}
		err = s.AssertMatch(map[rune]bool{'=': true})
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
