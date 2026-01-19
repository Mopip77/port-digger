package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"port-digger/logger"
	"strings"
	"time"
)

// Client is the LLM API client
type Client struct {
	httpClient *http.Client
	config     *LLMSettings
}

// NewClient creates a new LLM client
func NewClient(config *LLMSettings) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// ChatMessage represents a message in the chat format
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the request body for the chat API
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatResponse represents the response from the chat API
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// buildPrompt creates the prompt for service name extraction
func buildPrompt(command string) string {
	return fmt.Sprintf(`你是一个命令分析专家。你的任务是从原始命令中提取出简短的服务名称。

规则：
1. 识别命令实际运行的服务或工具名称
2. 输出应该简短，通常是一个单词或短名称
3. 如果无法识别具体服务，输出"未知"
4. 只输出服务名称，不要有任何其他解释

示例：
- 输入: node /opt/homebrew/bin/claude-code-ui --database-path /Users/xxx/.config/claude-code-ui/db.db
- 输出: claude-code-ui

- 输入: /usr/bin/python3 -m http.server 8000
- 输出: http.server

- 输入: /Applications/Antigravity.app/Contents/MacOS/Electron .
- 输出: Antigravity

- 输入: node a.js
- 输出: 未知

现在请分析以下命令：
%s`, command)
}

// RewriteProcessName calls the LLM to extract a service name from the command
func (c *Client) RewriteProcessName(command string) (string, error) {
	logger.Debug("LLM rewrite request started for command: %s", command)

	if c.config.URL == "" || c.config.APIKey == "" {
		err := fmt.Errorf("LLM not configured")
		logger.LogLLMRequest(command, "", err)
		return "", err
	}

	prompt := buildPrompt(command)

	reqBody := ChatRequest{
		Model: c.config.Model,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		err = fmt.Errorf("failed to marshal request: %w", err)
		logger.LogLLMRequest(command, "", err)
		return "", err
	}

	req, err := http.NewRequest("POST", c.config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		logger.LogLLMRequest(command, "", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	logger.Debug("Sending LLM API request to %s with model %s", c.config.URL, c.config.Model)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("request failed: %w", err)
		logger.LogLLMRequest(command, "", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		logger.LogLLMRequest(command, "", err)
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response: %w", err)
		logger.LogLLMRequest(command, "", err)
		return "", err
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		err = fmt.Errorf("failed to parse response: %w", err)
		logger.LogLLMRequest(command, "", err)
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		err = fmt.Errorf("no response choices")
		logger.LogLLMRequest(command, "", err)
		return "", err
	}

	result := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	// Log successful request
	logger.LogLLMRequest(command, result, nil)

	// Also print to console for visibility
	fmt.Printf("LLM Input: %s\n", command)
	fmt.Printf("LLM Output: %s\n", result)

	return result, nil
}
