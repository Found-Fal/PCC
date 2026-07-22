package ai

import (
	"context"
	"fmt"
)

var AI *Router

func InitAI() {
	AI = NewRouter("gemini")

	AI.Register(NewGemini())
	AI.Register(NewClaude())
	AI.Register(NewGroq())

	AI.SetFallback([]string{
		"gemini",
		"claude",
		"groq",
	})
}

// TanyaAi digunakan oleh wa.go
func TanyaAi(
	userID string,
	pertanyaan string,
) string {

	// Pastikan AI sudah diinisialisasi
	if AI == nil {
		return "AI belum diinisialisasi."
	}

	// Buat context untuk request AI
	ctx := context.Background()

	// Pesan yang dikirim ke AI
	messages := []Message{
		{
			Role:    "user",
			Content: pertanyaan,
		},
	}

	// Kirim ke Router
	jawaban, err := AI.Chat(
		ctx,
		messages,
	)

	if err != nil {
		return fmt.Sprintf(
			"Maaf, AI sedang mengalami masalah: %v",
			err,
		)
	}

	return jawaban
}
