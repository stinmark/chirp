package theme

import (
	"strings"
)

// TruncateString cuts a string to a specified max length.
func TruncateString(s string, maxLength int, addEllipsis bool) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)

	if len(runes) <= maxLength {
		return s
	}

	truncated := string(runes[:maxLength])

	if addEllipsis {
		truncated = strings.TrimSpace(truncated) + "..."
	}

	return truncated
}
