package typeset

import (
	"context"
	"golang.org/x/image/font/gofont/gomono"
	"image/color"
	"strings"
	"testing"

	"nimbleterminal.dev/ntdiff/diff"
	"nimbleterminal.dev/nttypeset"
)

func TestConvertPreservesStructureAndSyntax(t *testing.T) {
	d, e := diff.Parse("--- a/f\n+++ b/f\n@@ -1 +1,2 @@\n-old\n+new\n+\n\\ No newline at end of file\n")
	if e != nil {
		t.Fatal(e)
	}
	ink := color.RGBA{B: 200, A: 255}
	d.Files[0].Hunks[0].Lines[1].Tokens = []diff.Token{{Start: 0, End: 3, Kind: "keyword", Style: diff.TextStyle{Foreground: ink, Bold: true}}}
	doc := Convert(d, Options{})
	if len(doc.Blocks) != 7 || doc.Lines {
		t.Fatalf("%+v", doc)
	}
	var text []string
	for _, b := range doc.Blocks {
		if b.Kind != nttypeset.KindLine {
			t.Fatal("lost line layout")
		}
		var s strings.Builder
		for _, sp := range b.Spans {
			if sp.Style&nttypeset.Mono == 0 {
				t.Fatal("not monospace")
			}
			s.WriteString(sp.Text)
		}
		text = append(text, s.String())
	}
	want := "--- a/f\n+++ b/f\n@@ -1,1 +1,2 @@\n-old\n+new\n+\n\\ No newline at end of file"
	if strings.Join(text, "\n") != want {
		t.Fatalf("%q", text)
	}
	addition := doc.Blocks[4].Spans
	if addition[0].Ink == ink || addition[1].Ink != ink || addition[1].Style&nttypeset.Bold == 0 {
		t.Fatal("syntax/diff inks lost")
	}
	if doc.Blocks[3].Spans[0].Ink == addition[0].Ink {
		t.Fatal("addition/deletion indistinguishable")
	}
	if doc.Blocks[2].Spans[0].Style&nttypeset.Bold == 0 {
		t.Fatal("hunk unstyled")
	}
}

func TestOptionsAndTreeEdits(t *testing.T) {
	d, _ := diff.Parse("--- a/f\n+++ b/f\n@@ -1 +1 @@\n-a\n+\tb\n")
	l := &d.Files[0].Hunks[0].Lines[1]
	l.Text = "\tc"
	l.Tokens = []diff.Token{{Start: 1, End: 2, Kind: "name"}}
	calls := 0
	doc := Convert(d, Options{LineNumbers: true, TabWidth: 2, Size: 18, TokenStyle: func(kind string) diff.TextStyle { calls++; return diff.TextStyle{Italic: true} }})
	row := doc.Blocks[4]
	var s strings.Builder
	for _, sp := range row.Spans {
		s.WriteString(sp.Text)
	}
	if !strings.HasSuffix(s.String(), "+  c") || !strings.Contains(s.String(), "    1 │ ") || row.Size != 18 || calls != 1 {
		t.Fatalf("row %q calls %d", s.String(), calls)
	}
	if len(Convert(nil, Options{}).Blocks) != 0 {
		t.Fatal("nil not empty")
	}
}

// This exercises the real shaper and rasterizer with embedded, deterministic
// fonts, without depending on installed system fonts or a terminal protocol.
type testFonts struct{}

func (testFonts) Face(_, _ string) ([]byte, error) { return gomono.TTF, nil }

func TestDocumentLayoutsAndRasters(t *testing.T) {
	d, e := diff.Parse("--- a/f\n+++ b/f\n@@ -1 +1 @@\n-old\n+new\n")
	if e != nil {
		t.Fatal(e)
	}
	engine, e := nttypeset.NewEngine(nttypeset.WithFonts(testFonts{}))
	if e != nil {
		t.Fatal(e)
	}
	layout, e := engine.Layout(Convert(d, Options{LineNumbers: true}), 640, nttypeset.LayoutOptions{})
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	if e = layout.Complete(ctx); e != nil {
		t.Fatal(e)
	}
	img, e := layout.Raster(ctx, nttypeset.View{Width: 640, Height: layout.Extent(), Scale: 1}, nttypeset.RasterOptions{Palette: nttypeset.DarkPalette()})
	if e != nil {
		t.Fatal(e)
	}
	if img.Bounds().Dx() != 640 || img.Bounds().Dy() <= 100 {
		t.Fatal(img.Bounds())
	}
	red, green := 0, 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < 640; x++ {
			c := img.RGBAAt(x, y)
			if int(c.R) > int(c.G)*2 && c.R > 100 {
				red++
			}
			if int(c.G) > int(c.R)*2 && c.G > 100 {
				green++
			}
		}
	}
	if red == 0 || green == 0 {
		t.Fatalf("diff colors missing from raster: red=%d green=%d", red, green)
	}
}
