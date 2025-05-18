package textfmt_test

import (
	"testing"

	"github.com/davidjspooner/ci-utility/pkg/textfmt"
)

// GenericTestFunction is a reusable test function that takes a testing object, test name, input, and expected output.
func GenericWrapTestFunction(t *testing.T, testName string, wrapspec *textfmt.WrapSpec, input string, expectedOutput []string, expectedError string) {
	output, err := wrapspec.WordWrap(input)
	overlap := min(len(output), len(expectedOutput))
	for i := 0; i < overlap; i++ {
		if output[i] != expectedOutput[i] {
			t.Errorf("%s: line #%d\n  expected %q\n  got      %q", testName, i, expectedOutput[i], output[i])
		}
	}
	if overlap < len(expectedOutput) {
		//show extra lines in expected output
		for i := overlap; i < len(expectedOutput); i++ {
			t.Errorf("%s: expected extra line #%d:\n  %q", testName, i, expectedOutput[i])
		}
	}
	if overlap < len(output) {
		//show extra lines in actual output
		for i := overlap; i < len(output); i++ {
			t.Errorf("%s: got extra line #%d\n   %q", testName, i, output[i])
		}
	}
	if err != nil {
		if expectedError == "" {
			t.Errorf("%s: unexpected error:\n  %v", testName, err)
		} else if err.Error() != expectedError {
			t.Errorf("%s: expected error\n   %q\ngot\n   %q", testName, expectedError, err.Error())
		}
		return
	} else if expectedError != "" {
		t.Errorf("%s: expected error\n  %q\n  got nil", testName, expectedError)
		return
	}
}

func TestRealTabsAndNewlines(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "Hello\tWorld\nThis is\ta test"
	expected := []string{"Hello     ", "World This", "is a test "}
	GenericWrapTestFunction(t, "Real tabs and newlines treated as spaces", wrapspec, input, expected, "")
}

func TestRunsOfSpaces(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "Hello    World  This   is  a    test"
	expected := []string{"Hello     ", "World This", "is a test "}
	GenericWrapTestFunction(t, "Runs of spaces replaced with single space", wrapspec, input, expected, "")
}

func TestLinesSplitByNewline(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "Hello World\nThis is a test"
	expected := []string{"Hello     ", "World This", "is a test "}
	GenericWrapTestFunction(t, "Lines split by newline", wrapspec, input, expected, "")
}

func TestColorCodesZeroWidth(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "\u001b[31mHello\u001b[0m World"
	expected := []string{"Hello     ", "World     "}
	GenericWrapTestFunction(t, "Color codes treated as zero width", wrapspec, input, expected, "")
}

func TestWideUnicodeCharacters(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "你好世界 Hello"
	expected := []string{"你好世界 Hello"}
	GenericWrapTestFunction(t, "Wide unicode characters measured correctly", wrapspec, input, expected, "")
}

func TestAlignCenter(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Center, PadChar: ' '}
	input := "Centered text example"
	expected := []string{" Centered ", "   text   ", " example  "}
	GenericWrapTestFunction(t, "Text aligned to center", wrapspec, input, expected, "")
}

func TestAlignRight(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Right, PadChar: ' '}
	input := "Right aligned text"
	expected := []string{"     Right", "   aligned", "      text"}
	GenericWrapTestFunction(t, "Text aligned to right", wrapspec, input, expected, "")
}

func TestAlternatePadCharacter(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Center, PadChar: '*'}
	input := "Padded text"
	expected := []string{"**Padded**", "***text***"}
	GenericWrapTestFunction(t, "Text padded with alternate character", wrapspec, input, expected, "")
}

func TestOnlySpacesOrTabs(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "\t\t    \t"
	expected := []string{"          "}
	GenericWrapTestFunction(t, "Input with only spaces or tabs", wrapspec, input, expected, "")
}

func TestEmptyString(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := ""
	expected := []string{}
	GenericWrapTestFunction(t, "Input with empty string", wrapspec, input, expected, "")
}

func TestValidSGRCodes(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "\u001b[31mRed\u001b[0m Text"
	expected := []string{"Red Text  "}
	GenericWrapTestFunction(t, "Valid SGR codes", wrapspec, input, expected, "")
}

func TestInvalidSGRCodes(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "\u001b[99mInvalid\u001b[0m Code"
	expected := []string{}
	GenericWrapTestFunction(t, "Invalid SGR codes", wrapspec, input, expected, "invalid SGR code")
}

func TestStringMethodSGR(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "\u001b[1mBold\u001b[0m Text"
	expected := []string{"Bold Text "}
	GenericWrapTestFunction(t, "String method for SGR", wrapspec, input, expected, "")
}

func TestAlignmentVaryingWidths(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 15, Align: textfmt.Center, PadChar: ' '}
	input := "Center Align"
	expected := []string{" Center Align  "}
	GenericWrapTestFunction(t, "Center alignment with varying width", wrapspec, input, expected, "")

	wrapspec = &textfmt.WrapSpec{ExactWidth: 15, Align: textfmt.Right, PadChar: ' '}
	expected = []string{"   Center Align"}
	GenericWrapTestFunction(t, "Right alignment with varying width", wrapspec, input, expected, "")
}

func TestLongWordsExceedingWidth(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "Supercalifragilisticexpialidocious"
	expected := []string{"Supercalif", "ragilistic", "expialidoc", "ious      "}
	GenericWrapTestFunction(t, "Long words exceeding width", wrapspec, input, expected, "")
}

func TestMultipleNewlinesAndTabs(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' '}
	input := "Line1\n\tLine2\n\t\tLine3"
	expected := []string{"Line1     ", "Line2     ", "Line3     "}
	GenericWrapTestFunction(t, "Multiple newlines and tabs", wrapspec, input, expected, "")
}

func TestEmbeddedColorsAllowColorTrue(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 10, Align: textfmt.Left, PadChar: ' ', Color: textfmt.AllowColor}
	input := "\u001b[31mRed\u001b[0m and \u001b[32mGreen\u001b[0m"
	expected := []string{"\u001b[31mRed\u001b[0m and   ", "\u001b[32mGreen\u001b[0m     "}
	GenericWrapTestFunction(t, "Embedded colors with AllowColor true", wrapspec, input, expected, "")
}

func TestEmbeddedColorsAllowColorTrueMultipleLines(t *testing.T) {
	wrapspec := &textfmt.WrapSpec{ExactWidth: 15, Align: textfmt.Left, PadChar: ' ', Color: textfmt.AllowColor}
	input := "\u001b[31mRed\u001b[0m and \u001b[32mGreen\u001b[0m\n\u001b[34mBlue\u001b[0m"
	expected := []string{"\u001b[31mRed\u001b[0m and \u001b[32mGreen\u001b[0m  ", "\u001b[34mBlue\u001b[0m           "}
	GenericWrapTestFunction(t, "Embedded colors with AllowColor true across multiple lines", wrapspec, input, expected, "")
}
