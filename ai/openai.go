package ai

import (
	"bufio"
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"os"
	"strings"
)

func MulaiChatAi() {
	client := openai.NewClient(
		option.WithBaseURL("http://localhost:8080/v1/"),
		option.WithAPIKey("saksake-karena-gak-butuh-api-key"),
	)
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	var chatHistory []openai.ChatCompletionMessageParamUnion
	for {
		fmt.Print("\nAnda: ")
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)
		if strings.ToLower(userInput) == "keluar" {
			break
		}
		//Tambahkan input user ke histori
		chatHistory = append(chatHistory, openai.UserMessage(userInput))
		fmt.Print("FikomBot: ")
		//Request Streaming
		stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    openai.ChatModel("diisi-saksake"),
			Messages: chatHistory,
		})
		var fullResponse strings.Builder
		//Loop Membaca Potongan Kata (Tokens)
		for stream.Next() {
			chunk := stream.Current()
			// Ambil text dari delta pilihan pertama
			if len(chunk.Choices) > 0 {
				token := chunk.Choices[0].Delta.Content
				fmt.Print(token)
				fullResponse.WriteString(token)
			}
		}
		if err := stream.Err(); err != nil {
			fmt.Printf("\n[Eror saat streaming: %v]\n", err)
			continue
		}
		//menambahkan riwayat jawaban ke histori agar chatbot tidak lupa
		chatHistory = append(chatHistory, openai.AssistantMessage(fullResponse.String()))
		fmt.Println()
	}
}
