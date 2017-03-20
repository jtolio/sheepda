// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package sheepda

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

// Stream makes parsing a sequence of runes easier
type Stream struct {
	data *bufio.Reader
	next *rune
	err  error
}

// NewStream creates a new stream from an io.Reader
func NewStream(in io.Reader) *Stream {
	return &Stream{data: bufio.NewReader(in)}
}

// EOF returns if the stream has ended.
func (s *Stream) EOF() bool { return s.err == io.EOF }

func (s *Stream) readRune() (rune, error) {
	r, _, err := s.data.ReadRune()
	if err != nil {
		return 0, err
	}
	if r == unicode.ReplacementChar {
		return 0, fmt.Errorf("invalid unicode")
	}
	return r, nil
}

func (s *Stream) fillNext() error {
	if s.next != nil {
		return nil
	}
	if s.err != nil {
		return s.err
	}
	r, err := s.readRune()
	if err != nil {
		s.err = err
		return err
	}
	if r == '#' {
		for {
			r, err := s.readRune()
			if err != nil {
				s.err = err
				return err
			}
			if r == '\n' {
				break
			}
		}
		return s.fillNext()
	}
	s.next = &r
	return nil
}

// Peek returns the next rune but does not pop it out of the stream.
func (s *Stream) Peek() (rune, error) {
	err := s.fillNext()
	if err != nil {
		return 0, err
	}
	return *s.next, nil
}

// Next pops any current rune out of the stream. It won't advance the stream
// farther though.
func (s *Stream) Next() {
	s.next = nil
}

// Get will advance the stream and return the next rune.
func (s *Stream) Get() (r rune, err error) {
	r, err = s.Peek()
	s.Next()
	return r, err
}

// SwallowWhitespace will advance the stream past any whitespace.
func (s *Stream) SwallowWhitespace() error {
	for {
		r, err := s.Peek()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !unicode.IsSpace(r) {
			return nil
		}
		s.Next()
	}
}

// AssertMatch will make sure the current rune is in the set of possible
// options and error otherwise. Then it will swallow any whitespace.
func (s *Stream) AssertMatch(options map[rune]bool) error {
	r, err := s.Get()
	if err != nil {
		return err
	}
	if !options[r] {
		return fmt.Errorf("unexpected rune. expected %#v, got %#v",
			options, string(r))
	}
	return s.SwallowWhitespace()
}
