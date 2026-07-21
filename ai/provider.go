package ai

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Provider interface {
	Name() string
	Chat(ctx context.Context, messages []Message) (string, error)
}
