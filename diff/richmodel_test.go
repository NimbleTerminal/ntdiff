package diff

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

var ansiRE = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string { return ansiRE.ReplaceAllString(s, "") }

func makeRichDiff() RichDiff {
	return RichDiff{
		LeftLabel:  "A",
		RightLabel: "B",
		Lines: []RichLine{
			{Status: Collapsed, Meta: "5 shared"},
			{LeftText: "alpha", RightText: "beta", Status: Changed, Meta: "row1"},
			{LeftText: "gamma", RightText: "", Status: LeftOnly, Meta: "row2"},
		},
		ForkIndex: 1,
	}
}

func TestRichModel_LoadRichDiffStoresLines(t *testing.T) {
	m := NewRichModel()
	d := makeRichDiff()
	m.LoadRichDiff(d)
	if got := m.LineCount(); got != 3 {
		t.Fatalf("expected 3 lines, got %d", got)
	}
	if m.ForkIndex() != 1 {
		t.Fatalf("expected ForkIndex=1, got %d", m.ForkIndex())
	}
}

func TestRichModel_MetaRendererInvokedForEachLine(t *testing.T) {
	m := NewRichModel()
	var seen []any
	m.SetMetaRenderer(func(l RichLine) string {
		seen = append(seen, l.Meta)
		return "meta"
	})
	m.SetSize(80, 20)
	m.LoadRichDiff(makeRichDiff())
	_ = m.View()
	if len(seen) != 3 {
		t.Fatalf("expected MetaRenderer called for 3 lines, got %d (seen=%v)", len(seen), seen)
	}
}

func TestRichModel_RenderContainsLabelsAndForkMarker(t *testing.T) {
	m := NewRichModel()
	m.SetSize(80, 20)
	m.LoadRichDiff(makeRichDiff())
	out := m.View()
	for _, want := range []string{"A", "B", "Fork"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected rendered view to contain %q, got:\n%s", want, out)
		}
	}
}

func TestRichModel_UpdateAcceptsKeyMsgWithoutPanic(t *testing.T) {
	m := NewRichModel()
	m.SetSize(80, 20)
	m.LoadRichDiff(makeRichDiff())
	_, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	_, _ = m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
}

// TestRichModel_StyledRowAligns verifies that ANSI escape bytes from lipgloss
// styling do not corrupt column alignment. After stripping ANSI codes, the
// separator '│' must appear at exactly colWidth runes from the start of the row.
//
// With width=40 and no meta renderer:
//
//	usable   = 40 - 0 - 3 = 37
//	colWidth = 37 / 2     = 18
func TestRichModel_StyledRowAligns(t *testing.T) {
	d := RichDiff{
		LeftLabel:  "A",
		RightLabel: "B",
		Lines:      []RichLine{{LeftText: "x", RightText: "y", Status: Changed}},
	}
	m := NewRichModel()
	m.SetSize(40, 10)
	m.LoadRichDiff(d)
	out := m.View()

	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected >=2 output lines, got %d:\n%s", len(lines), out)
	}
	dataRow := stripANSI(lines[1])

	sep := strings.Index(dataRow, "│")
	if sep < 0 {
		t.Fatalf("no separator '│' found in data row: %q", dataRow)
	}

	// The left column must be exactly colWidth=18 runes wide before the separator.
	const wantColWidth = 18
	leftRunes := []rune(dataRow[:sep])
	if len(leftRunes) != wantColWidth {
		t.Errorf("expected %d runes before separator, got %d: %q",
			wantColWidth, len(leftRunes), dataRow[:sep])
	}
}

func TestRichModel_ForkRuleAppearsBeforeForkIndex(t *testing.T) {
	d := RichDiff{
		LeftLabel: "A", RightLabel: "B",
		Lines: []RichLine{
			{LeftText: "pre1", RightText: "pre1", Status: Shared},
			{LeftText: "pre2", RightText: "pre2", Status: Shared},
			{LeftText: "post1L", RightText: "post1R", Status: Changed},
		},
		ForkIndex: 2,
	}
	m := NewRichModel()
	m.SetSize(80, 20)
	m.LoadRichDiff(d)
	out := m.View()

	preIdx := strings.Index(out, "pre2")
	ruleIdx := strings.Index(out, "Fork @ step 2")
	postIdx := strings.Index(out, "post1L")

	if preIdx == -1 || ruleIdx == -1 || postIdx == -1 {
		t.Fatalf("expected all markers present, got out:\n%s", out)
	}
	if !(preIdx < ruleIdx && ruleIdx < postIdx) {
		t.Fatalf("expected order pre2 < forkrule < post1, got preIdx=%d ruleIdx=%d postIdx=%d", preIdx, ruleIdx, postIdx)
	}
}
