package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	defaultModel = "gemini-2.0-flash"
)

// Config holds all resolved configuration for llmcommit.
type Config struct {
	APIKey string
	Model  string
}

// Load returns a Config resolved from (highest to lowest priority):
// CLI flags > environment variables > project .llmcommit.yaml > ~/.llmcommit.yaml > defaults
func Load(modelFlag string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("model", defaultModel)
	v.SetDefault("api_key", "")

	// Load global config (~/.llmcommit.yaml) first
	home, err := os.UserHomeDir()
	if err == nil {
		v.SetConfigFile(filepath.Join(home, ".llmcommit.yaml"))
		// Ignore error if file doesn't exist
		v.ReadInConfig()
	}

	// Load project-local config (.llmcommit.yaml in cwd), merging on top
	cwd, err := os.Getwd()
	if err == nil {
		localPath := filepath.Join(cwd, ".llmcommit.yaml")
		if _, statErr := os.Stat(localPath); statErr == nil {
			localViper := viper.New()
			localViper.SetConfigFile(localPath)
			if readErr := localViper.ReadInConfig(); readErr == nil {
				// Merge local config values into v, overriding global
				for _, key := range localViper.AllKeys() {
					v.Set(key, localViper.Get(key))
				}
			}
		}
	}

	cfg := &Config{
		APIKey: v.GetString("api_key"),
		Model:  v.GetString("model"),
	}

	// Apply environment variables on top of file config (only when non-empty)
	if envKey := os.Getenv("GEMINI_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}
	if envModel := os.Getenv("LLMCOMMIT_MODEL"); envModel != "" {
		cfg.Model = envModel
	}

	// CLI flag overrides model (highest priority for Model)
	if modelFlag != "" {
		cfg.Model = modelFlag
	}

	// APIKey is required
	if cfg.APIKey == "" {
		return nil, errors.New("api key is required: set GEMINI_API_KEY or api_key in config")
	}

	return cfg, nil
}
