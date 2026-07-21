package telegram

import (
	"fmt"
	"log"
	"os"
	"strings"

	"main/ai"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI

// StartBot menjalankan Telegram Bot
func StartBot() error {

	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	if token == "" {
		return fmt.Errorf(
			"TELEGRAM_BOT_TOKEN belum diatur",
		)
	}

	// Membuat koneksi ke Telegram Bot API
	var err error

	bot, err = tgbotapi.NewBotAPI(token)

	if err != nil {
		return fmt.Errorf(
			"gagal membuat Telegram Bot: %w",
			err,
		)
	}

	log.Printf(
		"Telegram Bot aktif sebagai @%s",
		bot.Self.UserName,
	)

	// Konfigurasi menerima pesan
	config := tgbotapi.NewUpdate(0)

	config.Timeout = 60

	updates := bot.GetUpdatesChan(config)

	// Membaca pesan masuk
	for update := range updates {

		// Abaikan jika bukan pesan
		if update.Message == nil {
			continue
		}

		// Proses pesan
		go handleMessage(update.Message)
	}

	return nil
}

// handleMessage memproses pesan dari pengguna Telegram
func handleMessage(
	message *tgbotapi.Message,
) {

	if message == nil {
		return
	}

	// ID pengguna Telegram
	userID := fmt.Sprintf(
		"%d",
		message.From.ID,
	)

	// Nama pengguna
	username := message.From.UserName

	// Isi pesan asli
	pertanyaan := strings.TrimSpace(
		message.Text,
	)

	log.Printf(
		"Telegram message dari %s (%s): %s",
		userID,
		username,
		pertanyaan,
	)

	// Jika pesan kosong
	if pertanyaan == "" {
		return
	}

	if strings.ToLower(pertanyaan) == "/start" {

		kirimPesan(
			message.Chat.ID,
			"🤖 Halo! Saya adalah bot AI.\n\n"+
				"Silakan langsung kirim pesanmu.\n\n"+
				"Contoh:\n"+
				"Apa itu Golang?\n"+
				"Jelaskan tentang kecerdasan buatan.\n"+
				"Buatkan contoh program Go.",
		)

		return
	}

	if strings.ToLower(pertanyaan) == "/help" {

		kirimPesan(
			message.Chat.ID,
			"📖 Bantuan Bot\n\n"+
				"Kamu cukup mengirim pesan langsung.\n"+
				"Tidak perlu menggunakan [ai].\n\n"+
				"Contoh:\n"+
				"Apa itu Golang?\n"+
				"Jelaskan tentang kecerdasan buatan.\n"+
				"Buatkan contoh program Go.",
		)

		return
	}

	typing := tgbotapi.NewChatAction(
		message.Chat.ID,
		tgbotapi.ChatTyping,
	)

	_, err := bot.Send(typing)

	if err != nil {
		log.Println(
			"gagal mengirim status typing:",
			err,
		)
	}

	jawabanAI := ai.TanyaAi(
		userID,
		pertanyaan,
	)

	kirimPesan(
		message.Chat.ID,
		jawabanAI,
	)
}

func kirimPesan(
	chatID int64,
	isiPesan string,
) {

	msg := tgbotapi.NewMessage(
		chatID,
		isiPesan,
	)

	_, err := bot.Send(msg)

	if err != nil {
		log.Println(
			"gagal mengirim pesan Telegram:",
			err,
		)
	}
}
