package main

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

type Stream struct {
	data *bufio.Reader
	next *rune
	err  error
}

func NewStream(in io.Reader) *Stream {
	return &Stream{data: bufio.NewReader(in)}
}

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

func (s *Stream) Peek() (rune, error) {
	err := s.fillNext()
	if err != nil {
		return 0, err
	}
	return *s.next, nil
}

func (s *Stream) Next() {
	s.next = nil
}

func (s *Stream) Get() (r rune, err error) {
	r, err = s.Peek()
	s.Next()
	return r, err
}

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

func (s *Stream) AssertMatch(val rune) error {
	r, err := s.Get()
	if err != nil {
		return err
	}
	if r != val {
		return fmt.Errorf("unexpected rune. expected %#v, got %#v",
			string(val), string(r))
	}
	return s.SwallowWhitespace()
}
