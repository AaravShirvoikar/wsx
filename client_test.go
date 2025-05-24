package wsx

import (
	"log"
	"testing"
)

const echoServerURL = "localhost:6970"

func TestWebSocketClient(t *testing.T) {
	client := NewWebSocketClient(echoServerURL)
	if err := client.Connect(); err != nil {
		log.Fatal("failed to connect:", err)
	}

	msg := []byte("random data")
	op := OPCODE_TEXT

	if err := client.SendMessage(op, msg); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	resp, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	if resp.Chunks.Payload.String() != string(msg) {
		t.Errorf("expected response %s, got %s", msg, resp.Chunks.Payload.String())
	}
}
