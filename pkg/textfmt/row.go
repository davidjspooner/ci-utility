package textfmt

import (
	"fmt"
	"io"
	"strings"
)

type RowType int

const (
	RowTypeColumns RowType = iota
	RowTypeBanner
)

type Row struct {
	Columns []*Block
	RowType RowType
}

func NewRow(rowType RowType, cells ...string) *Row {
	r := &Row{
		RowType: rowType,
		Columns: make([]*Block, len(cells)),
	}
	for i := range cells {
		r.Columns[i] = NewBlock(cells[i])
	}
	return r
}

func (r *Row) RenderTo(w io.Writer, wrapSpecs []*WrapSpec, columnSeperator string) error {
	for i := range wrapSpecs {
		wrapSpecs[i].normalizeSpec()
	}
	if w == nil {
		return fmt.Errorf("writer is nil")
	}
	if len(r.Columns) != len(wrapSpecs) {
		return fmt.Errorf("number of cells (%d) does not match number of wrap specs (%d)", len(r.Columns), len(wrapSpecs))
	}
	wrapped := make([][]string, len(r.Columns))

	for i, cell := range r.Columns {
		lines, err := cell.WordWrap(wrapSpecs[i])
		if err != nil {
			return fmt.Errorf("error wrapping cell %d: %v", i, err)
		}
		wrapped[i] = lines
	}

	// Write the wrapped lines to the writer
	maxLines := 0
	for _, lines := range wrapped {
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for i := 0; i < maxLines; i++ {
		for j, lines := range wrapped {
			if i < len(lines) {
				fmt.Fprintf(w, "%-*s", wrapSpecs[j].ExactWidth, strings.TrimRight(lines[i], " "))
			} else {
				fmt.Fprintf(w, "%-*s", wrapSpecs[j].ExactWidth, "")
			}
			if j < len(wrapped)-1 {
				fmt.Fprint(w, columnSeperator)
			}
		}
		if i < maxLines-1 {
			fmt.Fprintln(w)
		}
	}

	return nil
}
