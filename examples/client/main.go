package main

import (
	"fmt"
	"log"

	"github.com/AaravShirvoikar/wsx"
)

func main() {
	addr := "localhost:9001"
	client := wsx.NewWebSocketClient(addr)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	msg := []byte("random data")
	op := wsx.OPCODE_TEXT

	if err := client.SendMessage(op, msg); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	resp, err := client.ReadMessage()
	if err != nil {
		log.Fatalf("Failed to read message: %v", err)
	}

	respMsg := resp.Payload.String()
	fmt.Printf("Message received: %v\n", respMsg)
}
