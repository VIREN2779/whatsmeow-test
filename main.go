package main

import (
	"context"   // for context.Context - handles cancellation/timeouts across function calls
	"fmt"       // for Println - printing to terminal
	"os"        // for os.Stdout, os.Signal - OS-level operations
	"os/signal" // for catching Ctrl+C / termination signals
	"strings"   // for strings.Contains / strings.ToLower - case-insensitive text matching
	"syscall"   // for SIGTERM - the actual OS signal constant

	qrterminal "github.com/mdp/qrterminal/v3" // renders QR codes in terminal
	"go.mau.fi/whatsmeow"                     // the core whatsmeow client library
	"go.mau.fi/whatsmeow/store/sqlstore"      // SQLite-backed session storage
	"go.mau.fi/whatsmeow/types"               // shared types like JID, StatusBroadcastJID
	"go.mau.fi/whatsmeow/types/events"        // event structs (Message, Picture, etc.)
	waLog "go.mau.fi/whatsmeow/util/log"      // whatsmeow's logging utility
	"google.golang.org/protobuf/proto"

	_ "modernc.org/sqlite" // underscore import: loads the driver but you never call it directly
)

// client is now package-level so eventHandler (and downloadImage) can use it
// without changing eventHandler's function signature.
var client *whatsmeow.Client

// SendTextMessage sends a local string message to a specific JID
func SendTextMessage(client *whatsmeow.Client, targetJID string, messageText string) {
	jid, err := types.ParseJID(targetJID)
	if err != nil {
		log.Fatalf("Invalid JID: %v", err)
	}

	msg := &waE2E.Message{
		Conversation: proto.String(messageText),
	}

	resp, err := client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		log.Printf("[SEND MESSAGE ERROR]: %v", err)
		return
	}
	fmt.Printf("[MESSAGE SEND SUCCESS] Timestamp: %s\n", resp.Timestamp)
}

func eventHandler(evt any) {
	switch v := evt.(type) {
	case *events.Message:
		// Status/story updates arrive as regular messages sent to the special "status" JID
		if v.Info.Chat == types.StatusBroadcastJID {
			caption := v.Message.GetConversation()
			if img := v.Message.GetImageMessage(); img != nil && caption == "" {
				fmt.Println("[STATUS - IMAGE] From:", v.Info.Sender, "| Caption:", caption)
				caption = img.GetCaption()
			}
			if vid := v.Message.GetVideoMessage(); vid != nil && caption == "" {
				fmt.Println("[STATUS - VIDEO] From:", v.Info.Sender, "| Caption:", caption)
				caption = vid.GetCaption()
			}
		} else {
			// Plain text messages come through GetConversation(); replies / messages with
			// link-preview formatting come through as ExtendedTextMessage instead, so we
			// fall back to that when GetConversation() is empty.
			text := v.Message.GetConversation()
			if text == "" {
				text = v.Message.GetExtendedTextMessage().GetText()
			}

			fmt.Println("[MESSAGE]", text)

			if strings.Contains(strings.ToLower(text), "@start") {
				fmt.Println("[COMMAND] @start detected from:", v.Info.Chat)
				SendTextMessage(client, v.Info.Chat.String(), "Hello, Whatsapp Bot from this side!")
			}
		}

	case *events.Picture:
		if v.Remove {
			fmt.Println("[DP REMOVED] JID:", v.JID, "| By:", v.Author)
		} else {
			fmt.Println("[DP CHANGED] JID:", v.JID, "| By:", v.Author, "| New ID:", v.PictureID)
		}

	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite", "file:examplestore.db?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000", dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("[LOGIN]", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
