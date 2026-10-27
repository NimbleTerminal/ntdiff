package diff

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var hunkPattern = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(?: (.*))?$`)

// Parse reads Git or plain unified diffs. It validates hunk counts and returns
// the partial tree with a line-numbered error on malformed input. Extended
// headers and binary payloads are retained without interpreting binary patches.
// Combined (merge) diffs are not supported.
func Parse(text string) (*Diff, error) {
	d := &Diff{}
	var f *File
	var h *Hunk
	oldLeft, newLeft, oldNum, newNum := 0, 0, 0, 0
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if text == "" {
		return d, nil
	}
	fail := func(i int, message string) (*Diff, error) { return d, fmt.Errorf("diff line %d: %s", i+1, message) }
	for i, s := range lines {
		if strings.HasPrefix(s, "diff --cc ") || strings.HasPrefix(s, "diff --combined ") || strings.HasPrefix(s, "@@@") {
			return fail(i, "combined diffs are unsupported")
		}
		if h != nil && s == `\ No newline at end of file` {
			if len(h.Lines) == 0 || h.Lines[len(h.Lines)-1].Kind == NoNewline {
				return fail(i, "orphan no-newline marker")
			}
			h.Lines = append(h.Lines, Line{Kind: NoNewline, Text: "No newline at end of file"})
			continue
		}
		if h != nil && (oldLeft > 0 || newLeft > 0) {
			if s == "" {
				return fail(i, "missing hunk line prefix")
			}
			l := Line{Text: s[1:]}
			switch s[0] {
			case ' ':
				l.Kind = Context
				l.OldNumber = number(oldNum)
				l.NewNumber = number(newNum)
				oldNum++
				newNum++
				oldLeft--
				newLeft--
			case '-':
				l.Kind = Deletion
				l.OldNumber = number(oldNum)
				oldNum++
				oldLeft--
			case '+':
				l.Kind = Addition
				l.NewNumber = number(newNum)
				newNum++
				newLeft--
			default:
				return fail(i, "incomplete hunk or invalid line prefix")
			}
			if oldLeft < 0 || newLeft < 0 {
				return fail(i, "hunk line count exceeded")
			}
			h.Lines = append(h.Lines, l)
			continue
		}
		if strings.HasPrefix(s, "diff --git ") {
			d.Files = append(d.Files, File{})
			f = &d.Files[len(d.Files)-1]
			h = nil
			f.OldPath, f.NewPath = gitPaths(strings.TrimPrefix(s, "diff --git "))
			f.Headers = append(f.Headers, Header{GitHeader, s})
			continue
		}
		if strings.HasPrefix(s, "--- ") && i+1 < len(lines) && strings.HasPrefix(lines[i+1], "+++ ") {
			if f == nil || h != nil || hasOldHeader(*f) {
				d.Files = append(d.Files, File{})
				f = &d.Files[len(d.Files)-1]
			}
			h = nil
			f.OldPath = headerPath(s[4:])
			f.Headers = append(f.Headers, Header{OldFileHeader, s})
			continue
		}
		if strings.HasPrefix(s, "+++ ") && f != nil && h == nil {
			f.NewPath = headerPath(s[4:])
			f.Headers = append(f.Headers, Header{NewFileHeader, s})
			continue
		}
		if strings.HasPrefix(s, "@@") {
			if f == nil {
				return fail(i, "hunk without file headers")
			}
			m := hunkPattern.FindStringSubmatch(s)
			if m == nil {
				return fail(i, "invalid hunk header")
			}
			vals := make([]int, 4)
			for j := 0; j < 4; j++ {
				v := m[j+1]
				if v == "" {
					v = "1"
				}
				n, err := strconv.Atoi(v)
				if err != nil {
					return fail(i, "hunk range overflows int")
				}
				vals[j] = n
			}
			maxInt := int(^uint(0) >> 1)
			if vals[0] > maxInt-vals[1] || vals[2] > maxInt-vals[3] {
				return fail(i, "hunk range overflows int")
			}
			if (vals[0] == 0 && vals[1] != 0) || (vals[2] == 0 && vals[3] != 0) {
				return fail(i, "nonempty range starts at zero")
			}
			f.Hunks = append(f.Hunks, Hunk{Old: Range{vals[0], vals[1]}, New: Range{vals[2], vals[3]}, Section: m[5]})
			h = &f.Hunks[len(f.Hunks)-1]
			oldNum, oldLeft, newNum, newLeft = vals[0], vals[1], vals[2], vals[3]
			continue
		}
		if h != nil {
			if s == "" {
				continue
			}
			return fail(i, "unexpected content after hunk")
		}
		if f == nil {
			d.Preamble = append(d.Preamble, s)
		} else {
			f.Headers = append(f.Headers, Header{ExtendedHeader, s})
			// Rename/copy headers provide exact paths even without text hunks.
			for _, prefix := range []string{"rename from ", "copy from "} {
				if strings.HasPrefix(s, prefix) {
					f.OldPath = unquote(strings.TrimPrefix(s, prefix))
				}
			}
			for _, prefix := range []string{"rename to ", "copy to "} {
				if strings.HasPrefix(s, prefix) {
					f.NewPath = unquote(strings.TrimPrefix(s, prefix))
				}
			}
		}
	}
	if oldLeft != 0 || newLeft != 0 {
		return fail(len(lines)-1, "truncated hunk")
	}
	return d, nil
}

func number(n int) *int { return &n }
func hasOldHeader(f File) bool {
	for _, h := range f.Headers {
		if h.Kind == OldFileHeader {
			return true
		}
	}
	return false
}
func unquote(s string) string {
	if u, e := strconv.Unquote(s); e == nil {
		return u
	}
	return s
}
func headerPath(s string) string { s = strings.SplitN(s, "\t", 2)[0]; return stripPrefix(unquote(s)) }
func stripPrefix(s string) string {
	if strings.HasPrefix(s, "a/") || strings.HasPrefix(s, "b/") {
		return s[2:]
	}
	return s
}
func gitPaths(s string) (string, string) {
	if strings.HasPrefix(s, `"`) {
		for i := 1; i < len(s); i++ {
			if s[i] == '\\' {
				i++
				continue
			}
			if s[i] == '"' {
				return stripPrefix(unquote(s[:i+1])), stripPrefix(unquote(strings.TrimSpace(s[i+1:])))
			}
		}
	}
	if i := strings.LastIndex(s, " b/"); i >= 0 {
		return stripPrefix(s[:i]), stripPrefix(s[i+1:])
	}
	parts := strings.Fields(s)
	if len(parts) == 2 {
		return stripPrefix(unquote(parts[0])), stripPrefix(unquote(parts[1]))
	}
	return "", ""
}

// ParseDiff is the legacy adapter. New callers should use Parse for errors and
// the structured tree. Diff and Content are derived snapshots, never UI state.
func ParseDiff(text string) []FileChange {
	d, _ := Parse(text)
	files := make([]FileChange, 0, len(d.Files))
	for _, f := range d.Files {
		a, r := f.Stats()
		var rows []string
		f.Rows(func(row Row) { rows = append(rows, rowText(row)) })
		raw := strings.Join(rows, "\n")
		files = append(files, FileChange{Path: f.Path(), Diff: raw, Content: raw, Additions: a, Deletions: r})
	}
	return files
}
