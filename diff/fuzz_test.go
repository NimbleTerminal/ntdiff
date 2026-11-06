package diff

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// Exercise malformed input as well as successful parses: partial trees are
// exposed by Parse and consumed by the legacy ParseDiff adapter.
func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"", "plain preamble\n", "diff --cc merged\n",
		"--- a/f\n+++ b/f\n@@ -1,2 +1,2 @@ section\n same\n-old\n+new\n\\ No newline at end of file\n",
		"--- /dev/null\n+++ b/new\n@@ -0,0 +1 @@\n+new\n",
		"--- a/old\n+++ /dev/null\n@@ -1 +0,0 @@\n-old\n",
		"diff --git a/old b/new\nrename from old\nrename to new\n",
		"diff --git a/f b/f\nGIT binary patch\nliteral 0\nHcmV?d00001\n",
		"--- a/f\n+++ b/f\n@@ -1,2 +1 @@\n-old\n+new\n",
		"--- a/f\n+++ b/f\n@@ -9223372036854775807,1 +1 @@\n-old\n+new\n",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		tree, err := Parse(text)
		if tree == nil {
			t.Fatal("nil partial tree")
		}
		legacy := ParseDiff(text)
		if len(legacy) != len(tree.Files) {
			t.Fatal("legacy file count differs")
		}
		for i, file := range tree.Files {
			adds, dels := 0, 0
			for _, h := range file.Hunks {
				old, new := h.Old.Start, h.New.Start
				for j, line := range h.Lines {
					wantOld, wantNew := false, false
					switch line.Kind {
					case Context:
						wantOld, wantNew = true, true
					case Deletion:
						wantOld = true
						dels++
					case Addition:
						wantNew = true
						adds++
					case NoNewline:
						if j == 0 || h.Lines[j-1].Kind == NoNewline {
							t.Fatal("orphan newline marker")
						}
					default:
						t.Fatalf("unknown line kind %d", line.Kind)
					}
					if wantOld {
						if line.OldNumber == nil || *line.OldNumber != old {
							t.Fatal("incorrect old line number")
						}
						old++
					} else if line.OldNumber != nil {
						t.Fatal("unexpected old line number")
					}
					if wantNew {
						if line.NewNumber == nil || *line.NewNumber != new {
							t.Fatal("incorrect new line number")
						}
						new++
					} else if line.NewNumber != nil {
						t.Fatal("unexpected new line number")
					}
				}
				if old-h.Old.Start > h.Old.Count || new-h.New.Start > h.New.Count {
					t.Fatal("partial hunk exceeds range")
				}
				if err == nil && (old-h.Old.Start != h.Old.Count || new-h.New.Start != h.New.Count) {
					t.Fatal("successful hunk does not fill range")
				}
			}
			a, d := file.Stats()
			if a != adds || d != dels || legacy[i].Additions != adds || legacy[i].Deletions != dels {
				t.Fatal("incorrect change counts")
			}
			var rows []string
			file.Rows(func(row Row) { rows = append(rows, rowText(row)) })
			if legacy[i].Path != file.Path() || legacy[i].Diff != strings.Join(rows, "\n") || legacy[i].Content != legacy[i].Diff {
				t.Fatal("legacy snapshot differs from tree")
			}
		}
	})
}

// Quoting must preserve path bytes, including escaped quotes, tabs, newlines,
// non-ASCII names, and invalid UTF-8, without confusing file boundaries.
func FuzzQuotedPaths(f *testing.F) {
	for _, paths := range [][2]string{{"old name", "new name"}, {"café", "日本語"}, {"a/quote\"\\\t\n", "b/\xff"}, {"", ""}} {
		f.Add(paths[0], paths[1])
	}
	f.Fuzz(func(t *testing.T, old, new string) {
		a, b := strconv.Quote("a/"+old), strconv.Quote("b/"+new)
		for _, text := range []string{
			fmt.Sprintf("diff --git %s %s\n", a, b),
			fmt.Sprintf("--- %s\told timestamp\n+++ %s\tnew timestamp\n", a, b),
		} {
			tree, err := Parse(text)
			if err != nil || len(tree.Files) != 1 {
				t.Fatalf("quoted paths rejected: %v", err)
			}
			if tree.Files[0].OldPath != old || tree.Files[0].NewPath != new {
				t.Fatalf("paths changed: got %q / %q, want %q / %q", tree.Files[0].OldPath, tree.Files[0].NewPath, old, new)
			}
		}
	})
}

// Arbitrary token ranges must never panic, drop, duplicate, or change text.
func FuzzTokenSpans(f *testing.F) {
	f.Add("héllo", 0, 3, 3, 6)
	f.Add("héllo", 1, 2, -1, 99)
	f.Add("abc", 0, 2, 1, 3)
	f.Add("\xff", 0, 1, 0, 0)
	f.Add("", 0, 0, 1, 2)
	f.Fuzz(func(t *testing.T, text string, a, b, c, d int) {
		line := Line{Text: text, Tokens: []Token{{Start: a, End: b, Kind: "first"}, {Start: c, End: d, Kind: "second"}}}
		var got strings.Builder
		for _, span := range line.spans() {
			got.WriteString(span.Text)
		}
		if got.String() != text {
			t.Fatal("token spans changed source text")
		}
	})
}
