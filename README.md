# llmcommit

A CLI that generates AI-powered [Conventional Commits](https://www.conventionalcommits.org/) from staged git diffs using Google Gemini.

---

## Installation

### Option 1: `go install` (recommended)

Requires Go 1.21+.

```sh
go install github.com/russellwtaylor/llmcommit@latest
```

This installs `llmcommit` to your `$GOPATH/bin`. Make sure that's on your `$PATH`:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Option 2: Build from source

```sh
git clone https://github.com/russellwtaylor/llmcommit.git
cd llmcommit
make install   # runs go install
```

Or build a local binary without installing:

```sh
make build
./bin/llmcommit
```

---

## Setup

### 1. Get a Gemini API key

Go to [aistudio.google.com/apikey](https://aistudio.google.com/apikey) and create a free API key.

### 2. Export the key

Add this to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.) so it persists across sessions:

```sh
export GEMINI_API_KEY=your_api_key_here
```

Then reload your shell:

```sh
source ~/.zshrc   # or ~/.bashrc
```

Alternatively, store the key in a [config file](#configuration) — though the env var approach is preferred to avoid accidentally committing credentials.

---

## Usage

### Basic workflow

```sh
git add .
llmcommit
```

`llmcommit` reads your staged diff, calls the Gemini API, and shows a proposed commit message:

```
Proposed commit:

feat(auth): add OAuth2 login support

- Implements Google and GitHub OAuth2 providers
- Adds callback handling and session persistence
- Stores session token in encrypted cookie

Continue?
[y] Yes   [e] Edit   [n] Cancel
>
```

Press **y** to commit, **e** to edit, or **n** to cancel.

### Preview without committing

Use `--dry-run` to see what llmcommit would generate without touching git:

```sh
git add .
llmcommit --dry-run
```

### Amend the last commit

Stage additional changes (or nothing, if you just want to reword), then:

```sh
git add .
llmcommit --amend
```

This runs `git commit --amend` with the newly generated message.

### Override the model

Use a more powerful model for complex diffs:

```sh
llmcommit --model gemini-1.5-pro
```

### Editing the message

When you press **e**, `llmcommit` opens the proposed message in your `$EDITOR`:

```sh
export EDITOR=vim    # or nano, code --wait, etc.
```

If `$EDITOR` is not set, it falls back to `vi`. Edit the message, save, and quit — `llmcommit` will use the updated text as the commit message.

### Large diffs

If your staged diff is very large (>12k tokens), `llmcommit` automatically summarizes each changed file individually before sending to the API. This keeps costs low and results accurate — no silent truncation.

---

## Configuration

Config is loaded in this priority order (highest wins):

| Priority | Source                                                      |
| -------- | ----------------------------------------------------------- |
| 1        | CLI flags (`--model`, etc.)                                 |
| 2        | Environment variables (`GEMINI_API_KEY`, `LLMCOMMIT_MODEL`) |
| 3        | Project config (`.llmcommit.yaml` in the repo root)         |
| 4        | Global config (`~/.llmcommit.yaml`)                         |

### Config file fields

| Field     | Default            | Description                               |
| --------- | ------------------ | ----------------------------------------- |
| `model`   | `gemini-2.0-flash` | Gemini model to use                       |
| `api_key` | —                  | API key (prefer `GEMINI_API_KEY` env var) |

### Example config file

Copy `.llmcommit.yaml.example` to get started:

```sh
cp .llmcommit.yaml.example .llmcommit.yaml
```

```yaml
# .llmcommit.yaml
model: gemini-2.0-flash
api_key: "" # leave empty and use GEMINI_API_KEY env var instead
```

Place the file at:

- **`.llmcommit.yaml`** in your repo root — applies to that project only
- **`~/.llmcommit.yaml`** — applies globally across all repos

> **Tip:** Add `.llmcommit.yaml` to your `.gitignore` if it contains an API key.

### Environment variables

| Variable          | Description                        |
| ----------------- | ---------------------------------- |
| `GEMINI_API_KEY`  | Your Gemini API key (required)     |
| `LLMCOMMIT_MODEL` | Model override (same as `--model`) |

---

## CLI Flags

| Flag        | Description                                             |
| ----------- | ------------------------------------------------------- |
| `--dry-run` | Print the generated message without committing          |
| `--amend`   | Amend the previous commit instead of creating a new one |
| `--model`   | Override the model (e.g. `--model gemini-1.5-pro`)      |

---

## Commit message format

`llmcommit` always generates [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description>

- Optional bullet point body
- Only included for complex or multi-file changes
```

Supported types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`, `build`.

---

## Building from Source

```sh
# Build binary to ./bin/llmcommit
make build

# Install to $GOPATH/bin
make install

# Run the test suite
make test

# Remove build artifacts
make clean
```

---

## Troubleshooting

**`api key is required`**
Set `GEMINI_API_KEY` in your environment or in a config file. Make sure you've reloaded your shell after editing your profile.

**`no staged files`**
Run `git add <files>` before `llmcommit`.

**`not in a git repository`**
Run `llmcommit` from inside a git repo (`git init` if needed).

**Generated message looks off**
Try `--model gemini-1.5-pro` for better results on complex diffs. You can also press `e` to edit the message before committing.
