# `ntdiff` CHANGELOG

## Unreleased

### Migration

- Move the repository to `NimbleTerminal/ntdiff` and the Go module to
  `nimbleterminal.dev/ntdiff`. Consumers must update their imports together and
  run `go mod tidy`; the `/diff` and `/typeset` package layout is unchanged.
- Require Go 1.26.8 or later. Update Bubble Tea to v2.1.1, Bubbles to v2.2.1,
  Lip Gloss to v2.0.6, nttypeset to v0.2.0, and indirect dependencies.
- Migrate the example to `nimbleterminal.dev/booba` v0.12.1, ntfonts v0.3.1,
  and ntpicture v0.1.5. Update the browser-only Bubble Tea pin and exported
  runtime assets from `ghostty-web` to `ntwebterm`.

### Added

- Structured diff trees: `Diff`, `File`, `Hunk`, and `Line`, with optional
  metadata, old/new line numbers, derived statistics, and row/span traversal.
- `Parse` with line-numbered errors and partial-tree results. Support Git and
  plain unified diffs, multiple files/hunks, quoted paths, renames, mode changes,
  binary metadata/payload preservation, and no-final-newline markers. Reject
  malformed hunk counts/ranges, integer overflow, and combined merge diffs.
- `FromGit` and `ComparePaths` helpers with context cancellation. The CLI uses
  the same parser and helpers, including comparisons outside a repository.
- Caller-supplied token annotations with portable foreground/bold/italic styles;
  invalid ranges preserve the original line text.
- `typeset.Convert` for nttypeset documents, including optional old/new line
  numbers, tab expansion, diff colors, and token-category styles.
- `LoadTree`, selection/layout accessors, and `ViewWithDiffContent` for embedding
  a custom diff pane alongside the existing file list.
- Native and browser demos sharing a parsed four-file patch, file navigation,
  terminal/typeset view switching, line-number controls, and Booba action buttons.
- `task ci` with formatting, module tidiness, checksum verification, race tests,
  fuzzing, vet, native builds, and WASM smoke checks. GitHub Actions runs it
  before deploying the demo; manual workflow runs are also supported.
- Fuzz targets for parser invariants and legacy snapshots, quoted paths, and
  token-span text preservation. `task fuzz` runs bounded campaigns with a
  configurable per-target duration.

### Changed

- Derive legacy `ParseDiff` snapshots and terminal rendering from the structured
  tree. Legacy no-error wrappers retain parsed prefixes; callers needing strict
  validation should use `Parse`. Derived hunk headers use explicit counts.
- Include the examples module and Go tools in `task go-update`. Native builds
  use read-only module resolution; the demo binary is written to `bin/ntdiff-demo`.
- Compact the browser hero and view selectors, and version the stylesheet URL
  so browsers fetch the updated layout.
- Refresh installation, API, demo, contributor, and module-migration guidance.

## v0.0.1 (2026-06-09)

First tagged release, under `github.com/NimbleMarkets/ntdiff`.

- Interactive unified-diff viewer with a changed-file list, addition/deletion
  counts, colorized viewport, keyboard navigation, and terminal resizing.
- CLI input from stdin, patch files, path comparisons, and Git diffs.
- Embeddable Bubble Tea model, diff parsing, and colorization helpers.
- `RichDiff`, `RichLine`, and `LineStatus` for caller-supplied comparisons;
  `RichModel` for side-by-side rendering with optional metadata, shared/collapsed
  rows, and fork markers.
- Display-width-aware row padding and fork-marker placement regression tests.
- Contributor Code of Conduct.

[Unreleased changes](https://github.com/NimbleTerminal/ntdiff/compare/v0.0.1...HEAD)
