package ai

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// PrepareDiff checks the token count of diff and returns either:
// - the original diff if under TokenThreshold
// - a summarized representation if over threshold
//
// The summarized form is a list of per-file summaries joined by newlines.
func PrepareDiff(
	ctx context.Context,
	diff string,
	fileDiffs map[string]string,
	counter TokenCounter,
	summarizer FileSummarizer,
) (string, error) {
	count, err := counter.CountTokens(ctx, diff)
	if err != nil {
		return "", err
	}

	if count <= TokenThreshold {
		return diff, nil
	}

	// Sort file keys for deterministic output order
	files := make([]string, 0, len(fileDiffs))
	for filename := range fileDiffs {
		files = append(files, filename)
	}
	sort.Strings(files)

	summaries := make([]string, 0, len(files))
	for _, filename := range files {
		summary, err := summarizer.SummarizeFile(ctx, filename, fileDiffs[filename])
		if err != nil {
			return "", fmt.Errorf("prepare diff: %w", err)
		}
		summaries = append(summaries, summary)
	}

	result := "[Summarized due to large diff]\n" + strings.Join(summaries, "\n")
	return result, nil
}
