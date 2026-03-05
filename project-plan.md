# llmcommit — Project Plan

## Overview

A production-quality Go CLI tool that generates AI-powered [Conventional Commits](https://www.conventionalcommits.org/) from staged Git diffs.

---

## User Flow

```
git add .
llmcommit
```

1. Read staged diff (`git diff --cached`)
2. Send diff to Gemini API
3. Display proposed commit message
4. Prompt user: **[y] Confirm / [e] Edit / [n] Cancel**
5. On confirm: run `git commit -m "<message>"`
6. On edit: open `$EDITOR`, pre-fill message, capture modified output

---

## Project Structure

```
llmcommit/
├── cmd/
│   └── root.go          # Cobra root command + CLI flags
├── internal/
│   ├── git/
│   │   └── diff.go      # Read staged diff, per-file summarization logic
│   ├── ai/
│   │   └── generate.go  # LLM interface + Gemini implementation
│   ├── commit/
│   │   └── commit.go    # Execute git commit, --amend support
│   └── ui/
│       └── prompt.go    # Interactive confirm/edit/cancel prompt
├── config/
│   └── config.go        # Viper config + env var loading
├── main.go
├── go.mod
├── Makefile
├── .llmcommit.yaml.example
└── README.md
```

---

## Tech Stack

| Concern         | Choice                                  |
| --------------- | --------------------------------------- |
| Language        | Go                                      |
| CLI framework   | [cobra](https://github.com/spf13/cobra) |
| Config          | [viper](https://github.com/spf13/viper) |
| Git interaction | `os/exec` (no go-git)                   |
| LLM provider    | Google Gemini API                       |
| Default model   | `gemini-2.0-flash`                      |
| API key source  | `GEMINI_API_KEY` env var or config file |

---

## Configuration

Config is loaded in this priority order (highest to lowest):

1. CLI flags
2. Project-local config (`.llmcommit.yaml` in repo root)
3. Global config (`~/.llmcommit.yaml`)
4. Environment variables

Example config file (`.llmcommit.yaml.example`):

```yaml
model: gemini-2.0-flash
api_key: "" # or set GEMINI_API_KEY env var
```

---

## Architecture

### LLM Interface

```go
type LLM interface {
    GenerateCommit(diff string) (CommitMessage, error)
}

type CommitMessage struct {
    Title string
    Body  string  // empty string if not warranted
}
```

Gemini is the initial concrete implementation. The interface allows future providers (OpenAI, Anthropic, Ollama, etc.).

### Body Generation Logic

The body (bullet points) is only generated when the diff is complex enough to warrant it. Complexity signals include:

- Multiple files changed
- Non-trivial logic changes (not just formatting/renaming)
- New features or breaking changes detected

The LLM prompt instructs the model to omit the body for simple, self-explanatory changes.

### Large Diff Handling

If the diff exceeds ~12k tokens, the tool does **not** truncate blindly. Instead:

1. Split the diff by file
2. For each file, generate a short natural-language summary (`os/exec` context is preserved)
3. Pass the summarized representation to the LLM instead of the raw diff

### LLM Prompt Contract

The prompt instructs the model to output **only** this format (no markdown, no commentary):

```
COMMIT: <type>(<scope>): <short description>

BODY:
- <bullet>
- <bullet>
```

- Body section is omitted entirely for simple changes
- No emojis in output
- The tool parses `COMMIT:` and `BODY:` sections

---

## CLI Flags (MVP)

| Flag        | Description                                      |
| ----------- | ------------------------------------------------ |
| `--dry-run` | Show generated message, do not commit            |
| `--amend`   | Run `git commit --amend` instead of a new commit |
| `--model`   | Override default model (e.g. `gemini-1.5-pro`)   |

---

## Functional Requirements

### 1. Staged Diff

- Run `git diff --cached`
- Exit with helpful message if diff is empty
- If diff exceeds ~12k tokens, summarize per file before sending to LLM

### 2. AI Generation

- Clean interface with Gemini implementation
- Return structured `CommitMessage{Title, Body}`
- Generate body only when change complexity warrants it
- Handle API errors gracefully

### 3. Commit Preview UX

```
Proposed commit:

feat(auth): add JWT validation middleware

- Adds middleware for validating JWT tokens
- Extracts user from token claims
- Adds unit tests

Continue?
[y] Yes   [e] Edit   [n] Cancel
```

### 4. Commit Execution

```
git commit -m "<title>" -m "<body>"
```

- No body -> omit second `-m`
- `--amend` flag appends `--amend` to the git command

---

## Edge Cases to Handle

- No staged files
- Binary files in diff
- Git not installed or not in a repo
- Diff exceeds token limit (handle via per-file summarization)
- API failure / timeout
- LLM returns malformed output

---

## Build & Install

### Makefile targets

| Target         | Action                            |
| -------------- | --------------------------------- |
| `make build`   | Build binary to `./bin/llmcommit` |
| `make install` | Run `go install`                  |
| `make clean`   | Remove build artifacts            |

### go install

```
go install github.com/<user>/llmcommit@latest
```

README will include both `go install` instructions and Makefile usage.

---

## Output Deliverables

- [ ] Full working Go scaffold
- [ ] `go.mod` with pinned dependencies
- [ ] `.llmcommit.yaml.example` config file
- [ ] `Makefile` with `build`, `install`, `clean` targets
- [ ] `README.md` with usage, config, and build instructions
- [ ] `go build` succeeds with no errors

---

## Future Scope (Post-MVP)

- Additional LLM providers (OpenAI, Anthropic, Ollama)
- Git hook mode (`prepare-commit-msg`)
- Breaking change detection (`BREAKING CHANGE:` footer)
- Token usage reporting
- Scope inference from changed file paths
