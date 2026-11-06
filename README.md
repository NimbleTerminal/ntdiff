# ntdiff - Nimble Terminal Diff

<p>
    <a href="https://github.com/NimbleTerminal/ntdiff/actions/workflows/pages.yml"><img src="https://github.com/NimbleTerminal/ntdiff/actions/workflows/pages.yml/badge.svg" alt="Deploy to GitHub Pages Status"></a>
    <a href="https://github.com/NimbleTerminal/ntdiff/tags"><img src="https://img.shields.io/github/tag/NimbleTerminal/ntdiff.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/nimbleterminal.dev/ntdiff?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/NimbleTerminal/ntdiff/blob/main/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg"  alt="Code Of Conduct"></a>
</p>

`ntdiff` is a Golang Terminal Diff viewer for the Bubble Tea Framework and other TUIs. It parses unified diffs and provides an interactive split-pane terminal interface to explore code changes.

---

## Features

- **Split-Pane Layout:** List of changed files with addition/deletion counts on the left; scrollable, colorized unified diff viewport on the right.
- **Dynamic Adjustments:** Responsive UI that adapts to terminal resize events automatically.
- **Syntax Highlighting:** Uses Lip Gloss to colorize diff components (additions in green, deletions in red, hunk headers in cyan, diff headers in magenta).
- **Interactive Navigation:** Smooth pane switching (`Tab` or `h`/`l`) and scrolling/list navigation (`↑`/`↓` or `k`/`j`).
- **Flexible Inputs:** Automatically detects piped stdin, accepts direct file/directory paths, or runs `git diff` fallback.
- **WebAssembly Ready:** Full browser demo utilizing the Ghostty terminal emulator engine (`ntwebterm`) via `booba`.

---

## Installation & CLI Usage

### Install the CLI
```bash
go install nimbleterminal.dev/ntdiff/cmd/ntdiff@latest
```

### Usage
```bash
# Run git diff inside a repository
ntdiff

# Read a diff from a file
ntdiff patch.diff

# Compare two files or directories outside a git repository
ntdiff file1.txt file2.txt

# Pipe a command output directly
git diff HEAD~1 | ntdiff
```

---

## Library Usage

`ntdiff` can also be integrated into other Bubble Tea applications as a model.

```go
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"nimbleterminal.dev/ntdiff/diff"
)

func main() {
	diffText := `diff --git a/hello.txt b/hello.txt
--- a/hello.txt
+++ b/hello.txt
@@ -1,1 +1,2 @@
 hello
+world!
`

	model := diff.NewModel()
	model.LoadDiff(diffText)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
```

---

## BubbleTea Version Compatibility

We have migrated to **Bubble Tea v2**. You should import `ntdiff` as so:

```go
import "nimbleterminal.dev/ntdiff/diff"
```

Please note that the dependencies are tied to `charm.land/bubbletea/v2` and `charm.land/lipgloss/v2` for WebAssembly and rendering enhancements.

---

## Web Demo

The browser demo shows the same parsed four-file patch in the existing terminal
UI and as an nttypeset document rasterized through ntpicture. Press `v` to switch
views without reparsing or losing the selected file, `j`/`k` to navigate, `Tab` to switch panes in either view, and
`n` to toggle typeset line numbers. Browser buttons advertise the available keys.

```bash
# Run the same demo natively (Go 1.26.8+)
task run-demo
# Start in typeset mode
task run-demo -- -typeset

# Build the browser demo and run its Node smoke tests (Node 22+)
task build-wasm-site

# Serve at http://localhost:8000/ntdiff/
task serve-wasm-site
```

The demo lives in `examples/demo` with its own module, embedded sample patch,
and a small illustrative syntax annotator. `web/` contains the static site.
`wasm.work` pins the team's browser-compatible Bubble Tea and clipboard forks;
native builds keep upstream dependencies. Generated WASM and booba/ntwebterm assets
are ignored by Git and recreated by `go -C examples tool booba-assets --no-index`.
The GitHub Pages workflow builds and smoke-tests the demo before publishing.

`task ci` runs formatting, module tidiness and checksum checks, race tests,
bounded fuzzing, `go vet`, native builds, and the WASM build and smoke tests. It requires Go
1.26.8+, a C compiler for the race detector, Task, and Node 22+. Module tidiness
checks do not rewrite files; use `task go-tidy` to fix them.

`task fuzz` exercises parser line numbers and hunk counts (including partial
trees), quoted file paths, and text preservation under arbitrary token ranges.
Each target runs for 10 seconds with two workers. Use `task fuzz FUZZTIME=1m`
for a longer campaign. Seed cases also run with ordinary `go test`; Go saves
failing inputs under `diff/testdata/fuzz` for reproduction and regression tests.

