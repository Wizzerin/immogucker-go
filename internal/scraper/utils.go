package scraper

import (
	"strings"
	"unicode"
)

// extractDigits is a helper function that removes all non-numeric characters from a string.
// Because it starts with a lowercase letter, it is private to the 'scraper' package
// but accessible by all files within it (wggesucht.go, kleinanzeigen.go, etc.).
func extractDigits(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}
