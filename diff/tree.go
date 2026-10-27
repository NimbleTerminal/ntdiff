package diff

import (
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"
)

// Diff is the parsed source of truth. Metadata is owned by the caller.
type Diff struct {
	Preamble []string
	Files    []File
	Metadata map[string]any
}

// File holds paths, extended headers (including binary notices), and text hunks.
// Paths retain /dev/null for creations/deletions; Git's a/ and b/ prefixes are removed.
type File struct {
	OldPath, NewPath string
	Headers          []Header
	Hunks            []Hunk
	Metadata         map[string]any
}

func (f File) Path() string {
	if f.NewPath != "" && f.NewPath != "/dev/null" {
		return f.NewPath
	}
	return f.OldPath
}

// Stats derives counts from the tree, never from headers or cached raw text.
func (f File) Stats() (additions, deletions int) {
	for _, h := range f.Hunks {
		for _, l := range h.Lines {
			if l.Kind == Addition {
				additions++
			}
			if l.Kind == Deletion {
				deletions++
			}
		}
	}
	return
}

type HeaderKind uint8

const (
	ExtendedHeader HeaderKind = iota
	GitHeader
	OldFileHeader
	NewFileHeader
)

type Header struct {
	Kind HeaderKind
	Text string
}

type Range struct{ Start, Count int }

type Hunk struct {
	Old, New Range
	Section  string
	Lines    []Line
	Metadata map[string]any
}

func (h Hunk) Header() string {
	s := fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.Old.Start, h.Old.Count, h.New.Start, h.New.Count)
	if h.Section != "" {
		s += " " + h.Section
	}
	return s
}

type LineKind uint8

const (
	Context LineKind = iota
	Addition
	Deletion
	// NoNewline annotates the preceding line and consumes no line number.
	NoNewline
)

type Line struct {
	Kind                 LineKind
	Text                 string // content without the unified-diff prefix
	OldNumber, NewNumber *int
	Tokens               []Token
	Metadata             map[string]any
}

// Token annotates a half-open UTF-8 byte range of Line.Text. Kind is a caller's
// lexer category; Style is optional portable styling shared by both renderers.
// Tokens must be ordered, non-overlapping, and lie on UTF-8 boundaries.
type Token struct {
	Start, End int
	Kind       string
	Style      TextStyle
}

type TextStyle struct {
	Foreground   color.RGBA // zero alpha inherits the diff color
	Bold, Italic bool
}

// Span and Row are a derived presentation stream, not additional stored state.
type Span struct {
	Text      string
	TokenKind string
	Style     TextStyle
}
type RowKind uint8

const (
	MetadataRow RowKind = iota
	GitRow
	FileRow
	HunkRow
	ContextRow
	AdditionRow
	DeletionRow
	NoNewlineRow
)

type Row struct {
	Kind                 RowKind
	Spans                []Span
	OldNumber, NewNumber *int
}

// Rows visits headers and hunks in reading order. Invalid token annotations are
// ignored as a whole so they can never corrupt or duplicate the source text.
func (f File) Rows(yield func(Row)) {
	for _, h := range f.Headers {
		kind := MetadataRow
		if h.Kind != ExtendedHeader {
			kind = FileRow
		}
		if h.Kind == GitHeader {
			kind = GitRow
		}
		yield(Row{Kind: kind, Spans: []Span{{Text: h.Text}}})
	}
	if len(f.Headers) == 0 && f.Path() != "" {
		yield(Row{Kind: FileRow, Spans: []Span{{Text: f.Path()}}})
	}
	for _, h := range f.Hunks {
		yield(Row{Kind: HunkRow, Spans: []Span{{Text: h.Header()}}})
		for _, l := range h.Lines {
			kind, prefix := ContextRow, " "
			switch l.Kind {
			case Addition:
				kind, prefix = AdditionRow, "+"
			case Deletion:
				kind, prefix = DeletionRow, "-"
			case NoNewline:
				kind, prefix = NoNewlineRow, "\\ "
			}
			spans := []Span{{Text: prefix}}
			spans = append(spans, l.spans()...)
			yield(Row{Kind: kind, Spans: spans, OldNumber: l.OldNumber, NewNumber: l.NewNumber})
		}
	}
}

func (l Line) spans() []Span {
	end := 0
	for _, t := range l.Tokens {
		if t.Start < end || t.End <= t.Start || t.End > len(l.Text) || !utf8.ValidString(l.Text[t.Start:t.End]) || !utf8.ValidString(l.Text[end:t.Start]) {
			return []Span{{Text: l.Text}}
		}
		end = t.End
	}
	var spans []Span
	end = 0
	for _, t := range l.Tokens {
		if t.Start > end {
			spans = append(spans, Span{Text: l.Text[end:t.Start]})
		}
		spans = append(spans, Span{Text: l.Text[t.Start:t.End], TokenKind: t.Kind, Style: t.Style})
		end = t.End
	}
	if end < len(l.Text) {
		spans = append(spans, Span{Text: l.Text[end:]})
	}
	return spans
}

func rowText(r Row) string {
	var b strings.Builder
	for _, s := range r.Spans {
		b.WriteString(s.Text)
	}
	return b.String()
}
