package textfmt

import (
	"fmt"
	"io"
)

type Table struct {
	Rows            []*Row
	WrapSpec        []*WrapSpec
	BannerSpec      *WrapSpec
	ColumnSeparator string
}

func NewTable(defs ...*WrapSpec) *Table {
	t := &Table{
		WrapSpec:        defs,
		ColumnSeparator: " ",
		BannerSpec:      &WrapSpec{},
	}
	for i := range t.WrapSpec {
		t.WrapSpec[i].normalizeSpec()
	}
	t.BannerSpec.normalizeSpec()
	return t
}

func (t *Table) AddBanner(text string) error {
	row := &Row{
		RowType: RowTypeBanner,
		Columns: []*Block{NewBlock(text)},
	}
	t.Rows = append(t.Rows, row)
	return nil
}

func (t *Table) AddRow(cells ...string) error {
	r := &Row{
		RowType: RowTypeColumns,
		Columns: make([]*Block, len(cells)),
	}
	for i, cell := range cells {
		r.Columns[i] = NewBlock(cell)
	}
	t.Rows = append(t.Rows, r)
	return nil
}

func (t *Table) RenderTo(w io.Writer) error {
	// Helper function to calculate the maximum of two integers
	max := func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}

	// Create a copy of WrapSpecs for columns
	tempWrapSpecs := make([]*WrapSpec, len(t.WrapSpec))
	for i, spec := range t.WrapSpec {
		temp := *spec // Copy the WrapSpec
		tempWrapSpecs[i] = &temp
		tempWrapSpecs[i].ExactWidth = 0 // Initialize ExactWidth to 0
	}

	// Update the width of the temporary WrapSpecs for columns
	for _, row := range t.Rows {
		if row.RowType == RowTypeColumns {
			for i, block := range row.Columns {
				tmp := tempWrapSpecs[i]
				tmp.ExactWidth = max(block.Width(), tmp.ExactWidth)
			}
		}
	}

	totalWidth := 0
	for i, tmp := range tempWrapSpecs {
		if t.WrapSpec[i].ExactWidth == 0 {
			tmp.ExactWidth = min(max(tmp.ExactWidth, tmp.MinWidth), tmp.MaxWidth)
		} else {
			tmp.ExactWidth = t.WrapSpec[i].ExactWidth
		}
		totalWidth += tmp.ExactWidth
	}
	// Add space for column separators
	totalWidth += (len(tempWrapSpecs) - 1) * len(t.ColumnSeparator)

	// Create a copy of the WrapSpec for banners
	tempBannerSpec := *t.BannerSpec // Copy the WrapSpec
	tempBannerSpec.ExactWidth = totalWidth

	// Render each row
	for _, row := range t.Rows {
		var wrapSpecs []*WrapSpec
		if row.RowType == RowTypeColumns {
			wrapSpecs = tempWrapSpecs
		} else if row.RowType == RowTypeBanner {
			wrapSpecs = []*WrapSpec{&tempBannerSpec}
		}
		if err := row.RenderTo(w, wrapSpecs, t.ColumnSeparator); err != nil {
			return fmt.Errorf("error rendering row: %v", err)
		}
		w.Write([]byte("\n"))
	}

	// Ensure the function always returns a value
	return nil
}
