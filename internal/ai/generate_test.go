package ai

import (
	"strings"
	"testing"
)

// MockLLM is a test double that implements LLM.
type MockLLM struct {
	Response CommitMessage
	Err      error
}

func (m *MockLLM) GenerateCommit(diff string) (CommitMessage, error) {
	return m.Response, m.Err
}

func TestParseCommitMessage_TitleOnly(t *testing.T) {
	raw := "COMMIT: chore: update dependencies"

	msg, err := ParseCommitMessage(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Title != "chore: update dependencies" {
		t.Errorf("expected title %q, got %q", "chore: update dependencies", msg.Title)
	}
	if msg.Body != "" {
		t.Errorf("expected empty body, got %q", msg.Body)
	}
}

func TestParseCommitMessage_TitleAndBody(t *testing.T) {
	raw := `COMMIT: feat(auth): add OAuth2 login support

BODY:
- Add OAuth2 client configuration
- Implement token refresh logic
- Update user session handling`

	msg, err := ParseCommitMessage(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Title != "feat(auth): add OAuth2 login support" {
		t.Errorf("expected title %q, got %q", "feat(auth): add OAuth2 login support", msg.Title)
	}

	expectedBody := "- Add OAuth2 client configuration\n- Implement token refresh logic\n- Update user session handling"
	if msg.Body != expectedBody {
		t.Errorf("expected body:\n%q\ngot:\n%q", expectedBody, msg.Body)
	}
}

func TestParseCommitMessage_MissingCommitLine(t *testing.T) {
	raw := `BODY:
- some bullet point`

	_, err := ParseCommitMessage(raw)
	if err == nil {
		t.Fatal("expected error for missing COMMIT: line, got nil")
	}
}

func TestParseCommitMessage_ExtraWhitespace(t *testing.T) {
	raw := `

  COMMIT: fix(parser): handle nil pointer dereference


  BODY:
  - Check for nil before dereferencing
  - Add unit test for nil input

`

	msg, err := ParseCommitMessage(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Title != "fix(parser): handle nil pointer dereference" {
		t.Errorf("expected title %q, got %q", "fix(parser): handle nil pointer dereference", msg.Title)
	}
	if msg.Body == "" {
		t.Error("expected non-empty body")
	}
	if !contains(msg.Body, "- Check for nil before dereferencing") {
		t.Errorf("expected body to contain bullet point, got %q", msg.Body)
	}
}

func TestParseCommitMessage_ScopelessType(t *testing.T) {
	raw := "COMMIT: fix: correct typo in error message"

	msg, err := ParseCommitMessage(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Title != "fix: correct typo in error message" {
		t.Errorf("expected title %q, got %q", "fix: correct typo in error message", msg.Title)
	}
	if msg.Body != "" {
		t.Errorf("expected empty body, got %q", msg.Body)
	}
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
