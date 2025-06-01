package wsx

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"strings"
)

var ErrClientHandshake = errors.New("client handshake error")

type WebSocketClient struct {
	*WSConn
	host string
	key  string
}

func NewWebSocketClient(host string) *WebSocketClient {
	return &WebSocketClient{
		host: host,
	}
}

func (ws *WebSocketClient) Connect() error {
	conn, err := net.Dial("tcp", ws.host)
	if err != nil {
		return err
	}
	ws.WSConn = NewWSConn(conn, true)

	return ws.handshake()
}

func (ws *WebSocketClient) handshake() error {
	key := make([]byte, 16)
	rand.Read(key)
	ws.key = base64.StdEncoding.EncodeToString(key)

	handshake := "GET / HTTP/1.1\r\n" +
		"Host: " + ws.host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + ws.key + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"\r\n"

	if _, err := ws.conn.Write([]byte(handshake)); err != nil {
		return err
	}

	buf := make([]byte, 1024)
	_, err := ws.conn.Read(buf)
	if err != nil {
		return err
	}

	if strings.HasSuffix(string(buf), "\r\n") {
		return ErrClientHandshake
	}

	return nil
}
