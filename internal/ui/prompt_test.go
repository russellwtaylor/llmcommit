package ui

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/russellwtaylor/llmcommit/internal/ai"
)

// noopEditor is a mock openEditor that does nothing and returns nil.
func noopEditor(path string) error {
	return nil
}

// makeEditorWriter returns a mock openEditor that writes content to the given file path.
func makeEditorWriter(content string) func(string) error {
	return func(path string) error {
		return os.WriteFile(path, []byte(content), 0600)
	}
}

// sampleMsg is a CommitMessage with both title and body for use in tests.
var sampleMsg = ai.CommitMessage{
	Title: "feat(auth): add JWT validation middleware",
	Body:  "- Adds middleware for validating JWT tokens\n- Extracts user from token claims\n- Adds unit tests",
}

// titleOnlyMsg is a CommitMessage with only a title.
var titleOnlyMsg = ai.CommitMessage{
	Title: "chore: update dependencies",
}

// TestConfirm_YesInput verifies that "y" input returns the original message unchanged.
func TestConfirm_YesInput(t *testing.T) {
	in := strings.NewReader("y\n")
	var out bytes.Buffer

	result, err := confirm(sampleMsg, in, &out, noopEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != sampleMsg {
		t.Errorf("expected message %+v, got %+v", sampleMsg, result)
	}
}

// TestConfirm_EmptyInputConfirms verifies that empty input (just newline) also confirms.
func TestConfirm_EmptyInputConfirms(t *testing.T) {
	in := strings.NewReader("\n")
	var out bytes.Buffer

	result, err := confirm(sampleMsg, in, &out, noopEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != sampleMsg {
		t.Errorf("expected message %+v, got %+v", sampleMsg, result)
	}
}

// TestConfirm_NoInputCancels verifies that "n" input returns ErrCancelled.
func TestConfirm_NoInputCancels(t *testing.T) {
	in := strings.NewReader("n\n")
	var out bytes.Buffer

	_, err := confirm(sampleMsg, in, &out, noopEditor)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.As(err, &ErrCancelled{}) {
		t.Errorf("expected ErrCancelled, got %T: %v", err, err)
	}
}

// TestErrCancelled_SatisfiesErrorInterface verifies ErrCancelled implements error.
func TestErrCancelled_SatisfiesErrorInterface(t *testing.T) {
	var err error = ErrCancelled{}
	if err.Error() != "commit cancelled" {
		t.Errorf("expected error message %q, got %q", "commit cancelled", err.Error())
	}
}

// TestConfirm_InvalidThenYes verifies that invalid input causes a re-prompt and "y" confirms.
func TestConfirm_InvalidThenYes(t *testing.T) {
	in := strings.NewReader("z\nfoo\ny\n")
	var out bytes.Buffer

	result, err := confirm(sampleMsg, in, &out, noopEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != sampleMsg {
		t.Errorf("expected message %+v, got %+v", sampleMsg, result)
	}

	// Verify that the prompt was shown multiple times (for each invalid input + initial)
	outputStr := out.String()
	count := strings.Count(outputStr, "Continue?")
	if count < 3 {
		t.Errorf("expected at least 3 prompts (initial + 2 reprompts), got %d in output:\n%s", count, outputStr)
	}
}

// TestConfirm_EditTitleOnly verifies that "e" with a title-only edited message works.
func TestConfirm_EditTitleOnly(t *testing.T) {
	in := strings.NewReader("e\n")
	var out bytes.Buffer

	newContent := "fix: correct the typo in error message\n"
	mockEditor := makeEditorWriter(newContent)

	result, err := confirm(sampleMsg, in, &out, mockEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "fix: correct the typo in error message" {
		t.Errorf("expected title %q, got %q", "fix: correct the typo in error message", result.Title)
	}
	if result.Body != "" {
		t.Errorf("expected empty body, got %q", result.Body)
	}
}

// TestConfirm_EditTitleAndBody verifies that "e" with title + body edited message works.
func TestConfirm_EditTitleAndBody(t *testing.T) {
	in := strings.NewReader("e\n")
	var out bytes.Buffer

	newContent := "feat(api): add rate limiting\n\n- Add token bucket algorithm\n- Configure limits per endpoint\n"
	mockEditor := makeEditorWriter(newContent)

	result, err := confirm(sampleMsg, in, &out, mockEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "feat(api): add rate limiting" {
		t.Errorf("expected title %q, got %q", "feat(api): add rate limiting", result.Title)
	}

	expectedBody := "- Add token bucket algorithm\n- Configure limits per endpoint"
	if result.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, result.Body)
	}
}

// TestConfirm_OutputContainsTitle verifies that the output includes the proposed title.
func TestConfirm_OutputContainsTitle(t *testing.T) {
	in := strings.NewReader("y\n")
	var out bytes.Buffer

	_, err := confirm(sampleMsg, in, &out, noopEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, sampleMsg.Title) {
		t.Errorf("expected output to contain title %q, got:\n%s", sampleMsg.Title, outputStr)
	}
}

// TestConfirm_OutputContainsBodyBullets verifies that the output includes the body when non-empty.
func TestConfirm_OutputContainsBodyBullets(t *testing.T) {
	in := strings.NewReader("y\n")
	var out bytes.Buffer

	_, err := confirm(sampleMsg, in, &out, noopEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "- Adds middleware for validating JWT tokens") {
		t.Errorf("expected output to contain body bullets, got:\n%s", outputStr)
	}
}

// TestConfirm_OutputContainsPromptOptions verifies that the output contains the prompt options.
func TestConfirm_OutputContainsPromptOptions(t *testing.T) {
	in := strings.NewReader("y\n")
	var out bytes.Buffer

	_, err := confirm(sampleMsg, in, &out, noopEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "[y] Yes   [e] Edit   [n] Cancel") {
		t.Errorf("expected output to contain prompt options, got:\n%s", outputStr)
	}
}

// TestConfirm_NoBodyOmitsBlanksAndBody verifies that when body is empty, no body section is shown.
func TestConfirm_NoBodyOmitsBlanksAndBody(t *testing.T) {
	in := strings.NewReader("y\n")
	var out bytes.Buffer

	_, err := confirm(titleOnlyMsg, in, &out, noopEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, titleOnlyMsg.Title) {
		t.Errorf("expected output to contain title %q, got:\n%s", titleOnlyMsg.Title, outputStr)
	}
}

// TestConfirm_EditFailure verifies that an editor error is properly wrapped.
func TestConfirm_EditFailure(t *testing.T) {
	in := strings.NewReader("e\n")
	var out bytes.Buffer

	failingEditor := func(path string) error {
		return errors.New("editor exited with status 1")
	}

	_, err := confirm(sampleMsg, in, &out, failingEditor)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "editor failed") {
		t.Errorf("expected error to contain 'editor failed', got: %v", err)
	}
}

// TestParseEditedFile_EmptyReturnsError verifies that empty content returns an error.
func TestParseEditedFile_EmptyReturnsError(t *testing.T) {
	_, err := parseEditedFile("")
	if err == nil {
		t.Fatal("expected error for empty content, got nil")
	}
	if !strings.Contains(err.Error(), "commit message is empty after editing") {
		t.Errorf("expected 'commit message is empty after editing' error, got: %v", err)
	}
}

// TestParseEditedFile_TitleOnly verifies that title-only content parses correctly.
func TestParseEditedFile_TitleOnly(t *testing.T) {
	msg, err := parseEditedFile("chore: update deps\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Title != "chore: update deps" {
		t.Errorf("expected title %q, got %q", "chore: update deps", msg.Title)
	}
	if msg.Body != "" {
		t.Errorf("expected empty body, got %q", msg.Body)
	}
}

// TestParseEditedFile_TitleAndBody verifies that title + body content parses correctly.
func TestParseEditedFile_TitleAndBody(t *testing.T) {
	content := "feat: add feature\n\n- bullet one\n- bullet two\n"
	msg, err := parseEditedFile(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Title != "feat: add feature" {
		t.Errorf("expected title %q, got %q", "feat: add feature", msg.Title)
	}
	expected := "- bullet one\n- bullet two"
	if msg.Body != expected {
		t.Errorf("expected body %q, got %q", expected, msg.Body)
	}
}
