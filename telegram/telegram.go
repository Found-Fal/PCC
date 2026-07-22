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

// Batas aman panjang pesan Telegram.
// Batas Telegram sekitar 4096 karakter,
// sehingga digunakan 4000 sebagai batas aman.
const maxTelegramMessageLength = 4000

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

	// Perintah /start
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

	// Perintah /help
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

	// Menampilkan status typing
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

	// Meminta jawaban dari AI
	jawabanAI := ai.TanyaAi(
		userID,
		pertanyaan,
	)

	// Mengirim jawaban AI.
	// Jika terlalu panjang, otomatis dipecah
	// menjadi beberapa pesan Telegram.
	kirimPesan(
		message.Chat.ID,
		jawabanAI,
	)
}

// kirimPesan mengirim pesan ke Telegram.
// Jika pesan melebihi batas aman Telegram,
// pesan akan otomatis dipecah menjadi beberapa bagian.
func kirimPesan(
	chatID int64,
	isiPesan string,
) {

	if strings.TrimSpace(isiPesan) == "" {
		return
	}

	// Pecah pesan jika terlalu panjang
	pesanTerbagi := splitMessage(
		isiPesan,
		maxTelegramMessageLength,
	)

	// Kirim setiap bagian secara berurutan
	for i, bagian := range pesanTerbagi {

		msg := tgbotapi.NewMessage(
			chatID,
			bagian,
		)

		_, err := bot.Send(msg)

		if err != nil {
			log.Printf(
				"gagal mengirim pesan Telegram bagian %d/%d: %v",
				i+1,
				len(pesanTerbagi),
				err,
			)

			return
		}
	}
}

// splitMessage memecah pesan panjang menjadi beberapa bagian.
// Pemotongan diusahakan dilakukan pada newline atau spasi
// agar kata tidak terpotong.
func splitMessage(
	message string,
	maxLength int,
) []string {

	if len(message) <= maxLength {
		return []string{message}
	}

	var hasil []string

	for len(message) > maxLength {

		// Ambil bagian awal sesuai batas maksimal
		bagian := message[:maxLength]

		// Cari newline terakhir
		index := strings.LastIndex(
			bagian,
			"\n",
		)

		// Jika tidak ada newline,
		// cari spasi terakhir
		if index <= 0 {
			index = strings.LastIndex(
				bagian,
				" ",
			)
		}

		// Jika tidak ditemukan spasi atau newline,
		// potong langsung pada batas maksimal
		if index <= 0 {
			index = maxLength
		}

		// Ambil bagian pesan
		potongan := strings.TrimSpace(
			message[:index],
		)

		if potongan != "" {
			hasil = append(
				hasil,
				potongan,
			)
		}

		// Lanjutkan ke sisa pesan
		message = strings.TrimSpace(
			message[index:],
		)
	}

	// Tambahkan sisa pesan
	if strings.TrimSpace(message) != "" {
		hasil = append(
			hasil,
			strings.TrimSpace(message),
		)
	}

	return hasil
}
