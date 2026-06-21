package theme

import (
	_ "embed" // Required for go:embed directive compiler hooks
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/common-nighthawk/go-figure"
)

// 1. Instruct the Go compiler to bake bloody.flf into this byte slice
//
//go:embed Bloody.flf
var bloodyFontBytes []byte

// GenerateTexturedShadowTitle handles rendering the font as a flat, blocky structure
// with no drop-shadow tracing, prioritizing maximum readability and retaining internal line boxes.
func GenerateTexturedShadowTitle(input string, fgColor string) string {
	fgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(fgColor))

	// 2. Convert the embedded byte slice into an io.Reader on the fly
	fontReader := strings.NewReader(string(bloodyFontBytes))

	// 3. Pass the reader directly into go-figure. No errors, no disk files required!
	myFigure := figure.NewFigureWithFont(input, fontReader, true)
	rawASCII := myFigure.String()

	// 4. Split the raw ASCII into lines and process each character
	lines := strings.Split(rawASCII, "\n")
	var completedBanner []string

	for _, line := range lines {
		// Skip entirely empty padding rows if they appear
		if strings.TrimSpace(line) == "" {
			continue
		}

		runes := []rune(line)
		var builtLine strings.Builder

		for i := 0; i < len(runes); i++ {
			char := runes[i]

			// If it's a structural space, preserve it as an empty gap
			if char == ' ' || char == '\u00a0' {
				builtLine.WriteRune(' ')
			} else {
				// Style every single asset of the glyph (solid blocks, lines, '│')
				// identically with the foreground color to preserve the box framework
				builtLine.WriteString(fgStyle.Render(string(char)))
			}
		}
		completedBanner = append(completedBanner, builtLine.String())
	}

	return strings.Join(completedBanner, "\n")
}

// TruncateString cuts a string to a specified max length.
// If addEllipsis is true and the string is truncated, it appends "..."
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
