package diff

import (
	"strings"
)

// ParseDiff parses unified diff text and returns a list of FileChange elements.
func ParseDiff(diffText string) []FileChange {
	var files []FileChange
	var currentFile *FileChange
	var currentDiff []string

	lines := strings.Split(diffText, "\n")
	for _, line := range lines {
		// Detect the start of a new file block
		if strings.HasPrefix(line, "diff --git ") {
			if currentFile != nil {
				currentFile.Diff = strings.Join(currentDiff, "\n")
				currentFile.Content = currentFile.Diff
				files = append(files, *currentFile)
			}
			currentFile = &FileChange{}
			currentDiff = []string{line}

			// Initial path extraction attempt from diff --git a/path b/path
			parts := strings.Split(line, " b/")
			if len(parts) > 1 {
				currentFile.Path = strings.TrimPrefix(parts[1], "b/")
			} else {
				// Fallback: split by spaces and take the last element
				sp := strings.Fields(line)
				if len(sp) > 0 {
					currentFile.Path = strings.TrimPrefix(sp[len(sp)-1], "b/")
				}
			}
		} else {
			if currentFile != nil {
				currentDiff = append(currentDiff, line)

				// Refine path name if we encounter explicit file headers
				if strings.HasPrefix(line, "--- a/") {
					path := strings.TrimPrefix(line, "--- a/")
					if currentFile.Path == "" {
						currentFile.Path = path
					}
				} else if strings.HasPrefix(line, "+++ b/") {
					path := strings.TrimPrefix(line, "+++ b/")
					currentFile.Path = path
				}

				// Count additions and deletions
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					currentFile.Additions++
				} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
					currentFile.Deletions++
				}
			}
		}
	}

	if currentFile != nil {
		currentFile.Diff = strings.Join(currentDiff, "\n")
		currentFile.Content = currentFile.Diff
		files = append(files, *currentFile)
	}

	return files
}
