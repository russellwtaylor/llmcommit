package ai

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// CommitMessage holds the structured output from the LLM.
type CommitMessage struct {
	Title string
	Body  string // empty if the change is simple enough not to warrant a body
}

// LLM is the interface all AI providers must satisfy.
type LLM interface {
	GenerateCommit(diff string) (CommitMessage, error)
}

// GeminiLLM implements LLM using the Google Gemini API.
type GeminiLLM struct {
	APIKey string
	Model  string
}

// NewGeminiLLM creates a new GeminiLLM with the given API key and model.
func NewGeminiLLM(apiKey, model string) *GeminiLLM {
	return &GeminiLLM{
		APIKey: apiKey,
		Model:  model,
	}
}

const systemPrompt = `You are an expert software engineer writing a git commit message.

Given a git diff, generate a Conventional Commit message.

Output ONLY this exact format, no markdown, no code fences, no emojis, no commentary:

COMMIT: <type>(<scope>): <short description>

BODY:
- <bullet point>
- <bullet point>

Rules:
- Omit the entire BODY section for simple, self-explanatory changes
- Scope is optional — "type: description" is valid if no scope applies
- Valid types: feat, fix, docs, style, refactor, test, chore, perf, ci, build
- Keep the title (COMMIT line) under 72 characters
- Include BODY only when: multiple files changed, non-trivial logic, new feature, or breaking change
- Output ONLY the specified format — nothing before COMMIT:, nothing after the last bullet point`

// GenerateCommit generates a commit message for the given git diff.
func (g *GeminiLLM) GenerateCommit(diff string) (CommitMessage, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  g.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return CommitMessage{}, fmt.Errorf("gemini API error: failed to create client: %w", err)
	}

	prompt := systemPrompt + "\n\nGit diff:\n" + diff

	result, err := client.Models.GenerateContent(
		ctx,
		g.Model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return CommitMessage{}, fmt.Errorf("gemini API error: %w", err)
	}

	raw := result.Text()

	msg, err := ParseCommitMessage(raw)
	if err != nil {
		return CommitMessage{}, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return msg, nil
}

// ParseCommitMessage parses the LLM response format into a CommitMessage.
// Returns an error if the COMMIT: line is missing or malformed.
func ParseCommitMessage(raw string) (CommitMessage, error) {
	lines := strings.Split(raw, "\n")

	var title string
	var bodyLines []string
	inBody := false
	foundCommit := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !foundCommit {
			if strings.HasPrefix(trimmed, "COMMIT: ") {
				title = strings.TrimPrefix(trimmed, "COMMIT: ")
				foundCommit = true
			}
			continue
		}

		if trimmed == "BODY:" {
			inBody = true
			continue
		}

		if inBody && strings.HasPrefix(trimmed, "- ") {
			bodyLines = append(bodyLines, trimmed)
		}
	}

	if !foundCommit {
		return CommitMessage{}, fmt.Errorf("missing COMMIT: line in LLM response")
	}

	body := strings.Join(bodyLines, "\n")

	return CommitMessage{
		Title: title,
		Body:  body,
	}, nil
}
