package wsx

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"strings"
)

var ErrClientHandshake = errors.New("client handshake error")

type Client struct {
	*Conn
	host string
	key  string
}

func NewClient(host string) *Client {
	return &Client{
		host: host,
	}
}

func (ws *Client) Connect() error {
	conn, err := net.Dial("tcp", ws.host)
	if err != nil {
		return err
	}
	ws.Conn = NewConn(conn, true)

	return ws.handshake()
}

func (ws *Client) handshake() error {
	key := make([]byte, 16)
	rand.Read(key)
	ws.key = base64.StdEncoding.EncodeToString(key)

	var handshake strings.Builder
	handshake.WriteString("GET / HTTP/1.1\r\n")
	handshake.WriteString("Host: " + ws.host + "\r\n")
	handshake.WriteString("Upgrade: websocket\r\n")
	handshake.WriteString("Connection: Upgrade\r\n")
	handshake.WriteString("Sec-WebSocket-Key: " + ws.key + "\r\n")
	handshake.WriteString("Sec-WebSocket-Version: 13\r\n")
	handshake.WriteString("\r\n")

	if _, err := ws.conn.Write([]byte(handshake.String())); err != nil {
		return err
	}

	reader := bufio.NewReader(ws.conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.Contains(statusLine, "101") {
		return ErrClientHandshake
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(line) == "" {
			break
		}
	}

	return nil
}
