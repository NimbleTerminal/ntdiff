// Package typeset converts ntdiff's tree into nttypeset documents. Layout,
// rasterization, fonts, and ntpicture display remain the embedding app's choice.
package typeset

import (
	"fmt"
	"image/color"
	"strings"

	"nimbleterminal.dev/ntdiff/diff"
	"nimbleterminal.dev/nttypeset"
)

// Options controls document presentation. Zero values use size 25, four-space
// tabs, and the default diff inks. TokenStyle can resolve a lexer category;
// it is called only for annotated source spans, never diff prefixes.
type Options struct {
	Size                       int
	TabWidth                   int
	LineNumbers                bool
	Addition, Deletion, Header color.RGBA
	TokenStyle                 func(kind string) diff.TextStyle
}

// Convert produces one monospace KindLine block per row. Line numbers are
// explicit old/new columns: nttypeset's automatic sequential gutter is disabled.
// Colors on the +/- prefixes survive syntax highlighting. A nil tree is empty.
func Convert(tree *diff.Diff, opts Options) *nttypeset.Document {
	doc := &nttypeset.Document{}
	if tree == nil {
		return doc
	}
	if opts.Size == 0 {
		opts.Size = 25
	}
	if opts.TabWidth <= 0 {
		opts.TabWidth = 4
	}
	if opts.Addition.A == 0 {
		opts.Addition = color.RGBA{R: 46, G: 160, B: 67, A: 255}
	}
	if opts.Deletion.A == 0 {
		opts.Deletion = color.RGBA{R: 248, G: 81, B: 73, A: 255}
	}
	if opts.Header.A == 0 {
		opts.Header = color.RGBA{R: 88, G: 166, B: 255, A: 255}
	}
	appendRow := func(r diff.Row) {
		ink := color.RGBA{}
		style := nttypeset.Mono
		switch r.Kind {
		case diff.AdditionRow:
			ink = opts.Addition
		case diff.DeletionRow:
			ink = opts.Deletion
		case diff.GitRow, diff.FileRow, diff.HunkRow:
			ink = opts.Header
			style |= nttypeset.Bold
		}
		b := nttypeset.Block{Kind: nttypeset.KindLine, Size: opts.Size}
		if opts.LineNumbers {
			number := func(n *int) string {
				if n == nil {
					return ""
				}
				return fmt.Sprint(*n)
			}
			b.Spans = append(b.Spans, nttypeset.Span{Text: fmt.Sprintf("%5s %5s │ ", number(r.OldNumber), number(r.NewNumber)), Style: nttypeset.Mono})
		}
		for _, s := range r.Spans {
			ts := s.Style
			if opts.TokenStyle != nil && s.TokenKind != "" {
				ts = opts.TokenStyle(s.TokenKind)
			}
			span := nttypeset.Span{Text: strings.ReplaceAll(s.Text, "\t", strings.Repeat(" ", opts.TabWidth)), Style: style, Ink: ink}
			if ts.Foreground.A != 0 {
				span.Ink = ts.Foreground
			}
			if ts.Bold {
				span.Style |= nttypeset.Bold
			}
			if ts.Italic {
				span.Style |= nttypeset.Italic
			}
			b.Spans = append(b.Spans, span)
		}
		// Keep empty metadata lines visible to the layout engine.
		empty := true
		for _, s := range b.Spans {
			if s.Text != "" {
				empty = false
			}
		}
		if empty {
			b.Spans = []nttypeset.Span{{Text: " ", Style: style}}
		}
		doc.Blocks = append(doc.Blocks, b)
	}
	for _, s := range tree.Preamble {
		appendRow(diff.Row{Spans: []diff.Span{{Text: s}}})
	}
	for _, f := range tree.Files {
		f.Rows(appendRow)
	}
	return doc
}
