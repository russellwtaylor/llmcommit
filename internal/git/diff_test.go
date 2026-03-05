package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a temp directory with a git repo initialized.
// It sets up user.email and user.name so commits work in CI environments.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup command %v failed: %v\n%s", args, err, out)
		}
	}
	return dir
}

// runInDir runs a command in the given directory and fails the test on error.
func runInDir(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("command %v failed: %v\n%s", args, err, out)
	}
}

// writeFile writes content to a file inside dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
}

// TestGetStagedDiff_ReturnsDiffWhenFilesStaged verifies that GetStagedDiff returns
// a non-empty diff when files have been staged.
func TestGetStagedDiff_ReturnsDiffWhenFilesStaged(t *testing.T) {
	dir := initTestRepo(t)

	// Create and stage a file
	writeFile(t, dir, "hello.txt", "hello world\n")
	runInDir(t, dir, "git", "add", "hello.txt")

	// Change working directory to the repo so GetStagedDiff picks it up
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	diff, err := GetStagedDiff()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if diff == "" {
		t.Fatal("expected non-empty diff, got empty string")
	}
	if !strings.Contains(diff, "hello.txt") {
		t.Errorf("expected diff to mention hello.txt, got:\n%s", diff)
	}
}

// TestGetStagedDiff_ErrNoStagedFiles verifies that ErrNoStagedFiles is returned
// when no files have been staged.
func TestGetStagedDiff_ErrNoStagedFiles(t *testing.T) {
	dir := initTestRepo(t)

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	_, err = GetStagedDiff()
	if !errors.Is(err, ErrNoStagedFiles) {
		t.Fatalf("expected ErrNoStagedFiles, got: %v", err)
	}
}

// TestSplitDiffByFile_MultiFile verifies that a multi-file diff is split correctly.
func TestSplitDiffByFile_MultiFile(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
index 0000000..1111111 100644
--- a/foo.go
+++ b/foo.go
@@ -0,0 +1,3 @@
+package main
+
+// foo
diff --git a/bar.go b/bar.go
index 0000000..2222222 100644
--- a/bar.go
+++ b/bar.go
@@ -0,0 +1,3 @@
+package main
+
+// bar
`

	result := SplitDiffByFile(diff)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(result), result)
	}

	for _, key := range []string{"foo.go", "bar.go"} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in result, keys: %v", key, mapKeys(result))
		}
	}

	if !strings.Contains(result["foo.go"], "diff --git a/foo.go b/foo.go") {
		t.Errorf("foo.go section missing header, got:\n%s", result["foo.go"])
	}
	if !strings.Contains(result["bar.go"], "diff --git a/bar.go b/bar.go") {
		t.Errorf("bar.go section missing header, got:\n%s", result["bar.go"])
	}
	// Ensure sections don't bleed into each other
	if strings.Contains(result["foo.go"], "bar.go") {
		t.Errorf("foo.go section unexpectedly contains bar.go content")
	}
}

// TestSplitDiffByFile_SingleFile verifies that a single-file diff produces a map
// with exactly one entry.
func TestSplitDiffByFile_SingleFile(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index 0000000..abcdef0 100644
--- a/main.go
+++ b/main.go
@@ -0,0 +1,5 @@
+package main
+
+func main() {
+}
`

	result := SplitDiffByFile(diff)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d: %v", len(result), result)
	}
	if _, ok := result["main.go"]; !ok {
		t.Errorf("expected key 'main.go', got keys: %v", mapKeys(result))
	}
}

// TestGetStagedDiff_BinaryFile verifies that binary files in the staged diff do
// not cause an error and appear in the returned diff string.
func TestGetStagedDiff_BinaryFile(t *testing.T) {
	dir := initTestRepo(t)

	// Write a minimal PNG (1x1 white pixel) as a binary file
	binaryContent := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk length + type
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // width=1, height=1
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, // bit depth, color, crc...
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
		0x00, 0x05, 0xFE, 0x02, 0xFE, 0xA7, 0x35, 0x81,
		0x84, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	path := filepath.Join(dir, "image.png")
	if err := os.WriteFile(path, binaryContent, 0644); err != nil {
		t.Fatal(err)
	}
	runInDir(t, dir, "git", "add", "image.png")

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	diff, err := GetStagedDiff()
	if err != nil {
		t.Fatalf("expected no error for binary file, got: %v", err)
	}
	if diff == "" {
		t.Fatal("expected non-empty diff for binary file, got empty string")
	}
	// Git marks binary files; the diff should mention image.png
	if !strings.Contains(diff, "image.png") {
		t.Errorf("expected diff to mention image.png, got:\n%s", diff)
	}
}

// mapKeys returns the keys of a map for use in error messages.
func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
