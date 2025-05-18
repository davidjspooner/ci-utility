package textfmt

import (
	"strings"
)

type lineBuilder struct {
	line     strings.Builder
	width    int
	sgr      *sgrState
	wrapSpec *WrapSpec
}

func newLineBuilder(sgr *sgrState, wrapSpec *WrapSpec) *lineBuilder {
	return &lineBuilder{
		sgr:      sgr.copy(),
		wrapSpec: wrapSpec,
	}
}

func (b *lineBuilder) reset() {
	b.line.Reset()
	b.width = 0
	if b.wrapSpec.Color == AllowColor && len(b.sgr.attrs) > 0 {
		b.line.WriteString(b.sgr.string())
	}
}

func (b *lineBuilder) flushAsString() string {
	text := b.line.String()
	text = strings.TrimRight(text, " \t")
	text = strings.TrimLeft(text, " ") // Remove surplus spaces at the start of a line
	result := applyAlignment(text, b.wrapSpec.ExactWidth, b.wrapSpec.Align, b.wrapSpec.PadChar)
	if b.wrapSpec.Color == AllowColor && len(b.sgr.attrs) > 0 {
		result += "\x1b[0m"
	}
	b.reset() //includes start of new line with sgr
	return result
}

func (b *lineBuilder) writeString(s string, w int) {
	b.line.WriteString(s)
	b.width += w
}

func (b *lineBuilder) writeRune(r rune, w int) {
	b.line.WriteRune(r)
	b.width += w
}

func (b *lineBuilder) writeSpace() {
	b.line.WriteRune(' ')
	b.width++
}

func (b *lineBuilder) canFit(w int) bool {
	return b.width+w <= b.wrapSpec.ExactWidth
}

func (b *lineBuilder) remaining() int {
	return b.wrapSpec.ExactWidth - b.width
}

func (b *lineBuilder) len() int {
	return b.line.Len()
}
