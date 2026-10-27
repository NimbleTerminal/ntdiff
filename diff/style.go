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

// RenderFile renders the structured file with Lip Gloss. Syntax colors override
// body ink while the +/- prefix always retains the diff kind's color.
func RenderFile(f File) string {
	var lines []string
	f.Rows(func(r Row) {
		base := lipgloss.NewStyle()
		switch r.Kind {
		case AdditionRow:
			base = base.Foreground(ColorAddition)
		case DeletionRow:
			base = base.Foreground(ColorDeletion)
		case HunkRow:
			base = base.Foreground(ColorHunkHeader)
		case GitRow:
			base = base.Foreground(ColorDiffHeader).Bold(true)
		case FileRow:
			base = base.Foreground(ColorFileHeader).Bold(true)
		}
		var b strings.Builder
		for _, s := range r.Spans {
			st := base
			if s.Style.Foreground.A != 0 {
				st = st.Foreground(s.Style.Foreground)
			}
			if s.Style.Bold {
				st = st.Bold(true)
			}
			if s.Style.Italic {
				st = st.Italic(true)
			}
			b.WriteString(st.Render(s.Text))
		}
		lines = append(lines, b.String())
	})
	return strings.Join(lines, "\n")
}

// ColorizeDiff is a compatibility wrapper. Parsing happens once; RenderFile
// performs no prefix recognition on source text.
func ColorizeDiff(text string) string {
	d, _ := Parse(text)
	parts := append([]string(nil), d.Preamble...)
	for _, f := range d.Files {
		parts = append(parts, RenderFile(f))
	}
	return strings.Join(parts, "\n")
}
