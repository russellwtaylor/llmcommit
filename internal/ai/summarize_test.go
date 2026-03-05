package ai

import (
	"testing"
)

func TestNewGeminiFileSummarizer(t *testing.T) {
	summarizer := NewGeminiFileSummarizer("test-api-key", "gemini-2.0-flash")

	if summarizer.APIKey != "test-api-key" {
		t.Errorf("expected APIKey %q, got %q", "test-api-key", summarizer.APIKey)
	}
	if summarizer.Model != "gemini-2.0-flash" {
		t.Errorf("expected Model %q, got %q", "gemini-2.0-flash", summarizer.Model)
	}
}

func TestGeminiFileSummarizer_ImplementsInterface(t *testing.T) {
	// Compile-time check that GeminiFileSummarizer implements FileSummarizer.
	var _ FileSummarizer = (*GeminiFileSummarizer)(nil)
}
