package diff

import (
	"testing"
)

func TestParseDiff(t *testing.T) {
	diffText := `diff --git a/foo.txt b/foo.txt
index 0000000..1234567 100644
--- a/foo.txt
+++ b/foo.txt
@@ -1,2 +1,3 @@
 hello
-world
+there
+world!
`

	files := ParseDiff(diffText)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if f.Path != "foo.txt" {
		t.Errorf("expected path 'foo.txt', got '%s'", f.Path)
	}
	if f.Additions != 2 {
		t.Errorf("expected 2 additions, got %d", f.Additions)
	}
	if f.Deletions != 1 {
		t.Errorf("expected 1 deletion, got %d", f.Deletions)
	}
}
