package main

import (
	"fmt"
	"os"
	"port-digger/llm"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/llmtest/main.go \"<command>\"")
		fmt.Println("Example: go run cmd/llmtest/main.go \"node /opt/homebrew/bin/claude-code-ui --database-path /Users/xxx/.config/claude-code-ui/db.db\"")
		os.Exit(1)
	}

	// Join all arguments as the command (in case of spaces)
	command := strings.Join(os.Args[1:], " ")

	// Load config
	config, err := llm.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if !config.LLM.Enabled {
		fmt.Println("LLM is disabled in config. Enable it in ~/.config/port-digger/config.yaml")
		os.Exit(1)
	}

	if config.LLM.APIKey == "" {
		fmt.Println("API key is empty. Set it in ~/.config/port-digger/config.yaml")
		os.Exit(1)
	}

	fmt.Printf("Config: URL=%s, Model=%s\n", config.LLM.URL, config.LLM.Model)
	fmt.Printf("Command: %s\n", command)
	fmt.Println("---")

	// Create client and call LLM
	client := llm.NewClient(&config.LLM)
	result, err := client.RewriteProcessName(command)
	if err != nil {
		fmt.Printf("LLM Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Result: %s\n", result)
}
