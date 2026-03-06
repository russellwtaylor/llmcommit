package commit

import (
	"fmt"
	"os/exec"

	"github.com/russellwtaylor/llmcommit/internal/ai"
)

// Runner executes a command and returns combined output and error.
type Runner interface {
	Run(name string, args ...string) ([]byte, error)
}

// ExecRunner is the real implementation using os/exec.
type ExecRunner struct{}

// Run executes the named command with the given args and returns combined stdout+stderr output.
func (e ExecRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return output, err
}

// Run executes `git commit` using the default ExecRunner.
func Run(msg ai.CommitMessage, amend bool) error {
	return RunWith(ExecRunner{}, msg, amend)
}

// RunWith executes `git commit` using the provided Runner (for testing).
func RunWith(runner Runner, msg ai.CommitMessage, amend bool) error {
	args := []string{"commit", "-m", msg.Title}

	if msg.Body != "" {
		args = append(args, "-m", msg.Body)
	}

	if amend {
		args = append(args, "--amend")
	}

	output, err := runner.Run("git", args...)
	if err != nil {
		return fmt.Errorf("git commit failed: %s", output)
	}

	return nil
}
