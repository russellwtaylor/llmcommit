package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// mockTokenCounter is a test double for TokenCounter.
type mockTokenCounter struct {
	count int
	err   error
}

func (m *mockTokenCounter) CountTokens(ctx context.Context, text string) (int, error) {
	return m.count, m.err
}

// mockFileSummarizer is a test double for FileSummarizer.
type mockFileSummarizer struct {
	summary string
	err     error
}

func (m *mockFileSummarizer) SummarizeFile(ctx context.Context, filename, fileDiff string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return filename + ": " + m.summary, nil
}

func TestPrepareDiff_UnderThreshold(t *testing.T) {
	ctx := context.Background()
	diff := "some diff content"
	fileDiffs := map[string]string{
		"main.go": "diff content for main.go",
	}

	counter := &mockTokenCounter{count: TokenThreshold - 1}
	summarizer := &mockFileSummarizer{summary: "summary"}

	result, err := PrepareDiff(ctx, diff, fileDiffs, counter, summarizer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != diff {
		t.Errorf("expected original diff %q, got %q", diff, result)
	}
}

func TestPrepareDiff_AtThreshold(t *testing.T) {
	ctx := context.Background()
	diff := "some diff content"
	fileDiffs := map[string]string{
		"main.go": "diff content for main.go",
	}

	counter := &mockTokenCounter{count: TokenThreshold}
	summarizer := &mockFileSummarizer{summary: "summary"}

	result, err := PrepareDiff(ctx, diff, fileDiffs, counter, summarizer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != diff {
		t.Errorf("expected original diff when at threshold %q, got %q", diff, result)
	}
}

func TestPrepareDiff_OverThreshold_ReturnsSummarized(t *testing.T) {
	ctx := context.Background()
	diff := "some very large diff content"
	fileDiffs := map[string]string{
		"main.go": "diff content for main.go",
	}

	counter := &mockTokenCounter{count: TokenThreshold + 1}
	summarizer := &mockFileSummarizer{summary: "does something important"}

	result, err := PrepareDiff(ctx, diff, fileDiffs, counter, summarizer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == diff {
		t.Error("expected summarized result, got original diff")
	}
}

func TestPrepareDiff_OverThreshold_HasPrefix(t *testing.T) {
	ctx := context.Background()
	diff := "some very large diff content"
	fileDiffs := map[string]string{
		"main.go": "diff content for main.go",
	}

	counter := &mockTokenCounter{count: TokenThreshold + 1}
	summarizer := &mockFileSummarizer{summary: "does something important"}

	result, err := PrepareDiff(ctx, diff, fileDiffs, counter, summarizer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "[Summarized due to large diff]") {
		t.Errorf("expected result to start with [Summarized due to large diff], got %q", result)
	}
}

func TestPrepareDiff_OverThreshold_ContainsFileSummaries(t *testing.T) {
	ctx := context.Background()
	diff := "some very large diff content"
	fileDiffs := map[string]string{
		"internal/auth/jwt.go": "diff for jwt.go",
		"main.go":              "diff for main.go",
	}

	counter := &mockTokenCounter{count: TokenThreshold + 1}
	summarizer := &mockFileSummarizer{summary: "adds new functionality"}

	result, err := PrepareDiff(ctx, diff, fileDiffs, counter, summarizer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for filename := range fileDiffs {
		if !strings.Contains(result, filename) {
			t.Errorf("expected result to contain filename %q, got %q", filename, result)
		}
	}
}

func TestPrepareDiff_OverThreshold_SortedOrder(t *testing.T) {
	ctx := context.Background()
	diff := "some very large diff content"
	fileDiffs := map[string]string{
		"z_last.go":   "diff for z_last.go",
		"a_first.go":  "diff for a_first.go",
		"m_middle.go": "diff for m_middle.go",
	}

	counter := &mockTokenCounter{count: TokenThreshold + 1}
	summarizer := &mockFileSummarizer{summary: "makes changes"}

	result, err := PrepareDiff(ctx, diff, fileDiffs, counter, summarizer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that files appear in sorted order
	aIdx := strings.Index(result, "a_first.go")
	mIdx := strings.Index(result, "m_middle.go")
	zIdx := strings.Index(result, "z_last.go")

	if aIdx == -1 || mIdx == -1 || zIdx == -1 {
		t.Fatalf("not all files found in result: %q", result)
	}
	if !(aIdx < mIdx && mIdx < zIdx) {
		t.Errorf("expected sorted order a < m < z, got positions a=%d, m=%d, z=%d in result:\n%s", aIdx, mIdx, zIdx, result)
	}
}

func TestPrepareDiff_TokenCounterError(t *testing.T) {
	ctx := context.Background()
	diff := "some diff"
	fileDiffs := map[string]string{
		"main.go": "diff content",
	}

	expectedErr := errors.New("token count API error")
	counter := &mockTokenCounter{err: expectedErr}
	summarizer := &mockFileSummarizer{summary: "some summary"}

	_, err := PrepareDiff(ctx, diff, fileDiffs, counter, summarizer)
	if err == nil {
		t.Fatal("expected error from TokenCounter, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestPrepareDiff_SummarizerError(t *testing.T) {
	ctx := context.Background()
	diff := "some very large diff content"
	fileDiffs := map[string]string{
		"main.go": "diff content",
	}

	expectedErr := errors.New("summarize API error")
	counter := &mockTokenCounter{count: TokenThreshold + 1}
	summarizer := &mockFileSummarizer{err: expectedErr}

	_, err := PrepareDiff(ctx, diff, fileDiffs, counter, summarizer)
	if err == nil {
		t.Fatal("expected error from FileSummarizer, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
