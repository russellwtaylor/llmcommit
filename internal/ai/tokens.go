package ai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// TokenThreshold is the maximum number of tokens before summarization is triggered.
const TokenThreshold = 12000

// TokenCounter counts tokens for a given text using the LLM provider.
type TokenCounter interface {
	CountTokens(ctx context.Context, text string) (int, error)
}

// GeminiTokenCounter implements TokenCounter using the Gemini countTokens API.
type GeminiTokenCounter struct {
	APIKey string
	Model  string
}

// NewGeminiTokenCounter creates a new GeminiTokenCounter with the given API key and model.
func NewGeminiTokenCounter(apiKey, model string) *GeminiTokenCounter {
	return &GeminiTokenCounter{
		APIKey: apiKey,
		Model:  model,
	}
}

// CountTokens counts the number of tokens in the given text using the Gemini API.
func (g *GeminiTokenCounter) CountTokens(ctx context.Context, text string) (int, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  g.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return 0, fmt.Errorf("token count failed: %w", err)
	}

	result, err := client.Models.CountTokens(ctx, g.Model, genai.Text(text), nil)
	if err != nil {
		return 0, fmt.Errorf("token count failed: %w", err)
	}

	return int(result.TotalTokens), nil
}
