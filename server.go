package wsx

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

var ErrServerHandshake = errors.New("server handshake error")

type WebSocketServer struct {
	listenAddr string
	listener   net.Listener
}

func NewWebSocketServer(listenAddr string) *WebSocketServer {
	return &WebSocketServer{
		listenAddr: listenAddr,
	}
}

func (ws *WebSocketServer) ListenAndServe() error {
	ln, err := net.Listen("tcp", ws.listenAddr)
	if err != nil {
		return err
	}
	ws.listener = ln

	go ws.acceptLoop()
	return nil
}

func (ws *WebSocketServer) acceptLoop() {
	for {
		conn, err := ws.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Println("error accepting connection:", err)
			continue
		}

		fmt.Printf("Accepted connection: %v\n", conn.RemoteAddr())
		wsconn := NewWSConn(conn, false)
		go ws.handleConn(wsconn)
	}
}

func (ws *WebSocketServer) handleConn(wsconn *WSConn) {
	defer wsconn.Close()

	if err := ws.handshake(wsconn); err != nil {
		return
	}

	msg := []byte("random data")
	op := OPCODE_TEXT
	if err := wsconn.SendMessage(op, msg); err != nil {
		return
	}

	time.Sleep(100 * time.Millisecond)
}

func (ws *WebSocketServer) handshake(wsconn *WSConn) error {
	buf := make([]byte, 1024)
	_, err := wsconn.conn.Read(buf)
	if err != nil {
		return err
	}

	headers := make(map[string]string)
	for header := range strings.SplitSeq(string(buf), "\r\n") {
		splitHeader := strings.SplitN(header, ":", 2)
		if len(splitHeader) != 2 {
			continue
		}
		headers[splitHeader[0]] = strings.Trim(splitHeader[1], " ")
	}

	secKey, ok := headers["Sec-WebSocket-Key"]
	if !ok {
		return ErrServerHandshake
	}

	handshake := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + ws.genSecAccept(secKey) + "\r\n" +
		"\r\n"

	if _, err := wsconn.conn.Write([]byte(handshake)); err != nil {
		return err
	}

	return nil
}

func (ws *WebSocketServer) genSecAccept(secKey string) string {
	hash := sha1.Sum(fmt.Appendf(nil, "%s258EAFA5-E914-47DA-95CA-C5AB0DC85B11", secKey))
	return base64.StdEncoding.EncodeToString(hash[:])
}
