package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/russellwtaylor/llmcommit/config"
	"github.com/russellwtaylor/llmcommit/internal/ai"
	"github.com/russellwtaylor/llmcommit/internal/commit"
	"github.com/russellwtaylor/llmcommit/internal/git"
	"github.com/russellwtaylor/llmcommit/internal/ui"
)

var (
	dryRun bool
	amend  bool
	model  string
)

// confirmFn is the function used to confirm the commit message with the user.
// It can be replaced in tests to avoid interactive stdin.
var confirmFn = func(msg ai.CommitMessage) (ai.CommitMessage, error) {
	return ui.Confirm(msg)
}

// commitRunFn is the function used to run the git commit.
// It can be replaced in tests to avoid real git operations.
var commitRunFn = func(msg ai.CommitMessage, amend bool) error {
	return commit.Run(msg, amend)
}

var rootCmd = &cobra.Command{
	Use:   "llmcommit",
	Short: "Generate AI-powered Conventional Commits from staged changes",
	RunE:  run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show generated message without committing")
	rootCmd.Flags().BoolVar(&amend, "amend", false, "Amend the previous commit instead of creating a new one")
	rootCmd.Flags().StringVar(&model, "model", "", "Override the LLM model (e.g. gemini-1.5-pro)")
}

func run(cmd *cobra.Command, args []string) error {
	// Suppress cobra's automatic usage and error printing so we control the output format.
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	// Step 1: Load config.
	cfg, err := config.Load(model)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}

	// Step 2: Get staged diff.
	rawDiff, err := git.GetStagedDiff()
	if err != nil {
		if errors.Is(err, git.ErrNoStagedFiles) {
			fmt.Fprintln(os.Stderr, err)
			return nil
		}
		return err
	}

	// Step 3: Prepare diff (token check + optional summarization).
	fileDiffs := git.SplitDiffByFile(rawDiff)
	counter := ai.NewGeminiTokenCounter(cfg.APIKey, cfg.Model)
	summarizer := ai.NewGeminiFileSummarizer(cfg.APIKey, cfg.Model)
	diff, err := ai.PrepareDiff(context.Background(), rawDiff, fileDiffs, counter, summarizer)
	if err != nil {
		return err
	}

	// Step 4: Generate commit message.
	llm := ai.NewGeminiLLM(cfg.APIKey, cfg.Model)

	return runWith(cfg, diff, llm, dryRun, amend)
}

// runWith is the testable core of run. It accepts an already-prepared diff and
// an LLM implementation, making it possible to inject mocks in tests.
func runWith(cfg *config.Config, diff string, llm ai.LLM, dryRun, amend bool) error {
	msg, err := llm.GenerateCommit(diff)
	if err != nil {
		return err
	}

	// Step 5: Dry-run — print message and exit without prompting or committing.
	if dryRun {
		fmt.Println(msg.Title)
		if msg.Body != "" {
			fmt.Println()
			fmt.Println(msg.Body)
		}
		return nil
	}

	// Step 6: Prompt user to confirm/edit/cancel.
	finalMsg, err := confirmFn(msg)
	if err != nil {
		var cancelled ui.ErrCancelled
		if errors.As(err, &cancelled) {
			fmt.Fprintln(os.Stderr, "Commit cancelled.")
			return nil
		}
		return err
	}

	// Step 7: Commit.
	return commitRunFn(finalMsg, amend)
}
