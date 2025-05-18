package textfmt

import (
	"unicode"
	"unicode/utf8"
)

type tokenType int

const (
	tokenEOF        tokenType = iota
	tokenColor                // ANSI CSI sequences, like "\x1b[31m"
	tokenWhitespace           // Unicode whitespace
	tokenOther                // Anything else
)

type Token struct {
	Type  tokenType
	Value string
	Width int // 0 for tab and color, display width otherwise
}

type scanner struct {
	input string
	pos   int
	width int
}

func newScanner(input string) *scanner {
	return &scanner{input: input}
}

func (s *scanner) nextToken() Token {
	s.skipInvalid()

	if s.eof() {
		return Token{Type: tokenEOF}
	}

	start := s.pos
	r := s.next()

	if r == '\x1b' && s.peek() == '[' {
		s.next() // consume '['
		for {
			p := s.peek()
			if p == 0 {
				break
			}
			if ('A' <= p && p <= 'Z') || ('a' <= p && p <= 'z') {
				s.next()
				break
			}
			if ('0' <= p && p <= '9') || p == ';' || p == '?' {
				s.next()
				continue
			}
			break
		}
		value := s.input[start:s.pos]
		return Token{Type: tokenColor, Value: value, Width: 0}
	}

	if unicode.IsSpace(r) {
		for unicode.IsSpace(s.peek()) {
			s.next()
		}
		text := s.input[start:s.pos]
		return Token{Type: tokenWhitespace, Value: text, Width: displayWidth(text)}
	}

	for {
		p := s.peek()
		if p == 0 || p == '\\' || unicode.IsSpace(p) || p == '\x1b' {
			break
		}
		s.next()
	}
	text := s.input[start:s.pos]
	return Token{Type: tokenOther, Value: text, Width: displayWidth(text)}
}

func (s *scanner) skipInvalid() {
	for !s.eof() {
		r := s.peek()
		if r != utf8.RuneError {
			break
		}
		s.next()
	}
}

func (s *scanner) next() rune {
	if s.eof() {
		s.width = 0
		return 0
	}
	r, w := utf8.DecodeRuneInString(s.input[s.pos:])
	s.width = w
	s.pos += w
	return r
}

func (s *scanner) peek() rune {
	if s.eof() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(s.input[s.pos:])
	return r
}

func (s *scanner) eof() bool {
	return s.pos >= len(s.input)
}

func displayWidth(s string) int {
	width := 0
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		if r == utf8.RuneError || unicode.Is(unicode.Mn, r) {
			continue
		}
		width++
	}
	return width
}
