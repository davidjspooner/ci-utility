package textfmt_test

import (
	"strings"
	"testing"

	"github.com/davidjspooner/ci-utility/pkg/textfmt"
)

type testCase struct {
	name            string
	wrapspec        []*textfmt.WrapSpec
	columnSeperator string
	inputCells      []string
	expectedLines   []string
	expectedError   string
}

func (tc *testCase) runTest(t *testing.T) {
	//make a row of cells
	row := textfmt.NewRow(textfmt.RowTypeColumns, tc.inputCells...)
	b := strings.Builder{}
	err := row.RenderTo(&b, tc.wrapspec, " | ")
	if err != nil {
		if tc.expectedError == "" {
			t.Errorf("%s: unexpected error:\n  %v", tc.name, err)
		} else if err.Error() != tc.expectedError {
			t.Errorf("%s: expected error\n   %q\ngot\n   %q", tc.name, tc.expectedError, err.Error())
		}
		return
	} else if tc.expectedError != "" {
		t.Errorf("%s: expected error\n  %q\n  got nil", tc.name, tc.expectedError)
		return
	}
	gotLines := strings.Split(b.String(), "\n")
	overlap := min(len(gotLines), len(tc.expectedLines))
	for i := 0; i < overlap; i++ {
		if gotLines[i] != tc.expectedLines[i] {
			t.Errorf("%s: line #%d\n  expected %q\n  got      %q", tc.name, i, tc.expectedLines[i], gotLines[i])
		}
	}
	if overlap < len(tc.expectedLines) {
		//show extra lines in expected output
		for i := overlap; i < len(tc.expectedLines); i++ {
			t.Errorf("%s: expected extra line #%d:\n  %q", tc.name, i, tc.expectedLines[i])
		}
	}
	if overlap < len(gotLines) {
		//show extra lines in actual output
		for i := overlap; i < len(gotLines); i++ {
			t.Errorf("%s: got extra line #%d\n   %q", tc.name, i, gotLines[i])
		}
	}
}

func TestSimpleRowRendering(t *testing.T) {
	tc := &testCase{
		name: "Basic row rendering",
		wrapspec: []*textfmt.WrapSpec{
			{ExactWidth: 20, Align: textfmt.Left, PadChar: ' '},
			{ExactWidth: 16, Align: textfmt.Center, PadChar: ' '},
			{ExactWidth: 16, Align: textfmt.Right, PadChar: ' '},
		},
		columnSeperator: " | ",
		inputCells:      []string{"Column1 is very very long", "Column2", "Column3"},
		expectedLines: []string{
			"Column1 is very very |     Column2      |          Column3",
			"long                 |                  |                 ",
		},
		expectedError: "",
	}
	tc.runTest(t)
}

func TestMultilineRowRendering(t *testing.T) {
	tc := &testCase{
		name: "Multiline row rendering",
		wrapspec: []*textfmt.WrapSpec{
			{ExactWidth: 20, Align: textfmt.Left, PadChar: ' '},
			{ExactWidth: 16, Align: textfmt.Center, PadChar: ' '},
			{ExactWidth: 16, Align: textfmt.Right, PadChar: ' '},
		},
		columnSeperator: " | ",
		inputCells:      []string{"Column1 is very very long.\\nIt has an explicit line break", "Column2", "Column3"},
		expectedLines: []string{
			"Column1 is very very |     Column2      |          Column3",
			"long.                |                  |                 ",
			"It has an explicit   |                  |                 ",
			"line break           |                  |                 ",
		},
		expectedError: "",
	}
	tc.runTest(t)
}