`task test` tests both modules. `task check-site` checks already-built web assets
and exercises the real WASM bridge under Node in both starting views, including
live mode switching, action publication, resizing, and clean exit.

## Structured diffs and typeset documents

The module requires Go 1.26.8 or later and uses the canonical import path
`nimbleterminal.dev/ntdiff`.

Parse once, then use the same tree for either renderer:

```go
tree, err := diff.Parse(diffText)
if err != nil {
    return err // includes the input line number; tree contains the parsed prefix
}
model := diff.NewModel()
model.LoadTree(tree)

// Import diffTypeset "nimbleterminal.dev/ntdiff/typeset".
doc := diffTypeset.Convert(tree, diffTypeset.Options{LineNumbers: true})
// doc is *nttypeset.Document, ready for an app-owned nttypeset.Engine.
```

The tree consists of `Diff.Files`, `File.Hunks`, and `Hunk.Lines`. Lines carry a
`LineKind`, source text without its diff prefix, and optional `*int` old/new
numbers. Zero-length hunk ranges are distinct from absent line numbers.
`File.Stats()` derives counts from the tree. `File.Headers` preserves extended
Git metadata, binary notices/payloads, and file headers in order; `Diff.Preamble`
preserves introductory text. Each tree level has an optional caller-owned
`Metadata` map. Binary payloads are retained, not decoded.

`diff.FromGit(ctx, dir, args...)` and `diff.ComparePaths(ctx, old, new)` acquire
and parse input. The latter works inside or outside a repository and handles
Git's “differences found” exit status. The CLI uses these same helpers.

`Parse` accepts Git and plain unified diffs, including multiple files/hunks,
renames, mode changes, quoted Git paths, and no-final-newline markers. Malformed
hunk ranges/counts and combined merge diffs return errors. This is a display
model, not a byte-for-byte patch serializer: hunk headers use explicit counts,
and trailing patch separators are not retained inside hunks.

`Line.Tokens` optionally annotates ordered, non-overlapping UTF-8 byte ranges of
`Line.Text`. Each token has a lexer `Kind` and portable `TextStyle` (RGBA ink,
bold, italic). Both renderers use those styles; typeset callers can additionally
map categories with `Options.TokenStyle`. Invalid token ranges fall back to the
whole line's plain diff style. Tokenization is caller-owned: highlight complete
old/new source when available, then attach tokens using the line numbers. This
avoids guessing multiline lexical state from partial hunks. The +/- marker
retains its diff color even when syntax colors override the body.

`File.Rows` is the shared, derived row/span traversal. `diff.RenderFile` applies
Lip Gloss styling; `typeset.Convert` creates monospace `KindLine` blocks with
styled spans. Optional old/new line-number columns are explicit spans;
nttypeset's sequential document gutter is disabled. Tabs default to four spaces.
The document uses unified rows, matching the existing file-list/viewport UI.

The embedding app owns fonts, viewport sizing, layout, rasterization, and
ntpicture's terminal protocol lifecycle. For example, given a configured engine:

```go
layout, err := engine.Layout(doc, 1000, nttypeset.LayoutOptions{})
if err != nil { return err }
img, err := layout.Raster(ctx,
    nttypeset.View{Width: 1000, Height: 800, Scale: 1},
    nttypeset.RasterOptions{Palette: nttypeset.DarkPalette()},
)
if err != nil { return err }
// Pass img (image.Image) to the app's ntpicture model.
```

This keeps the font bundles and image-display lifecycle out of the diff library.
The `typeset` package is optional to import; terminal-only binaries do not link
nttypeset. Its dependency still sets the module-wide Go version requirement.

Existing `NewModel`, `LoadDiff`, `ParseDiff`, and `ColorizeDiff` entry points remain
available. Prefer `Parse` to surface errors: the legacy no-error wrappers use the
parsed prefix. `ParseDiff` returns derived `FileChange` snapshots with normalized
hunk headers. Treat a loaded tree as immutable and call `LoadTree` again after
editing it; that resets selection and scrolling. The separate caller-supplied
`RichDiff` / `RichModel` API is unchanged.

### Module migration

The repository lives at [NimbleTerminal/ntdiff](https://github.com/NimbleTerminal/ntdiff)
and the canonical Go module path is `nimbleterminal.dev/ntdiff`. Existing consumers
should change their imports from `github.com/NimbleMarkets/ntdiff` to
`nimbleterminal.dev/ntdiff`, including the `/diff` and `/typeset` packages, and run
`go mod tidy`. The package layout is unchanged.

Update all imports together: Go treats the old and new module paths as different
type identities. A GitHub repository redirect alone does not migrate Go imports.
