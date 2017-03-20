// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package sheepda

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type byteStream struct {
	in  *bufio.Reader
	err error
}

func (c *byteStream) readByte() (byte, error) {
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

type byteVal byte

func (b byteVal) String() string {
	return fmt.Sprintf("byte(%x)", string(b))
}

type Builtin struct {
	Name      string
	Transform func(Value) (val Value, cacheable bool, err error)
}

func (b *Builtin) String() string {
	return fmt.Sprintf("builtin(%s)", b.Name)
}

func nextByte(v Value) (Value, bool, error) {
	if t, ok := v.(byteVal); ok {
		return byteVal(t + 1), true, nil
	}
	return nil, false, fmt.Errorf("type %T is not a byte", v)
}

func printByte(out io.Writer, v Value) (Value, bool, error) {
	if t, ok := v.(byteVal); ok {
		if out != nil {
			_, err := fmt.Fprint(out, string(t))
			return v, false, err
		}
		return v, true, nil
	}
	return nil, false, fmt.Errorf("type %T is not a byte", v)
}

var (
	eofValue = ChurchPair(ChurchBool(false), ChurchNumeral(0))
)

func readByte(stream *byteStream, v Value) (Value, bool, error) {
	if stream == nil {
		return eofValue, true, nil
	}
	b, err := stream.readByte()
	if err != nil {
		if err == io.EOF {
			return eofValue, true, nil
		}
		return nil, false, err
	}
	return ChurchPair(ChurchBool(true), ChurchNumeral(uint(b))), false, nil
}

// NewScopeWithBuiltins creates an empty scope (like NewScope) but then defines
// two functions.
//
// PRINT_BYTE takes a Church-encoded numeral, turns it into the corresponding
// byte, then writes it to out, if out is not nil. The return value is the
// original numeral.
//
// READ_BYTE throws away its argument, then returns a Church-encoded pair where
// the first element is true if data was read and the second element is a
// Church-encoded numeral representing the byte that was read. The only reason
// the first element could be false is due to EOF. Read errors, like other
// errors with builtins, halt the interpreter.
func NewScopeWithBuiltins(out io.Writer, in io.Reader) *Scope {
	printExpr, err := Parse(NewStream(bytes.NewReader([]byte(
		`\n.(\_.\v.v (print (n next null)) n)`))))
	if err != nil {
		panic(err)
	}
	printVal := NewClosure(
		NewScope().
			Set("null", byteVal(0)).
			SetBuiltin("print", func(v Value) (Value, bool, error) {
				return printByte(out, v)
			}).
			SetBuiltin("next", nextByte),
		printExpr.Expr.(*LambdaExpr))

	var instream *byteStream
	if in != nil {
		instream = &byteStream{in: bufio.NewReader(in)}
	}
	readExpr, err := Parse(NewStream(bytes.NewReader([]byte(
		`\x.(read x)`))))
	if err != nil {
		panic(err)
	}
	readVal := NewClosure(
		NewScope().
			SetBuiltin("read", func(v Value) (Value, bool, error) {
				return readByte(instream, v)
			}),
		readExpr.Expr.(*LambdaExpr))

	return NewScope().
		Set("PRINT_BYTE", printVal).
		Set("READ_BYTE", readVal)
}

// SetBuiltin defines a builtin in the scope called name that applies fn to
// the given value. A result value should be returned. If cacheable is true,
// then the result of the call may get memoized and the function may never be
// called again. Cacheable should be false for functions with side-effects (or
// I/O). If err is non-nil, the interpreter will be halted.
func (s *Scope) SetBuiltin(name string,
	fn func(Value) (v Value, cacheable bool, err error)) *Scope {
	return s.Set(name, &Builtin{Name: name, Transform: fn})
}
