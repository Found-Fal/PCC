package ai

import (
	"context"
	"fmt"
)

var AI *Router

func InitAI() {
	// Groq menjadi provider utama
	AI = NewRouter("groq")

	// Daftarkan semua provider AI
	AI.Register(NewGroq())
	AI.Register(NewClaude())
	AI.Register(NewGemini())

	// Urutan fallback
	AI.SetFallback([]string{
		"groq",
		"claude",
		"gemini",
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
