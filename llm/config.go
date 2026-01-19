package llm

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the LLM configuration
type Config struct {
	LLM LLMSettings `yaml:"llm"`
}

// LLMSettings contains the LLM-specific settings
type LLMSettings struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
	APIKey  string `yaml:"apikey"`
	Model   string `yaml:"model"`
}

// configDir returns the configuration directory path
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "port-digger"), nil
}

// ConfigPath returns the full path to the config file
func ConfigPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// ensureConfigDir creates the config directory if it doesn't exist
func ensureConfigDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// LoadConfig loads the configuration from disk
// Returns default config if file doesn't exist
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			return &Config{
				LLM: LLMSettings{
					Enabled: false,
					URL:     "https://api.openai.com/v1/chat/completions",
					APIKey:  "",
					Model:   "gpt-4o-mini",
				},
			}, nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// EnsureDefaultConfig creates the default config file if it doesn't exist
func EnsureDefaultConfig() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return nil // File exists, nothing to do
	}

	// Create default config
	defaultConfig := &Config{
		LLM: LLMSettings{
			Enabled: false,
			URL:     "https://api.openai.com/v1/chat/completions",
			APIKey:  "",
			Model:   "gpt-4o-mini",
		},
	}

	return SaveConfig(defaultConfig)
}
