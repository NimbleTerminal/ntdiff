# ntdiff — One diff, two views

<p>
    <a href="https://pkg.go.dev/nimbleterminal.dev/ntdiff/diff"><img src="https://pkg.go.dev/badge/nimbleterminal.dev/ntdiff/diff.svg" alt="Go Reference"></a>
    <a href="https://github.com/NimbleTerminal/ntdiff/actions/workflows/pages.yml"><img src="https://github.com/NimbleTerminal/ntdiff/actions/workflows/pages.yml/badge.svg" alt="CI and demo deployment"></a>
    <a href="./CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="Code Of Conduct"></a>
</p>

Browse unified diffs in a [Bubble Tea v2](https://github.com/charmbracelet/bubbletea)
terminal UI, or turn the same parsed tree into a typeset document. Use ntdiff as
an interactive command-line viewer or embed its components in your own app.

- **`diff.Parse`** produces files, hunks, and numbered lines, preserving Git
  metadata and reporting malformed input with a line-numbered error.
- **`diff.Model`** pairs a changed-file list with a scrollable, colorized unified
  diff. Both panes respond to terminal resizing and keyboard navigation.
- **`typeset.Convert`** produces an [nttypeset](https://github.com/NimbleTerminal/nttypeset)
  document with optional old/new line numbers and caller-supplied syntax styles.
- **`diff.RichModel`** displays caller-supplied side-by-side rows, a fork marker,
  and an optional metadata column.

[Open the browser demo](https://nimbleterminal.github.io/ntdiff/) ·
[Changelog](./CHANGELOG.md)

## Install

Requires Go 1.26.8 or later. Git is needed for repository diffs and comparisons
between two paths; parsing an existing patch does not invoke Git.

```sh
# CLI
go install nimbleterminal.dev/ntdiff/cmd/ntdiff@latest

# Library
go get nimbleterminal.dev/ntdiff@latest
```

To build from a checkout, run `task build` and then `bin/ntdiff`.

## CLI

```sh
ntdiff                         # unstaged changes in the current repository
ntdiff --cached                # staged changes; arguments are passed to git diff
ntdiff HEAD~1 -- path/to/file   # compare a revision, limiting the paths
ntdiff patch.diff              # read a unified diff file
ntdiff old.txt new.txt         # compare two files (directories also work)
git diff HEAD~1 | ntdiff       # read a patch from stdin
```

The CLI prefers piped input, then a single existing patch file, then two existing
paths. Other arguments go to `git diff`. Invalid patches produce an error instead
of opening a partially parsed view.

Use `Tab` or `h`/`l` to switch panes, arrows or `j`/`k` to navigate, and `q` or
`Ctrl+C` to quit. Added and deleted lines have distinct colors; source-language
syntax highlighting is optional and supplied by the embedding app.

## Quick start

```go
package main

import (
    "log"

    tea "charm.land/bubbletea/v2"
    "nimbleterminal.dev/ntdiff/diff"
)

func main() {
    patch := "--- a/hello.txt\n+++ b/hello.txt\n@@ -1 +1,2 @@\n hello\n+world!\n"
    tree, err := diff.Parse(patch)
    if err != nil {
        log.Fatal(err)
    }

    model := diff.NewModel()
    model.LoadTree(tree)
    if _, err := tea.NewProgram(model).Run(); err != nil {
        log.Fatal(err)
    }
}
```

In an existing Bubble Tea app, forward messages to the model's `Update`, run its
returned commands, and use its `View`. `SelectedFileIndex`, `FocusedPane`, and
`DiffPaneSize` expose the current layout and selection. `ViewWithDiffContent`
keeps the file list and pane chrome while your app supplies the diff pane.

## Structured diffs

`Diff.Files`, `File.Hunks`, and `Hunk.Lines` are the source of truth. Each line
contains its kind, text without the diff prefix, and optional old/new line
numbers. `File.Stats()` derives addition/deletion counts from those lines.
`File.Rows` visits the derived headers and styled spans in reading order;
`diff.RenderFile` renders them with Lip Gloss.

`Parse` handles Git and plain unified patches, multiple files and hunks,
creations, deletions, renames, mode changes, quoted paths, and no-final-newline
markers. It preserves preambles, extended headers, and binary payloads without
decoding binary patches. Invalid hunk ranges/counts and combined merge diffs
return errors. On error, the returned tree contains the successfully parsed
prefix.

This is a display model, not a byte-for-byte patch serializer. Derived hunk
headers use explicit counts. Treat a loaded tree as immutable; after editing it,
call `LoadTree` again to refresh the view and reset selection and scrolling.
The diff, file, hunk, and line types have caller-owned `Metadata` maps.

To acquire input directly:

```go
tree, err := diff.FromGit(ctx, repoDir, "--cached", "--", "file.go")
// Or compare files/directories inside or outside a repository:
tree, err = diff.ComparePaths(ctx, oldPath, newPath)
```

Both helpers honor context cancellation and accept Git's “differences found”
exit status. Git arguments are passed without a shell.

### Syntax annotations

Attach `Line.Tokens` using ordered, non-overlapping UTF-8 byte ranges of
`Line.Text`. Tokens carry a category and an optional `TextStyle` (RGBA foreground,
bold, italic). Both renderers preserve the diff color on the `+`/`-` marker while
allowing source colors in the body. Invalid ranges fall back to plain line text.

Tokenization belongs to your app. When full old/new source is available,
highlight that source and attach tokens using the line numbers; partial hunks
alone cannot reliably establish multiline lexer state.

## Typeset documents

```go
import diffTypeset "nimbleterminal.dev/ntdiff/typeset"

doc := diffTypeset.Convert(tree, diffTypeset.Options{
    LineNumbers: true,
    TabWidth:    4,
})
```

`Convert` returns an `*nttypeset.Document` with monospace `KindLine` blocks.
Old/new line-number columns are explicit spans rather than nttypeset's sequential
gutter. Options control text size, tab width, diff colors, and category-to-style
mapping through `TokenStyle`.

Your app owns fonts, viewport dimensions, layout, rasterization, and image
presentation. Given a configured nttypeset engine:

```go
layout, err := engine.Layout(doc, 1000, nttypeset.LayoutOptions{})
if err != nil { return err }
img, err := layout.Raster(ctx,
    nttypeset.View{Width: 1000, Height: 800, Scale: 1},
    nttypeset.RasterOptions{Palette: nttypeset.DarkPalette()},
)
if err != nil { return err }
// Present img through your app's ntpicture instance.
```

See [the demo](./examples/demo/main.go) for font setup, selected-file rendering,
and [ntpicture](https://github.com/NimbleTerminal/ntpicture) integration. Importing
only `diff` does not link the typeset renderer; nttypeset still contributes to the
module's Go version requirement.

## Side-by-side rows

`RichDiff` is a separate, caller-supplied representation for comparisons that
already have aligned rows. Each `RichLine` contains left/right text, a status
(`Shared`, `LeftOnly`, `RightOnly`, `Changed`, or `Collapsed`), and optional
metadata. `ForkIndex` marks a boundary between rows.

```go
model := diff.NewRichModel()
model.SetSize(100, 24)
model.LoadRichDiff(diff.RichDiff{
    LeftLabel: "Before", RightLabel: "After",
    Lines: []diff.RichLine{
        {LeftText: "old", RightText: "new", Status: diff.Changed},
    },
})
content := model.View()
```

Forward messages to `RichModel.Update` and retain the returned model. Use
`SetMetaRenderer` for the optional third column and `JumpToFork` to reveal the
fork marker. The host owns expanding/collapsing rows and other application keys.

## Demo

The native and browser demo share a four-file patch, a syntax annotator, and the
same parsed tree. Press `v` to switch terminal/typeset views, `n` to toggle line
numbers in typeset mode, and `t` to switch glyph/Kitty rendering when supported.

```sh
task run-demo                   # native terminal view
task run-demo -- -typeset        # start in typeset mode
task build-examples             # bin/ntdiff-demo
task build-wasm-site            # build and smoke-test the browser demo
task serve-wasm-site            # http://localhost:8000/ntdiff/
```

The [examples module](./examples/go.mod) keeps Booba and font bundles out of the
library's direct dependencies. `wasm.work` pins browser-compatible Bubble Tea and
clipboard forks; native builds use upstream Bubble Tea. The browser runtime and
WASM output are generated into `web/` and ignored by Git.

## Migration

The repository moved from NimbleMarkets to NimbleTerminal. Replace imports of
`github.com/NimbleMarkets/ntdiff` with `nimbleterminal.dev/ntdiff`, including the
`/diff` and `/typeset` packages, then run `go mod tidy`. Update all imports
together: Go treats different module paths as different type identities.

`NewModel`, `LoadDiff`, `ParseDiff`, and `ColorizeDiff` remain available. Prefer
`Parse` followed by `LoadTree` when you need errors; the legacy no-error wrappers
use the parsed prefix. `ParseDiff` returns derived `FileChange` snapshots with
normalized hunk headers. The `RichDiff`/`RichModel` API remains separate.

## Open Collaboration

Run `task ci` before submitting changes. It checks formatting, module tidiness,
dependency checksums, race tests, fuzzing, vet, native builds, and the WASM demo.
It requires Go 1.26.8+, a C compiler for the race detector, Task, and Node 22+.
GitHub Actions runs the same task before deploying Pages.

```sh
task test                 # tests and fuzz seed cases in both modules
task fuzz                 # three targets, 10 seconds each, two workers
task fuzz FUZZTIME=1m      # a longer campaign for each target
task go-tidy              # update module files when dependencies change
task check-site           # check already-built assets and both WASM views
```

Fuzzing checks parser line numbers and hunk counts (including partial trees),
quoted paths, and text preservation under arbitrary token ranges. Go saves
failing inputs under `diff/testdata/fuzz` for reproduction and regression tests.
Module tidiness checks in CI do not rewrite files.

Contributions and feedback are welcome. Please follow our
[Code of Conduct](./CODE_OF_CONDUCT.md).

- [Report an issue](https://github.com/NimbleTerminal/ntdiff/issues)
- [Submit a pull request](https://github.com/NimbleTerminal/ntdiff/pulls)

## Acknowledgements

Built on [Charm](https://charm.land)'s Bubble Tea, Bubbles, and Lip Gloss, with
[nttypeset](https://github.com/NimbleTerminal/nttypeset),
[ntpicture](https://github.com/NimbleTerminal/ntpicture), and
[Booba](https://github.com/NimbleTerminal/booba) powering the typeset/browser demo.

----
Made with :heart: and :fire: by the team behind
[Nimble Terminal](https://nimbleterminal.dev).
