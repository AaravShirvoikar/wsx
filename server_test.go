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

	time.Sleep(500 * time.Millisecond)

	var r string
	for {
		resp, err := client.ReadMessage()
		if err != nil {
			t.Fatalf("failed to read message: %v", err)
		}
		if resp != nil {
			r = resp.Chunks.Payload.String()
			break
		}
	}

	msg := "random data"
	if r != msg {
		t.Errorf("expected response %s, got %s", msg, r)
	}
}
