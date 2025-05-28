package wsx

import (
	"testing"
	"time"
)

const serverURL = "localhost:6971"

func TestWebSocketServer(t *testing.T) {
	ws := NewWebSocketServer(serverURL)
	ws.ListenAndServe()

	client := NewWebSocketClient(serverURL)
	if err := client.Connect(); err != nil {
		t.Errorf("failed to connect: %v", err)
	}

	time.Sleep(3000)

	resp, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	msg := "random data"
	if resp.Chunks.Payload.String() != msg {
		t.Errorf("expected response %s, got %s", msg, resp.Chunks.Payload.String())
	}
}
