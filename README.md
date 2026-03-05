# llmcommit

A CLI that generates AI-powered [Conventional Commits](https://www.conventionalcommits.org/) from staged git diffs.

## Installation

**Via `go install`:**

```sh
go install github.com/russtaylor/llmcommit@latest
```

**From source:**

```sh
git clone https://github.com/russtaylor/llmcommit.git
cd llmcommit
make install
```

## Setup

Export your Gemini API key:

```sh
export GEMINI_API_KEY=your_api_key_here
```

Or set `api_key` in your config file (see [Configuration](#configuration)).

## Usage

Stage your changes and run `llmcommit`:

```sh
git add .
llmcommit
```

`llmcommit` reads the staged diff, calls the Gemini API to generate a Conventional Commit message, and prompts you to confirm:

```
feat(auth): add OAuth2 login support

Implements Google and GitHub OAuth2 providers using the standard
oauth2 package. Adds callback handling and session persistence.

[y] Yes  [e] Edit  [n] Cancel
> _
```

- **y** — commits with the generated message
- **e** — opens the message in your `$EDITOR` before committing
- **n** — cancels without committing

## Configuration

Config files are loaded in this order (highest to lowest priority):

1. CLI flags
2. Environment variables (`GEMINI_API_KEY`)
3. `.llmcommit.yaml` in the current project directory
4. `~/.llmcommit.yaml` (global config)

**Config file fields:**

| Field     | Default             | Description                              |
|-----------|---------------------|------------------------------------------|
| `model`   | `gemini-2.0-flash`  | Gemini model to use for generation       |
| `api_key` | `""`                | Gemini API key (prefer env var instead)  |

**Example config** (copy from `.llmcommit.yaml.example`):

```yaml
# Google Gemini model to use (default: gemini-2.0-flash)
model: gemini-2.0-flash

# API key — prefer setting GEMINI_API_KEY env var instead of storing here
api_key: ""
```

Place this file at `.llmcommit.yaml` in your repo root for project-level config, or at `~/.llmcommit.yaml` for global defaults.

## CLI Flags

| Flag        | Description                                              |
|-------------|----------------------------------------------------------|
| `--dry-run` | Print the generated message without committing           |
| `--amend`   | Amend the previous commit instead of creating a new one  |
| `--model`   | Override the model (e.g. `--model gemini-1.5-pro`)       |

## Building from Source

```sh
# Build binary to ./bin/llmcommit
make build

# Run tests
make test

# Remove build artifacts
make clean
```
