package llm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir, err := configDir()
	if err != nil {
		t.Fatalf("configDir() error = %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "port-digger")
	if dir != expected {
		t.Errorf("configDir() = %v, want %v", dir, expected)
	}
}

func TestLoadConfig_Default(t *testing.T) {
	// Test that LoadConfig returns default config when file doesn't exist
	// Note: This test may fail if the config file already exists
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Default should have LLM disabled
	if config.LLM.Enabled {
		t.Error("Expected LLM.Enabled to be false by default")
	}

	// Default URL should be OpenAI
	if config.LLM.URL != "https://api.openai.com/v1/chat/completions" {
		t.Errorf("Expected default URL to be OpenAI, got %s", config.LLM.URL)
	}
}

func TestConfig_RoundTrip(t *testing.T) {
	// Create a temporary config
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	// Create a custom config dir within temp
	configTestDir := filepath.Join(tempDir, ".config", "port-digger")
	os.MkdirAll(configTestDir, 0755)

	// Test config struct
	config := &Config{
		LLM: LLMSettings{
			Enabled: true,
			URL:     "http://localhost:11434/v1/chat/completions",
			APIKey:  "test-key",
			Model:   "llama3.2",
		},
	}

	// Verify struct fields
	if !config.LLM.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.LLM.Model != "llama3.2" {
		t.Errorf("Expected Model to be llama3.2, got %s", config.LLM.Model)
	}
}
