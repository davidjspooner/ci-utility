package textfmt

import (
	"testing"
)

func TestParseSGRCodes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCodes []int
		expectedError string
	}{
		{
			name:          "Valid SGR codes",
			input:         "\u001b[31m",
			expectedCodes: []int{31},
			expectedError: "",
		},
		{
			name:          "Multiple valid SGR codes",
			input:         "\u001b[1;32;40m",
			expectedCodes: []int{1, 32, 40},
			expectedError: "",
		},
		{
			name:          "Invalid SGR code",
			input:         "\u001b[99m",
			expectedCodes: nil,
			expectedError: "invalid SGR code",
		},
		{
			name:          "Invalid format",
			input:         "\u001b[31",
			expectedCodes: nil,
			expectedError: "invalid SGR sequence format",
		},
		{
			name:          "Empty sequence",
			input:         "\u001b[m",
			expectedCodes: []int{},
			expectedError: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			codes, err := parseSGRCodes(test.input)
			if err != nil {
				if test.expectedError == "" {
					t.Errorf("unexpected error: %v", err)
				} else if err.Error() != test.expectedError {
					t.Errorf("expected error %q, got %q", test.expectedError, err.Error())
				}
				return
			} else if test.expectedError != "" {
				t.Errorf("expected error %q, got nil", test.expectedError)
				return
			}

			if len(codes) != len(test.expectedCodes) {
				t.Errorf("expected codes %v, got %v", test.expectedCodes, codes)
				return
			}
			for i, code := range codes {
				if code != test.expectedCodes[i] {
					t.Errorf("at index %d, expected code %d, got %d", i, test.expectedCodes[i], code)
				}
			}
		})
	}
}
