package wa

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"gorm.io/gorm"
	"main/ai"
	"main/models"
)

// variabel untuk client whatsapp
var clientWa *whatsmeow.Client
var DB *gorm.DB

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())

		fmt.Println(" => dari saya = ", v.Info.IsFromMe)
		fmt.Println(" => server = ", v.Info.MessageSource.Chat.Server)
		fmt.Println(" => apakah group = ", v.Info.IsGroup)
		fmt.Println(" => apakah broadcast = ", v.Info.IsIncomingBroadcast())

		//filter pesan
		if !v.Info.IsFromMe &&
			v.Info.MessageSource.Chat.Server == "lid" &&
			!v.Info.IsGroup &&
			!v.Info.IsIncomingBroadcast() {

			fmt.Println("PENGIRIM = ", v.Info.Sender.User)
			pesan := v.Message.GetConversation()
			fmt.Println("PESAN = " + pesan)

			//membuat array id_wa
			var id_wa []string
			id_wa = append(id_wa, v.Info.ID)

			//status pesan dibaca
			clientWa.MarkRead(context.Background(), id_wa, time.Now(), v.Info.Chat, v.Info.Sender)

			//pengirim akan menerima status
			clientWa.SubscribePresence(context.Background(), v.Info.Sender)

			//status online
			clientWa.SendPresence(context.Background(), types.PresenceAvailable)

			//jeda 3 detik
			time.Sleep(3 * time.Second)

			//status mengetik
			clientWa.SendChatPresence(context.Background(), v.Info.Sender, types.ChatPresenceComposing, types.ChatPresenceMediaText)

			//jeda 3 detik
			time.Sleep(3 * time.Second)

			//status berhenti mengetik
			clientWa.SendChatPresence(context.Background(), v.Info.Sender, types.ChatPresencePaused, types.ChatPresenceMediaText)

			pesanAsli := pesan
			pesan = strings.ToLower(pesan)

			if strings.HasPrefix(pesan, "[ai]") {
				pertanyaan := strings.TrimSpace(pesanAsli[4:])
				if pertanyaan != "" {
					jawabanai := ai.TanyaAi(v.Info.Sender.User, pertanyaan)
					kirimPesanText(v.Info.Sender, jawabanai)
				} else {
					kirimPesanText(v.Info.Sender, "Masukkan pertanyaan setelah prefiks [ai].\nContoh: [ai] Selamat pagi*")
				}
			} else if pesan == "tes" {
				kirimPesan(v.Info.Sender)
			} else {
				kirimPesanDatabase(v.Info.Sender, pesan)
			}
			//untuk uji coba balasan hanya untuk pesan "tes"
			pesan = strings.ToLower(pesan)
			if pesan == "tes" {
				kirimPesan(v.Info.Sender)
			} else {
				kirimPesanDatabase(v.Info.Sender, pesan)
			}
		}
	}
}

func KonekWa(db *gorm.DB) {
	// |------------------------------------------------------------------------------------------------------|
	// | NOTE: You must also import the appropriate DB connector, e.g. github.com/mattn/go-sqlite3 for SQLite |
	// |------------------------------------------------------------------------------------------------------|

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:wa.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	if deviceStore != nil {
		deviceStore.Platform = "macOS"
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	//mengisi variabel client Wa dengan client
	clientWa = client
	DB = db

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}

func kirimPesan(IDPenerima types.JID) {
	clientWa.SendMessage(context.Background(), IDPenerima, &waE2E.Message{Conversation: proto.String("[UJI COBA] - PESAN OTOMATIS")})
}

func kirimPesanDatabase(IDPenerima types.JID, kode string) {
	var pesan models.Pesan
	//mencari berdasarkan primary key (Kode)
	result := DB.Where("kode = ?", kode).First(&pesan)
	if result.Error == nil {
		kirimPesanText(IDPenerima, pesan.Balasan)
	}
}

func kirimPesanText(IDPenerima types.JID, isiPesan string) {
	clientWa.SendMessage(
		context.Background(),
		IDPenerima,
		&waE2E.Message{
			Conversation: proto.String(isiPesan),
		},
	)
}
