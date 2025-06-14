package main

import (
	"fmt"
	"log"

	"github.com/AaravShirvoikar/wsx"
)

func main() {
	addr := "127.0.0.1:9001"
	client := wsx.NewClient(addr)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	msg := []byte("random data")
	op := wsx.OpcodeText

	if err := client.SendMessage(op, msg); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	resp, err := client.ReadMessage()
	if err != nil {
		log.Fatalf("Failed to read message: %v", err)
	}

	respMsg := string(resp.Payload)
	fmt.Printf("Message received: %v\n", respMsg)

	client.Close()
}
