package wsx

import (
	"testing"
)

func TestWebSocketServer(t *testing.T) {
	addr := "localhost:6971"
	ws := NewWebSocketServer(addr)
	ws.ListenAndServe()

	client := NewWebSocketClient(addr)
	if err := client.Connect(); err != nil {
		t.Errorf("failed to connect: %v", err)
	}
}
