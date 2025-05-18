package textfmt_test

import (
	"testing"

	"github.com/davidjspooner/ci-utility/pkg/textfmt"
)

func TestLineWidth(t *testing.T) {
	line := &textfmt.Line{Text: "Hello World"}
	width := line.Width()
	if width != 12 {
		t.Errorf("expected width 12, got %d", width)
	}
}
