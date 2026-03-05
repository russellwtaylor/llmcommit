package commit

import (
	"errors"
	"testing"

	"github.com/russtaylor/llmcommit/internal/ai"
)

// MockRunner records calls and returns configurable output/error.
type MockRunner struct {
	CalledWith []string // the args slice passed to Run
	Output     []byte
	Err        error
}

func (m *MockRunner) Run(name string, args ...string) ([]byte, error) {
	m.CalledWith = append([]string{name}, args...)
	return m.Output, m.Err
}

func TestRunWith_TitleOnly(t *testing.T) {
	mock := &MockRunner{}
	msg := ai.CommitMessage{Title: "feat: add thing"}

	if err := RunWith(mock, msg, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"git", "commit", "-m", "feat: add thing"}
	if !sliceEqual(mock.CalledWith, want) {
		t.Errorf("expected args %v, got %v", want, mock.CalledWith)
	}
}

func TestRunWith_TitleAndBody(t *testing.T) {
	mock := &MockRunner{}
	msg := ai.CommitMessage{Title: "feat: add thing", Body: "- detail"}

	if err := RunWith(mock, msg, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"git", "commit", "-m", "feat: add thing", "-m", "- detail"}
	if !sliceEqual(mock.CalledWith, want) {
		t.Errorf("expected args %v, got %v", want, mock.CalledWith)
	}
}

func TestRunWith_Amend(t *testing.T) {
	mock := &MockRunner{}
	msg := ai.CommitMessage{Title: "feat: add thing"}

	if err := RunWith(mock, msg, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"git", "commit", "-m", "feat: add thing", "--amend"}
	if !sliceEqual(mock.CalledWith, want) {
		t.Errorf("expected args %v, got %v", want, mock.CalledWith)
	}
}

func TestRunWith_ReturnsWrappedErrorOnFailure(t *testing.T) {
	mock := &MockRunner{
		Output: []byte("fatal: not a git repository"),
		Err:    errors.New("exit status 128"),
	}
	msg := ai.CommitMessage{Title: "feat: add thing"}

	err := RunWith(mock, msg, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	wantMsg := "git commit failed: fatal: not a git repository"
	if err.Error() != wantMsg {
		t.Errorf("expected error %q, got %q", wantMsg, err.Error())
	}
}

func TestRunWith_BodyWithNewlines(t *testing.T) {
	mock := &MockRunner{}
	body := "- foo\n- bar"
	msg := ai.CommitMessage{Title: "feat: add thing", Body: body}

	if err := RunWith(mock, msg, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"git", "commit", "-m", "feat: add thing", "-m", "- foo\n- bar"}
	if !sliceEqual(mock.CalledWith, want) {
		t.Errorf("expected args %v, got %v", want, mock.CalledWith)
	}
}

// sliceEqual compares two string slices for equality.
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
