package textfmt

import (
	"regexp"
	"unicode"
)

var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

type Line struct {
	Text  string
	width int
}

func (f *Line) Width() int {
	if f.width > 0 {
		return f.width
	}

	scanner := newScanner(f.Text)
	width := 0

	for {
		tok := scanner.nextToken()
		if tok.Type == tokenEOF {
			break
		}
		switch tok.Type {
		default:
			width += tok.Width
		}
	}

	f.width = width
	return width
}

func (f *Line) StripColors() {
	f.Text = ansiEscapeRe.ReplaceAllString(f.Text, "")
	f.width = 0
}

func (f *Line) WordWrap(spec *WrapSpec) ([]string, error) {
	wrappedLines, err := spec.WordWrap(f.Text)
	if err != nil {
		return nil, err
	}
	return wrappedLines, nil
}

func isWideRune(r rune) bool {
	return unicode.In(r, unicode.Han, unicode.Hangul, unicode.Hiragana, unicode.Katakana)
}
