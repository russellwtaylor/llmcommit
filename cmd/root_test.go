package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/russellwtaylor/llmcommit/config"
	"github.com/russellwtaylor/llmcommit/internal/ai"
	"github.com/russellwtaylor/llmcommit/internal/ui"
)

// mockLLM is a test double for ai.LLM.
type mockLLM struct {
	response ai.CommitMessage
	err      error
}

func (m *mockLLM) GenerateCommit(diff string) (ai.CommitMessage, error) {
	return m.response, m.err
}

// testCfg returns a minimal Config suitable for tests (no real API calls needed
// because we inject a mock LLM).
func testCfg() *config.Config {
	return &config.Config{
		APIKey: "test-key",
		Model:  "gemini-test",
	}
}

// captureStdout replaces os.Stdout temporarily and returns whatever was printed.
func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	orig := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = orig

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// captureStderr replaces os.Stderr temporarily and returns whatever was printed.
func captureStderr(fn func()) string {
	r, w, _ := os.Pipe()
	orig := os.Stderr
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = orig

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestDryRunTitleOnly verifies that --dry-run prints just the title when there
// is no body, and returns nil without committing.
func TestDryRunTitleOnly(t *testing.T) {
	llm := &mockLLM{response: ai.CommitMessage{Title: "feat(auth): add JWT validation"}}

	// Ensure commitRunFn is never called.
	origCommit := commitRunFn
	commitRunFn = func(msg ai.CommitMessage, amend bool) error {
		t.Fatal("commitRunFn should not be called in dry-run mode")
		return nil
	}
	defer func() { commitRunFn = origCommit }()

	var out string
	var runErr error
	out = captureStdout(func() {
		runErr = runWith(testCfg(), "some diff", llm, true, false)
	})

	if runErr != nil {
		t.Fatalf("expected nil error, got %v", runErr)
	}

	out = strings.TrimRight(out, "\n")
	if out != "feat(auth): add JWT validation" {
		t.Errorf("unexpected output: %q", out)
	}
}

// TestDryRunWithBody verifies that --dry-run prints title, a blank line, then
// the body when a body is present.
func TestDryRunWithBody(t *testing.T) {
	llm := &mockLLM{response: ai.CommitMessage{
		Title: "feat(auth): add JWT validation middleware",
		Body:  "- Adds middleware for validating JWT tokens\n- Extracts user from token claims",
	}}

	origCommit := commitRunFn
	commitRunFn = func(msg ai.CommitMessage, amend bool) error {
		t.Fatal("commitRunFn should not be called in dry-run mode")
		return nil
	}
	defer func() { commitRunFn = origCommit }()

	var out string
	out = captureStdout(func() {
		runWith(testCfg(), "some diff", llm, true, false) //nolint:errcheck
	})

	expected := "feat(auth): add JWT validation middleware\n\n- Adds middleware for validating JWT tokens\n- Extracts user from token claims\n"
	if out != expected {
		t.Errorf("unexpected output:\ngot:  %q\nwant: %q", out, expected)
	}
}

// TestConfirmCancelled verifies that when confirmFn returns ErrCancelled, the
// function prints "Commit cancelled." to stderr and returns nil.
func TestConfirmCancelled(t *testing.T) {
	llm := &mockLLM{response: ai.CommitMessage{Title: "fix: something"}}

	origConfirm := confirmFn
	confirmFn = func(msg ai.CommitMessage) (ai.CommitMessage, error) {
		return ai.CommitMessage{}, ui.ErrCancelled{}
	}
	defer func() { confirmFn = origConfirm }()

	origCommit := commitRunFn
	commitRunFn = func(msg ai.CommitMessage, amend bool) error {
		t.Fatal("commitRunFn should not be called when cancelled")
		return nil
	}
	defer func() { commitRunFn = origCommit }()

	var errOut string
	var runErr error
	errOut = captureStderr(func() {
		runErr = runWith(testCfg(), "some diff", llm, false, false)
	})

	if runErr != nil {
		t.Fatalf("expected nil error on cancellation, got %v", runErr)
	}

	errOut = strings.TrimRight(errOut, "\n")
	if errOut != "Commit cancelled." {
		t.Errorf("unexpected stderr: %q", errOut)
	}
}

// TestConfirmAndCommit verifies that when confirmFn succeeds, commitRunFn is
// called with the returned message.
func TestConfirmAndCommit(t *testing.T) {
	finalMsg := ai.CommitMessage{Title: "chore: update deps", Body: "- bumped go modules"}
	llm := &mockLLM{response: ai.CommitMessage{Title: "chore: update deps", Body: "- bumped go modules"}}

	origConfirm := confirmFn
	confirmFn = func(msg ai.CommitMessage) (ai.CommitMessage, error) {
		return finalMsg, nil
	}
	defer func() { confirmFn = origConfirm }()

	committed := false
	var committedMsg ai.CommitMessage
	var committedAmend bool

	origCommit := commitRunFn
	commitRunFn = func(msg ai.CommitMessage, amend bool) error {
		committed = true
		committedMsg = msg
		committedAmend = amend
		return nil
	}
	defer func() { commitRunFn = origCommit }()

	err := runWith(testCfg(), "some diff", llm, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !committed {
		t.Fatal("expected commitRunFn to be called")
	}
	if committedMsg.Title != finalMsg.Title {
		t.Errorf("committed wrong title: %q", committedMsg.Title)
	}
	if committedAmend {
		t.Error("expected amend=false")
	}
}

// TestLLMError verifies that an error from GenerateCommit propagates.
func TestLLMError(t *testing.T) {
	llm := &mockLLM{err: errors.New("api error")}

	err := runWith(testCfg(), "some diff", llm, false, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "api error") {
		t.Errorf("unexpected error message: %v", err)
	}
}
