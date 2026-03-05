package ai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// FileSummarizer generates a short natural-language summary of a single file's diff.
type FileSummarizer interface {
	SummarizeFile(ctx context.Context, filename, fileDiff string) (string, error)
}

// GeminiFileSummarizer implements FileSummarizer using the Gemini API.
type GeminiFileSummarizer struct {
	APIKey string
	Model  string
}

// NewGeminiFileSummarizer creates a new GeminiFileSummarizer with the given API key and model.
func NewGeminiFileSummarizer(apiKey, model string) *GeminiFileSummarizer {
	return &GeminiFileSummarizer{
		APIKey: apiKey,
		Model:  model,
	}
}

const summarizePrompt = `Summarize what changed in this file's diff in a single concise sentence of at most 20 words.
Output plain prose only — no markdown, no emojis, no bullet points.
Start with the filename followed by a colon.
Example: "internal/auth/jwt.go: adds JWT validation middleware and extracts user from token claims"

File: %s

Diff:
%s`

// SummarizeFile generates a short natural-language summary of the given file's diff.
func (g *GeminiFileSummarizer) SummarizeFile(ctx context.Context, filename, fileDiff string) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  g.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("summarize %s: %w", filename, err)
	}

	prompt := fmt.Sprintf(summarizePrompt, filename, fileDiff)

	result, err := client.Models.GenerateContent(ctx, g.Model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("summarize %s: %w", filename, err)
	}

	return result.Text(), nil
}
