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

const GroqAPI = "https://api.groq.com/openai/v1/chat/completions"

type Groq struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGroq() *Groq {
	model := os.Getenv("GROQ_MODEL")

	if model == "" {
		model = "llama-3.3-70b-versatile"
	}

	return &Groq{
		apiKey: os.Getenv("GROQ_API_KEY"),
		model:  model,
		client: &http.Client{},
	}
}

func (g *Groq) Name() string {
	return "groq"
}

type groqRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type groqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (g *Groq) Chat(
	ctx context.Context,
	messages []Message,
) (string, error) {

	if g.apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY belum diatur")
	}

	requestBody := groqRequest{
		Model:    g.model,
		Messages: messages,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("gagal membuat request Groq: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		GroqAPI,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf("gagal membuat HTTP request Groq: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal menghubungi Groq: %w", err)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca response Groq: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(
			"Groq API error (%d): %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var result groqResponse

	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return "", fmt.Errorf(
			"gagal membaca response JSON Groq: %w",
			err,
		)
	}

	if result.Error != nil {
		return "", fmt.Errorf(
			"Groq API error: %s",
			result.Error.Message,
		)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("Groq tidak memberikan jawaban")
	}

	return result.Choices[0].Message.Content, nil
}
