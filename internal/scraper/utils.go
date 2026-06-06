package scraper

import (
	"strings"
	"unicode"
)

// extractDigits is a helper function that removes all non-numeric characters from a string.
// Because it starts with a lowercase letter, it is private to the 'scraper' package
// but accessible by all files within it (wggesucht.go, kleinanzeigen.go, etc.).
// extractDigits removes decimals and non-numeric characters from a price string.
// Example: "780,00 €" -> "780", "700 - 790 €" -> "700"
func extractDigits(s string) string {
	parts := strings.Split(s, ",")
	if len(parts) == 1 {
		parts = strings.Split(s, ".")
	}

	var result strings.Builder
	for _, r := range parts[0] {
		if unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}
