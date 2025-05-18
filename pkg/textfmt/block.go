package textfmt

import (
	"strings"
	"unicode"
)

type Block struct {
	Lines []*Line
}

func NewBlock(input string) *Block {
	trimmed := strings.TrimSpace(input)
	lines := strings.Split(trimmed, "\\n")
	normalizedLines := make([]*Line, len(lines))
	for i, line := range lines {
		normalized := collapseBreakableSpaces(line)
		normalizedLines[i] = &Line{Text: normalized}
	}
	return &Block{Lines: normalizedLines}
}

func (b *Block) Width() int {
	w := 0
	for _, l := range b.Lines {
		w = max(w, l.Width())
	}
	return w
}

func (b *Block) WordWrap(spec *WrapSpec) ([]string, error) {
	var wrapped []string
	for _, l := range b.Lines {
		lines, err := l.WordWrap(spec)
		if err != nil {
			return nil, err
		}
		wrapped = append(wrapped, lines...)
	}
	return wrapped, nil
}

func collapseBreakableSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	spaceRun := false
	for _, r := range s {
		if unicode.Is(unicode.White_Space, r) {
			if !spaceRun {
				b.WriteRune(' ')
				spaceRun = true
			}
		} else {
			b.WriteRune(r)
			spaceRun = false
		}
	}
	return b.String()
}
