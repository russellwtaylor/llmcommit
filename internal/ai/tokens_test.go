package ai

import (
	"testing"
)

func TestNewGeminiTokenCounter(t *testing.T) {
	counter := NewGeminiTokenCounter("test-api-key", "gemini-2.0-flash")

	if counter.APIKey != "test-api-key" {
		t.Errorf("expected APIKey %q, got %q", "test-api-key", counter.APIKey)
	}
	if counter.Model != "gemini-2.0-flash" {
		t.Errorf("expected Model %q, got %q", "gemini-2.0-flash", counter.Model)
	}
}

func TestGeminiTokenCounter_ImplementsInterface(t *testing.T) {
	// Compile-time check that GeminiTokenCounter implements TokenCounter.
	var _ TokenCounter = (*GeminiTokenCounter)(nil)
}
