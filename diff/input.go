package diff

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// FromGit runs git diff in dir (empty means the current directory). Arguments
// are passed directly to Git, without a shell; output is forced to unified text.
func FromGit(ctx context.Context, dir string, args ...string) (*Diff, error) {
	gitArgs := []string{"diff"}
	// Force formatting before the path separator, if supplied.
	split := len(args)
	for i, arg := range args {
		if arg == "--" {
			split = i
			break
		}
	}
	gitArgs = append(gitArgs, args[:split]...)
	gitArgs = append(gitArgs, "--no-color", "--no-ext-diff", "--no-textconv")
	gitArgs = append(gitArgs, args[split:]...)
	return runGit(ctx, dir, gitArgs)
}

// ComparePaths compares two files or directories, even inside a Git repository.
func ComparePaths(ctx context.Context, oldPath, newPath string) (*Diff, error) {
	return runGit(ctx, "", []string{"diff", "--no-index", "--no-color", "--no-ext-diff", "--no-textconv", "--", oldPath, newPath})
}

func runGit(ctx context.Context, dir string, args []string) (*Diff, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		var exit *exec.ExitError
		// Git's --no-index and --exit-code report differences with status 1.
		if !errors.As(err, &exit) || exit.ExitCode() != 1 || (len(out) == 0 && stderr.Len() > 0) {
			return nil, fmt.Errorf("git diff: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
	}
	return Parse(string(out))
}
