package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/russellwtaylor/llmcommit/internal/ai"
)

// ErrCancelled is returned when the user chooses to cancel.
type ErrCancelled struct{}

func (e ErrCancelled) Error() string { return "commit cancelled" }

// Confirm displays the proposed commit message and prompts the user to
// confirm, edit, or cancel. Returns the final (possibly edited) message,
// or ErrCancelled if the user declines.
func Confirm(msg ai.CommitMessage) (ai.CommitMessage, error) {
	return confirm(msg, os.Stdin, os.Stdout, func(path string) error {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		cmd := exec.Command(editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
}

// confirm is the testable core of Confirm. It reads from in, writes to out,
// and calls openEditor to launch an editor for the "e" option.
func confirm(msg ai.CommitMessage, in io.Reader, out io.Writer, openEditor func(path string) error) (ai.CommitMessage, error) {
	printMessage(msg, out)

	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "> ")
		if !scanner.Scan() {
			// EOF or error — treat as cancel
			return ai.CommitMessage{}, ErrCancelled{}
		}

		input := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch input {
		case "y", "":
			return msg, nil
		case "n":
			return ai.CommitMessage{}, ErrCancelled{}
		case "e":
			edited, err := openEditorFlow(msg, openEditor)
			if err != nil {
				return ai.CommitMessage{}, err
			}
			return edited, nil
		default:
			// Invalid input: re-prompt
			printMessage(msg, out)
		}
	}
}

// printMessage writes the display format to out.
func printMessage(msg ai.CommitMessage, out io.Writer) {
	fmt.Fprintln(out, "Proposed commit:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, msg.Title)
	if msg.Body != "" {
		fmt.Fprintln(out)
		fmt.Fprintln(out, msg.Body)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Continue?")
	fmt.Fprintln(out, "[y] Yes   [e] Edit   [n] Cancel")
}

// openEditorFlow writes msg to a temp file, opens an editor, reads back the result,
// and parses it into a CommitMessage.
func openEditorFlow(msg ai.CommitMessage, openEditor func(path string) error) (ai.CommitMessage, error) {
	// Write the current commit message to a temp file
	tmpFile, err := os.CreateTemp("", "llmcommit-*.txt")
	if err != nil {
		return ai.CommitMessage{}, fmt.Errorf("editor failed: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write title, blank line, then body (if non-empty)
	content := msg.Title
	if msg.Body != "" {
		content += "\n\n" + msg.Body
	}
	if _, err := fmt.Fprint(tmpFile, content); err != nil {
		tmpFile.Close()
		return ai.CommitMessage{}, fmt.Errorf("editor failed: %w", err)
	}
	tmpFile.Close()

	// Open the editor
	if err := openEditor(tmpPath); err != nil {
		return ai.CommitMessage{}, fmt.Errorf("editor failed: %w", err)
	}

	// Read back the edited file
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return ai.CommitMessage{}, fmt.Errorf("editor failed: %w", err)
	}

	return parseEditedFile(string(data))
}

// parseEditedFile parses the content of an edited commit message file.
// First non-empty line = Title; remaining non-empty lines after a blank line = Body.
func parseEditedFile(content string) (ai.CommitMessage, error) {
	lines := strings.Split(content, "\n")

	var title string
	var bodyLines []string
	foundTitle := false
	pastBlank := false

	for _, line := range lines {
		if !foundTitle {
			if strings.TrimSpace(line) != "" {
				title = strings.TrimSpace(line)
				foundTitle = true
			}
			continue
		}

		// We have a title; look for blank line then body
		if !pastBlank {
			if strings.TrimSpace(line) == "" {
				pastBlank = true
			}
			continue
		}

		// Collect body lines (non-empty)
		if strings.TrimSpace(line) != "" {
			bodyLines = append(bodyLines, line)
		}
	}

	if !foundTitle || title == "" {
		return ai.CommitMessage{}, fmt.Errorf("commit message is empty after editing")
	}

	body := strings.Join(bodyLines, "\n")

	return ai.CommitMessage{
		Title: title,
		Body:  body,
	}, nil
}
