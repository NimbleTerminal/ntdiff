# ntdiff - Nimble Terminal Diff

<p>
    <a href="https://github.com/NimbleMarkets/ntdiff/actions/workflows/pages.yml"><img src="https://github.com/NimbleMarkets/ntdiff/actions/workflows/pages.yml/badge.svg" alt="Deploy to GitHub Pages Status"></a>
    <a href="https://github.com/NimbleMarkets/ntdiff/tags"><img src="https://img.shields.io/github/tag/NimbleMarkets/ntdiff.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/NimbleMarkets/ntdiff?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/NimbleMarkets/ntdiff/blob/main/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg"  alt="Code Of Conduct"></a>
</p>

`ntdiff` is a Golang Terminal Diff viewer for the Bubble Tea Framework and other TUIs. It parses unified diffs and provides an interactive split-pane terminal interface to explore code changes.

---

## Features

- **Split-Pane Layout:** List of changed files with addition/deletion counts on the left; scrollable, colorized unified diff viewport on the right.
- **Dynamic Adjustments:** Responsive UI that adapts to terminal resize events automatically.
- **Syntax Highlighting:** Uses Lip Gloss to colorize diff components (additions in green, deletions in red, hunk headers in cyan, diff headers in magenta).
- **Interactive Navigation:** Smooth pane switching (`Tab` or `h`/`l`) and scrolling/list navigation (`↑`/`↓` or `k`/`j`).
- **Flexible Inputs:** Automatically detects piped stdin, accepts direct file/directory paths, or runs `git diff` fallback.
- **WebAssembly Ready:** Full browser demo utilizing the Ghostty terminal emulator engine (`ghostty-web`) via `booba`.

---

## Installation & CLI Usage

### Install the CLI
```bash
go install github.com/NimbleMarkets/ntdiff/cmd/ntdiff@latest
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
	"github.com/NimbleMarkets/ntdiff/diff"
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
import "github.com/NimbleMarkets/ntdiff/diff"
```

Please note that the dependencies are tied to `charm.land/bubbletea/v2` and `charm.land/lipgloss/v2` for WebAssembly and rendering enhancements.

---

## Web Demo

`ntdiff` features a browser-based demo that runs the compiled WebAssembly binary inside a native high-fidelity web terminal engine (`ghostty-web`).

### Building and Serving the Site Locally
```bash
# Build the WebAssembly site assets
task build-wasm-site

# Serve locally at http://localhost:8000/ntdiff/
task serve-wasm-site
```
