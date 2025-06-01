package wsx

import (
	"testing"
	"time"
)

const serverURL = "localhost:6971"

func TestWebSocketServer(t *testing.T) {
	ws := NewWebSocketServer(serverURL)
	err := ws.ListenAndServe()
	if err != nil {
		t.Errorf("failed to start server")
	}

	client := NewWebSocketClient(serverURL)
	if err := client.Connect(); err != nil {
		t.Errorf("failed to connect: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	resp, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	respMsp := resp.Chunks.Payload.String()
	expectedMsg := "random data"
	if respMsp != expectedMsg {
		t.Errorf("expected response %s, got %s", expectedMsg, respMsp)
	}
}
