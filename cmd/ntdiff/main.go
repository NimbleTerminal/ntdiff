package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/NimbleMarkets/ntdiff/diff"
)

func main() {
	var diffText string

	if isInputFromPipe() {
		// Read diff from stdin
		data, err := os.ReadFile("/dev/stdin")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
		diffText = string(data)
	} else if len(os.Args) == 2 && fileExists(os.Args[1]) {
		// Read diff from a single file
		data, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		diffText = string(data)
	} else {
		// Run git diff, forwarding any arguments
		gitArgs := []string{"diff", "--no-color"}
		if len(os.Args) > 1 {
			gitArgs = append(gitArgs, os.Args[1:]...)
		}

		cmd := exec.Command("git", gitArgs...)
		var stderr strings.Builder
		cmd.Stderr = &stderr

		out, err := cmd.Output()
		if err != nil {
			stderrStr := stderr.String()
			isNotGitRepo := strings.Contains(stderrStr, "Not a git repository") ||
				strings.Contains(stderrStr, "not a git repository")

			// Fallback: try to compare two files/folders outside of a git repository
			if isNotGitRepo && len(os.Args) > 1 {
				retryArgs := []string{"diff", "--no-color", "--no-index"}
				retryArgs = append(retryArgs, os.Args[1:]...)
				retryCmd := exec.Command("git", retryArgs...)
				var retryStderr strings.Builder
				retryCmd.Stderr = &retryStderr

				if retryOut, retryErr := retryCmd.Output(); retryErr == nil {
					diffText = string(retryOut)
					goto diffLoaded
				} else {
					// Update error details to the retry attempt
					err = retryErr
					stderrStr = retryStderr.String()
				}
			}

			// If it fails because it's not a git repository, show a friendly usage guide
			if isNotGitRepo {
				fmt.Fprint(os.Stderr, "ntdiff: A terminal diff viewer.\n\n")
				fmt.Fprint(os.Stderr, "Usage:\n")
				fmt.Fprint(os.Stderr, "  ntdiff <file>             Read a diff from a file\n")
				fmt.Fprint(os.Stderr, "  ntdiff <file1> <file2>    Compare two files/directories outside a git repository\n")
				fmt.Fprint(os.Stderr, "  ntdiff [git diff flags]   Run git diff in a repository (e.g. --cached, HEAD~1)\n")
				fmt.Fprint(os.Stderr, "  some-command | ntdiff     Read a diff piped from stdin\n\n")
				fmt.Fprint(os.Stderr, "Error: Not currently in a git repository.\n")
				os.Exit(1)
			}

			fmt.Fprintf(os.Stderr, "Error running git diff: %v\n%s", err, stderrStr)
			os.Exit(1)
		}
		diffText = string(out)
	}

diffLoaded:
	model := diff.NewModel()
	model.LoadDiff(diffText)

	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// isInputFromPipe checks if standard input is being piped or redirected.
func isInputFromPipe() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// fileExists checks if a path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
