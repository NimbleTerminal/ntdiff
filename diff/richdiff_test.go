package diff

import "testing"

func TestRichDiff_StatusConstantsDistinct(t *testing.T) {
	statuses := map[LineStatus]bool{
		Shared:    true,
		LeftOnly:  true,
		RightOnly: true,
		Changed:   true,
		Collapsed: true,
	}
	if len(statuses) != 5 {
		t.Fatalf("expected 5 distinct LineStatus values, got %d", len(statuses))
	}
}

func TestRichDiff_ZeroValueIsEmpty(t *testing.T) {
	var d RichDiff
	if len(d.Lines) != 0 {
		t.Fatalf("expected empty Lines, got %d", len(d.Lines))
	}
	if d.ForkIndex != 0 {
		t.Fatalf("expected ForkIndex=0, got %d", d.ForkIndex)
	}
}

func TestRichLine_MetaIsAny(t *testing.T) {
	type custom struct{ N int }
	l := RichLine{Meta: custom{N: 7}}
	got, ok := l.Meta.(custom)
	if !ok || got.N != 7 {
		t.Fatalf("expected Meta to round-trip a custom struct, got %#v ok=%v", l.Meta, ok)
	}
}
