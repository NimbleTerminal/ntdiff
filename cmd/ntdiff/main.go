package main

import (
	"context"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"nimbleterminal.dev/ntdiff/diff"
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
		var tree *diff.Diff
		var err error
		if len(os.Args) == 3 && pathExists(os.Args[1]) && pathExists(os.Args[2]) {
			tree, err = diff.ComparePaths(context.Background(), os.Args[1], os.Args[2])
		} else {
			tree, err = diff.FromGit(context.Background(), "", os.Args[1:]...)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		run(tree)
		return
	}
	tree, err := diff.Parse(diffText)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	run(tree)
}

func pathExists(path string) bool { _, err := os.Stat(path); return err == nil }

func run(tree *diff.Diff) {
	model := diff.NewModel()
	model.LoadTree(tree)

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
