package diff

import (
	"context"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestTreeParsing(t *testing.T) {
	text := "--- a/f\n+++ b/f\n@@ -3,2 +8,3 @@ section\n same\n---source\n+++source\n+\n\\ No newline at end of file\n@@ -10 +16 @@\n-old\n+new\n"
	d, err := Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	f := d.Files[0]
	if f.Path() != "f" || len(f.Hunks) != 2 {
		t.Fatalf("%+v", f)
	}
	a, r := f.Stats()
	if a != 3 || r != 2 {
		t.Fatalf("stats %d %d", a, r)
	}
	ls := f.Hunks[0].Lines
	if *ls[0].OldNumber != 3 || *ls[0].NewNumber != 8 || *ls[1].OldNumber != 4 || ls[1].NewNumber != nil || ls[2].OldNumber != nil || *ls[2].NewNumber != 9 {
		t.Fatalf("wrong numbers: %+v", ls)
	}
	if ls[4].Kind != NoNewline || ls[4].OldNumber != nil || ls[4].NewNumber != nil {
		t.Fatal("marker consumes numbers")
	}
	if f.Hunks[1].Old.Count != 1 {
		t.Fatal("omitted count")
	}
}

func TestFileVariants(t *testing.T) {
	cases := []struct {
		name, text, path string
		hunks            int
	}{
		{"new", "--- /dev/null\n+++ b/new\n@@ -0,0 +1 @@\n+new\n", "new", 1},
		{"deleted", "--- a/gone\n+++ /dev/null\n@@ -1 +0,0 @@\n-old\n", "gone", 1},
		{"quoted", "diff --git \"a/caf\\303\\251\" \"b/caf\\303\\251\"\nold mode 100644\nnew mode 100755\n", "café", 0},
		{"renamed", "diff --git a/old name b/new name\nsimilarity index 100%\nrename from old name\nrename to new name\n", "new name", 0},
		{"binary", "diff --git a/a b/a\nBinary files a/a and b/a differ\n", "a", 0},
		{"timestamp", "--- old\t2026-01-01\n+++ new\t2026-01-02\n@@ -1 +1 @@\n-a\n+b\n", "new", 1},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d, e := Parse(tt.text)
			if e != nil {
				t.Fatal(e)
			}
			if len(d.Files) != 1 || d.Files[0].Path() != tt.path || len(d.Files[0].Hunks) != tt.hunks {
				t.Fatalf("%+v", d)
			}
		})
	}
	d, e := Parse(cases[0].text + cases[1].text)
	if e != nil || len(d.Files) != 2 {
		t.Fatalf("multiple plain files: %+v %v", d, e)
	}
}

func TestMalformedDiff(t *testing.T) {
	for _, s := range []string{
		"@@ -1 +1 @@\n-a\n+b", "diff --cc f", "--- a/f\n+++ b/f\n@@ bad", "--- a/f\n+++ b/f\n@@ -1,2 +1 @@\n-a\n+b", "--- a/f\n+++ b/f\n@@ -1 +1 @@\n+a\n+b", "--- a/f\n+++ b/f\n@@ -0 +1 @@\n-a\n+b",
	} {
		if _, e := Parse(s); e == nil {
			t.Errorf("accepted %q", s)
		}
	}
}

func TestTokensAndTerminal(t *testing.T) {
	line := Line{Kind: Addition, Text: "héllo", Tokens: []Token{{Start: 0, End: 3, Kind: "name", Style: TextStyle{Foreground: color.RGBA{R: 123, A: 255}, Bold: true}}}}
	f := File{Hunks: []Hunk{{Lines: []Line{line}}}}
	out := stripANSI(RenderFile(f))
	if !strings.Contains(out, "+héllo") {
		t.Fatal(out)
	}
	for _, tok := range []Token{{Start: -1, End: 2}, {Start: 1, End: 2}, {Start: 0, End: 90}} {
		line.Tokens = []Token{tok}
		var got strings.Builder
		for _, s := range line.spans() {
			got.WriteString(s.Text)
		}
		if got.String() != line.Text {
			t.Fatal(got.String())
		}
	}
}

func TestModelTreeNavigationAndReload(t *testing.T) {
	d := &Diff{Files: []File{{NewPath: "first", Hunks: []Hunk{{Lines: []Line{{Kind: Addition, Text: "alpha"}}}}}, {NewPath: "second", Hunks: []Hunk{{Lines: []Line{{Kind: Deletion, Text: "beta"}}}}}}}
	m := NewModel()
	m.LoadTree(d)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = updated.(Model)
	if m.selected != 1 || !strings.Contains(stripANSI(m.viewport.View()), "-beta") {
		t.Fatal("navigation did not use tree")
	}
	d.Files[0].Hunks[0].Lines[0].Text = "edited"
	m.LoadTree(d)
	if m.selected != 0 || m.files.Index() != 0 || !strings.Contains(stripANSI(m.viewport.View()), "+edited") {
		t.Fatal("reload did not reflect tree")
	}
	m.LoadTree(nil)
	if !strings.Contains(m.viewport.View(), "No diff loaded") {
		t.Fatal("stale viewport")
	}
}

func TestComparePaths(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	for p, s := range map[string]string{a: "old\n", b: "new\n"} {
		if e := os.WriteFile(p, []byte(s), 0600); e != nil {
			t.Fatal(e)
		}
	}
	d, e := ComparePaths(context.Background(), a, b)
	if e != nil {
		t.Fatal(e)
	}
	adds, dels := d.Files[0].Stats()
	if adds != 1 || dels != 1 {
		t.Fatal("missing changes")
	}
	d, e = ComparePaths(context.Background(), a, a)
	if e != nil || len(d.Files) != 0 {
		t.Fatalf("equal paths: %v %v", d, e)
	}
	if _, e = ComparePaths(context.Background(), a, filepath.Join(dir, "missing")); e == nil {
		t.Fatal("missing path accepted")
	}
}

func TestFromGitWithPathSeparator(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, e := cmd.CombinedOutput(); e != nil {
			t.Fatalf("git: %v: %s", e, out)
		}
	}
	git("init", "--quiet")
	if e := os.WriteFile(filepath.Join(dir, "file"), []byte("new\n"), 0600); e != nil {
		t.Fatal(e)
	}
	git("add", "file")
	d, e := FromGit(context.Background(), dir, "--cached", "--", "file")
	if e != nil {
		t.Fatal(e)
	}
	if len(d.Files) != 1 || d.Files[0].Path() != "file" {
		t.Fatalf("%+v", d)
	}
	a, r := d.Files[0].Stats()
	if a != 1 || r != 0 {
		t.Fatalf("stats %d %d", a, r)
	}
}

func TestCustomDiffPaneKeepsFileList(t *testing.T) {
	m := NewModel()
	if m.SelectedFileIndex() != -1 {
		t.Fatal("empty selection")
	}
	m.LoadTree(&Diff{Files: []File{{NewPath: "example.go"}}})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 25})
	m = updated.(Model)
	w, h := m.DiffPaneSize()
	if w != 58 || h != 22 {
		t.Fatalf("pane dimensions %d %d", w, h)
	}
	out := stripANSI(m.ViewWithDiffContent("custom document").Content)
	for _, s := range []string{"Changed Files", "example.go", "custom document", "Files Pane"} {
		if !strings.Contains(out, s) {
			t.Errorf("missing %q", s)
		}
	}
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = updated.(Model)
	if m.FocusedPane() != FocusDiff {
		t.Fatal("pane focus")
	}
}
