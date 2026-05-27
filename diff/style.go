package diff

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Theme color definitions
var (
	ColorActiveBorder   = lipgloss.Color("6")   // Cyan
	ColorInactiveBorder = lipgloss.Color("240") // Dark Gray

	ColorAddition   = lipgloss.Color("2")   // Green
	ColorDeletion   = lipgloss.Color("9")   // Red
	ColorHunkHeader = lipgloss.Color("6")   // Cyan
	ColorDiffHeader = lipgloss.Color("13")  // Magenta/Purple
	ColorFileHeader = lipgloss.Color("252") // Bold White
	ColorStatusText = lipgloss.Color("246") // Muted Gray
	ColorStatusBg   = lipgloss.Color("235") // Dark Gray
)

var (
	// ColorSharedLine tints lines that are identical on both sides of a RichDiff.
	ColorSharedLine = lipgloss.Color("8")
	// ColorForkRule tints the horizontal rule rendered at a RichDiff fork point.
	ColorForkRule = lipgloss.Color("6")
	// ColorRemoved tints lines/tokens present only on the left (removed) side.
	ColorRemoved = ColorDeletion
	// ColorAdded tints lines/tokens present only on the right (added) side.
	ColorAdded = ColorAddition
)

// ColorizeDiff parses a diff block line-by-line and applies appropriate Lipgloss styling.
func ColorizeDiff(diffText string) string {
	lines := strings.Split(diffText, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			lines[i] = lipgloss.NewStyle().Foreground(ColorAddition).Render(line)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			lines[i] = lipgloss.NewStyle().Foreground(ColorDeletion).Render(line)
		} else if strings.HasPrefix(line, "@@") {
			lines[i] = lipgloss.NewStyle().Foreground(ColorHunkHeader).Render(line)
		} else if strings.HasPrefix(line, "diff --git") {
			lines[i] = lipgloss.NewStyle().Bold(true).Foreground(ColorDiffHeader).Render(line)
		} else if strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			lines[i] = lipgloss.NewStyle().Bold(true).Foreground(ColorFileHeader).Render(line)
		}
	}
	return strings.Join(lines, "\n")
}
