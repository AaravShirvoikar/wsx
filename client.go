package wsx

import (
	"errors"
	"net"
)

var ErrClientHandshake = errors.New("client handshake error")

type WebSocketClient struct {
	*WSConn
	host string
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
	handshake := "GET / HTTP/1.1\r\n" +
		"Host: " + ws.host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"\r\n"

	if _, err := ws.conn.Write([]byte(handshake)); err != nil {
		return err
	}

	buf := make([]byte, 1024)
	n, err := ws.conn.Read(buf)
	if err != nil {
		return err
	}

	if n < 2 || buf[n-2] != '\r' || buf[n-1] != '\n' {
		return ErrClientHandshake
	}

	return nil
}
