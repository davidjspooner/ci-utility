package textfmt

import (
	"math"
	"strings"
	"unicode"
)

type Align int

const (
	Left Align = iota
	Center
	Right
	Unpadded
)

type ColorControl int

const (
	NoColor ColorControl = iota
	AllowColor
)

type WrapSpec struct {
	ExactWidth         int
	MinWidth, MaxWidth int
	Align              Align
	Color              ColorControl
	PadChar            rune
}

func NewWrapSpec(width int, align Align, color ColorControl, padChar rune) *WrapSpec {
	w := WrapSpec{
		ExactWidth: width,
		Align:      align,
		Color:      color,
		PadChar:    padChar,
	}
	w.normalizeSpec()
	return &w
}

// WordWrap wraps raw input into aligned, color-aware, padded lines.
func (w *WrapSpec) WordWrap(text string) ([]string, error) {
	scanner := newScanner(text)
	var lines []string
	var word strings.Builder
	var wordWidth int
	activeSGR := newSGRState()
	lb := newLineBuilder(activeSGR, w)

	for {
		tok := scanner.nextToken()
		if tok.Type == tokenEOF {
			break
		}

		switch tok.Type {
		case tokenColor:
			codes, err := parseSGRCodes(tok.Value)
			if err != nil {
				return nil, err
			}
			if w.Color == AllowColor {
				for _, code := range codes {
					activeSGR.apply(code)
				}
				word.WriteString(tok.Value)
			}
		case tokenWhitespace:
			if word.Len() > 0 {
				if !lb.canFit(wordWidth) {
					lines = append(lines, lb.flushAsString())
				}
				lb.writeString(word.String(), wordWidth)
				word.Reset()
				wordWidth = 0
			}
			if !lb.canFit(1) {
				lines = append(lines, lb.flushAsString())
			}
			lb.writeSpace()
		case tokenOther:
			if tok.Width > w.ExactWidth {
				if word.Len() > 0 {
					if !lb.canFit(wordWidth) {
						lines = append(lines, lb.flushAsString())
					}
					lb.writeString(word.String(), wordWidth)
					word.Reset()
					wordWidth = 0
				}
				for _, r := range tok.Value {
					if !lb.canFit(1) {
						lines = append(lines, lb.flushAsString())
					}
					lb.writeRune(r, 1)
				}
			} else {
				word.WriteString(tok.Value)
				wordWidth += tok.Width
			}
		}
	}

	if word.Len() > 0 {
		if !lb.canFit(wordWidth) {
			lines = append(lines, lb.flushAsString())
		}
		lb.writeString(word.String(), wordWidth)
	}

	if lb.len() > 0 {
		lines = append(lines, lb.flushAsString())
	}

	return lines, nil
}

func isValidSGRCode(code int) bool {
	// Define a set of valid SGR codes based on the ANSI standard.
	validCodes := map[int]bool{
		0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true,
		8: true, 9: true, 30: true, 31: true, 32: true, 33: true, 34: true, 35: true,
		36: true, 37: true, 40: true, 41: true, 42: true, 43: true, 44: true, 45: true,
		46: true, 47: true, 90: true, 91: true, 92: true, 93: true, 94: true, 95: true,
		96: true, 97: true, 100: true, 101: true, 102: true, 103: true, 104: true,
		105: true, 106: true, 107: true,
	}
	return validCodes[code]
}

func (s *WrapSpec) normalizeSpec() {
	if s.PadChar == 0 {
		s.PadChar = ' '
	}
	if s.MaxWidth == 0 {
		s.MaxWidth = math.MaxInt
	}
}

func applyAlignment(s string, width int, align Align, padChar rune) string {

	switch align {
	case Left:
		s = strings.TrimRight(s, " ")
		s = strings.TrimRight(s, string(padChar))
	case Right, Center:
		s = strings.TrimSpace(s)
		s = strings.Trim(s, string(padChar))
	}

	plain := ansiEscapeRe.ReplaceAllString(s, "")

	actualWidth := 0
	for _, r := range plain {
		if r == 0 || unicode.Is(unicode.Mn, r) {
			continue
		}
		if isWideRune(r) {
			actualWidth += 2
		} else {
			actualWidth++
		}
	}
	padding := width - actualWidth
	if padding <= 0 {
		return s
	}

	switch align {
	case Left:
		return s + strings.Repeat(string(padChar), padding)
	case Right:
		return strings.Repeat(string(padChar), padding) + s
	case Center:
		left := padding / 2
		right := padding - left
		return strings.Repeat(string(padChar), left) + s + strings.Repeat(string(padChar), right)
	case Unpadded:
		return s // Return the string as-is without any padding or trimming
	default:
		return s
	}
}
