package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const ClaudeAPI = "https://api.anthropic.com/v1/messages"

type Claude struct {
	apiKey string
	model  string
	client *http.Client
}

func NewClaude() *Claude {
	model := os.Getenv("CLAUDE_MODEL")

	if model == "" {
		model = "claude-3-5-sonnet-latest"
	}

	return &Claude{
		apiKey: os.Getenv("ANTHROPIC_API_KEY"),
		model:  model,
		client: &http.Client{},
	}
}

func (c *Claude) Name() string {
	return "claude"
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`

	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Claude) Chat(
	ctx context.Context,
	messages []Message,
) (string, error) {

	if c.apiKey == "" {
		return "", fmt.Errorf(
			"ANTHROPIC_API_KEY belum diatur",
		)
	}

	claudeMessages := make(
		[]claudeMessage,
		0,
		len(messages),
	)

	for _, message := range messages {

		role := message.Role

		// Claude menggunakan user dan assistant.
		// System message tidak dimasukkan ke messages.
		if role == "system" {
			continue
		}

		if role != "user" && role != "assistant" {
			role = "user"
		}

		claudeMessages = append(
			claudeMessages,
			claudeMessage{
				Role:    role,
				Content: message.Content,
			},
		)
	}

	requestBody := claudeRequest{
		Model:     c.model,
		MaxTokens: 4096,
		Messages:  claudeMessages,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(
			"gagal membuat request Claude: %w",
			err,
		)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ClaudeAPI,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf(
			"gagal membuat HTTP request Claude: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"x-api-key",
		c.apiKey,
	)

	req.Header.Set(
		"anthropic-version",
		"2023-06-01",
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf(
			"gagal menghubungi Claude: %w",
			err,
		)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf(
			"gagal membaca response Claude: %w",
			err,
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(
			"Claude API error (%d): %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var result claudeResponse

	err = json.Unmarshal(
		responseBody,
		&result,
	)

	if err != nil {
		return "", fmt.Errorf(
			"gagal membaca response JSON Claude: %w",
			err,
		)
	}

	if result.Error != nil {
		return "", fmt.Errorf(
			"Claude API error: %s",
			result.Error.Message,
		)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf(
			"Claude tidak memberikan jawaban",
		)
	}

	return result.Content[0].Text, nil
}
