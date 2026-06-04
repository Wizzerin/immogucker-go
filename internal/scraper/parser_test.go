package scraper

import (
	"testing"
)

// TestExtractDigits verifies the correct extraction of digits from price strings
func TestExtractDigits(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Standard price with Euro symbol",
			input:    "900 €",
			expected: "900",
		},
		{
			name:     "Price with thousands separator",
			input:    "1.200 €",
			expected: "1200",
		},
		{
			name:     "Price with extra text",
			input:    "Warmmiete: 550,00 EUR",
			expected: "55000", // extractDigits retrieves all numeric characters
		},
		{
			name:     "String without digits",
			input:    "Keine Angabe",
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Hidden special characters (non-breaking space)",
			input:    "850\u00A0€",
			expected: "850",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractDigits(tc.input)

			if result != tc.expected {
				t.Errorf("extractDigits(%q) = %q; expected %q", tc.input, result, tc.expected)
			}
		})
	}
}
