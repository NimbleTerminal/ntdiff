package diff

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// MetaRenderer is supplied by the caller to format the per-row metadata column.
// It receives the RichLine and returns a string already styled by the caller.
type MetaRenderer func(RichLine) string

// RichModel renders a RichDiff in three columns: left text, right text, and an
// optional caller-rendered meta column. It supports vertical scrolling, jump to
// the fork marker, and exposes hooks for the caller to intercept keys (e.g. to
// expand or collapse rows).
type RichModel struct {
	diff       RichDiff
	width      int
	height     int
	scroll     int
	metaRender MetaRenderer
	// SelectedLine is the index into diff.Lines of the highlighted row, or -1.
	SelectedLine int
}

// NewRichModel returns a RichModel with empty state.
func NewRichModel() RichModel {
	return RichModel{SelectedLine: -1}
}

// LoadRichDiff replaces the current RichDiff and resets scroll / selection.
func (m *RichModel) LoadRichDiff(d RichDiff) {
	m.diff = d
	m.scroll = 0
	if len(d.Lines) > 0 {
		m.SelectedLine = 0
	} else {
		m.SelectedLine = -1
	}
}

// SetMetaRenderer installs (or clears, when nil) the meta column renderer.
func (m *RichModel) SetMetaRenderer(r MetaRenderer) { m.metaRender = r }

// SetSize updates the model's render dimensions.
func (m *RichModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// LineCount returns the number of RichLines currently loaded.
func (m *RichModel) LineCount() int { return len(m.diff.Lines) }

// ForkIndex returns the fork point index for the currently loaded diff.
func (m *RichModel) ForkIndex() int { return m.diff.ForkIndex }

// JumpToFork scrolls so the fork marker is visible.
func (m *RichModel) JumpToFork() {
	if m.diff.ForkIndex < 0 {
		return
	}
	m.scroll = m.diff.ForkIndex
	m.SelectedLine = m.diff.ForkIndex
}

// Update handles built-in scroll keys; callers may pre-process keys before
// forwarding (e.g. to intercept Enter on Collapsed rows).
func (m RichModel) Update(msg tea.Msg) (RichModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			if m.SelectedLine < len(m.diff.Lines)-1 {
				m.SelectedLine++
			}
			m.scrollIntoView()
		case "k", "up":
			if m.SelectedLine > 0 {
				m.SelectedLine--
			}
			m.scrollIntoView()
		case "J":
			m.SelectedLine += 10
			if m.SelectedLine >= len(m.diff.Lines) {
				m.SelectedLine = len(m.diff.Lines) - 1
			}
			m.scrollIntoView()
		case "K":
			m.SelectedLine -= 10
			if m.SelectedLine < 0 {
				m.SelectedLine = 0
			}
			m.scrollIntoView()
		case "g":
			m.SelectedLine = 0
			m.scroll = 0
		case "G":
			m.SelectedLine = len(m.diff.Lines) - 1
			m.scrollIntoView()
		case "f":
			m.JumpToFork()
		}
	}
	return m, nil
}

func (m *RichModel) scrollIntoView() {
	visible := m.height - 2 // header + status
	if visible < 1 {
		visible = 1
	}
	if m.SelectedLine < m.scroll {
		m.scroll = m.SelectedLine
	}
	if m.SelectedLine >= m.scroll+visible {
		m.scroll = m.SelectedLine - visible + 1
	}
}

// View renders the diff as a string. Layout:
//
//	<header: LeftLabel | RightLabel>
//	<left col> | <right col> | <meta col, if MetaRenderer set>
//	   ...
//
// A horizontal rule is rendered between the row that would occupy index
// ForkIndex-1 and the row at ForkIndex.
func (m RichModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	hasMeta := m.metaRender != nil
	metaWidth := 0
	if hasMeta {
		metaWidth = 28
	}
	usable := m.width - metaWidth - 3 // 3 = two column separators + buffer
	if usable < 4 {
		usable = 4
	}
	colWidth := usable / 2

	header := lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("%-*s│ %-*s", colWidth, m.diff.LeftLabel, colWidth, m.diff.RightLabel),
	)

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")

	visible := m.height - 2
	if visible < 1 {
		visible = 1
	}

	end := m.scroll + visible
	if end > len(m.diff.Lines) {
		end = len(m.diff.Lines)
	}

	forkRule := lipgloss.NewStyle().Foreground(ColorForkRule).Render(
		strings.Repeat("─", colWidth) + "┼" + strings.Repeat("─", colWidth) +
			fmt.Sprintf(" Fork @ step %d ", m.diff.ForkIndex),
	)

	for i := m.scroll; i < end; i++ {
		if i == m.diff.ForkIndex && i > 0 {
			b.WriteString(forkRule)
			b.WriteString("\n")
		}
		line := m.diff.Lines[i]
		left := truncate(line.LeftText, colWidth)
		right := truncate(line.RightText, colWidth)

		switch line.Status {
		case Shared:
			st := lipgloss.NewStyle().Foreground(ColorSharedLine)
			left = st.Render(left)
			right = st.Render(right)
		case LeftOnly:
			left = lipgloss.NewStyle().Foreground(ColorRemoved).Render(left)
		case RightOnly:
			right = lipgloss.NewStyle().Foreground(ColorAdded).Render(right)
		case Collapsed:
			label := fmt.Sprintf("··· %d shared rows (enter to expand) ···", collapsedCount(line))
			left = lipgloss.NewStyle().Foreground(ColorSharedLine).Render(
				center(label, colWidth*2+1),
			)
			right = ""
		case Changed:
			left = lipgloss.NewStyle().Foreground(ColorRemoved).Render(left)
			right = lipgloss.NewStyle().Foreground(ColorAdded).Render(right)
		}

		if i == m.SelectedLine {
			left = lipgloss.NewStyle().Reverse(true).Render(left)
			right = lipgloss.NewStyle().Reverse(true).Render(right)
		}

		row := fmt.Sprintf("%-*s│ %-*s", colWidth, left, colWidth, right)
		if hasMeta {
			meta := m.metaRender(line)
			row += " │ " + truncate(meta, metaWidth)
		}
		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return string(runes[:n])
	}
	return string(runes[:n-1]) + "…"
}

func center(s string, n int) string {
	if len(s) >= n {
		return s
	}
	pad := (n - len(s)) / 2
	return strings.Repeat(" ", pad) + s
}

// CollapsedCounter is satisfied by Meta values that represent a collapsed run
// of shared rows. ntdiff calls CollapsedRowCount() to render the count label
// without depending on the caller's concrete type.
type CollapsedCounter interface{ CollapsedRowCount() int }

func collapsedCount(l RichLine) int {
	if c, ok := l.Meta.(CollapsedCounter); ok {
		return c.CollapsedRowCount()
	}
	return 0
}
