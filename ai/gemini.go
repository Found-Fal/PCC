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

type Gemini struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGemini() *Gemini {
	model := os.Getenv("GEMINI_MODEL")

	if model == "" {
		model = "gemini-3.6-flash"
	}

	return &Gemini{
		apiKey: os.Getenv("GEMINI_API_KEY"),
		model:  model,
		client: &http.Client{},
	}
}

func (g *Gemini) Name() string {
	return "gemini"
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`

	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (g *Gemini) Chat(
	ctx context.Context,
	messages []Message,
) (string, error) {

	if g.apiKey == "" {
		return "", fmt.Errorf(
			"GEMINI_API_KEY belum diatur",
		)
	}

	contents := make(
		[]geminiContent,
		0,
		len(messages),
	)

	for _, message := range messages {

		role := "user"

		if message.Role == "assistant" {
			role = "model"
		}

		contents = append(
			contents,
			geminiContent{
				Role: role,
				Parts: []geminiPart{
					{
						Text: message.Content,
					},
				},
			},
		)
	}

	requestBody := geminiRequest{
		Contents: contents,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf(
			"gagal membuat request Gemini: %w",
			err,
		)
	}

	apiURL := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.model,
		g.apiKey,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		apiURL,
		bytes.NewBuffer(body),
	)

	if err != nil {
		return "", fmt.Errorf(
			"gagal membuat HTTP request Gemini: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf(
			"gagal menghubungi Gemini: %w",
			err,
		)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf(
			"gagal membaca response Gemini: %w",
			err,
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(
			"Gemini API error (%d): %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var result geminiResponse

	err = json.Unmarshal(
		responseBody,
		&result,
	)

	if err != nil {
		return "", fmt.Errorf(
			"gagal membaca response JSON Gemini: %w",
			err,
		)
	}

	if result.Error != nil {
		return "", fmt.Errorf(
			"Gemini API error: %s",
			result.Error.Message,
		)
	}

	if len(result.Candidates) == 0 {
		return "", fmt.Errorf(
			"Gemini tidak memberikan jawaban",
		)
	}

	if len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf(
			"Gemini response tidak memiliki text",
		)
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}
