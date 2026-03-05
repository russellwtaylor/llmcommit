package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_GeminiAPIKeyEnvVar(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-api-key")
	t.Setenv("LLMCOMMIT_MODEL", "")

	// Ensure we run from a temp dir to avoid picking up any local .llmcommit.yaml
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "test-api-key" {
		t.Errorf("expected APIKey %q, got %q", "test-api-key", cfg.APIKey)
	}
}

func TestLoad_LLMCommitModelEnvVar(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-api-key")
	t.Setenv("LLMCOMMIT_MODEL", "gemini-1.5-pro")

	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Model != "gemini-1.5-pro" {
		t.Errorf("expected Model %q, got %q", "gemini-1.5-pro", cfg.Model)
	}
}

func TestLoad_ModelFlagOverridesEverything(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-api-key")
	t.Setenv("LLMCOMMIT_MODEL", "gemini-1.5-pro")

	tmp := t.TempDir()
	// Write a local config with a model value that should be overridden
	writeYAML(t, filepath.Join(tmp, ".llmcommit.yaml"), "model: gemini-1.0-pro\napi_key: from-file\n")

	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	cfg, err := Load("gemini-ultra")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Model != "gemini-ultra" {
		t.Errorf("expected Model %q, got %q", "gemini-ultra", cfg.Model)
	}
}

func TestLoad_ProjectLocalConfigFile(t *testing.T) {
	// Clear env vars so they don't interfere
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("LLMCOMMIT_MODEL", "")

	tmp := t.TempDir()
	writeYAML(t, filepath.Join(tmp, ".llmcommit.yaml"), "model: gemini-from-file\napi_key: file-api-key\n")

	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "file-api-key" {
		t.Errorf("expected APIKey %q, got %q", "file-api-key", cfg.APIKey)
	}
	if cfg.Model != "gemini-from-file" {
		t.Errorf("expected Model %q, got %q", "gemini-from-file", cfg.Model)
	}
}

func TestLoad_DefaultModel(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-api-key")
	t.Setenv("LLMCOMMIT_MODEL", "")

	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Model != "gemini-2.0-flash" {
		t.Errorf("expected default Model %q, got %q", "gemini-2.0-flash", cfg.Model)
	}
}

func TestLoad_ErrorWhenAPIKeyEmpty(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("LLMCOMMIT_MODEL", "")

	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error when api_key is empty, got nil")
	}
	want := "api key is required: set GEMINI_API_KEY or api_key in config"
	if err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
}

// writeYAML is a helper that writes content to a file, failing the test on error.
func writeYAML(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeYAML: %v", err)
	}
}
