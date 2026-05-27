package diff

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

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
