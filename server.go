package wsx

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
)

var ErrServerHandshake = errors.New("server handshake error")

type Handler func(wsconn *WSConn)

type WebSocketServer struct {
	listenAddr string
	listener   net.Listener
	handler    Handler
}

func NewWebSocketServer(listenAddr string, handler Handler) *WebSocketServer {
	return &WebSocketServer{
		listenAddr: listenAddr,
		handler:    handler,
	}
}

func (ws *WebSocketServer) ListenAndServe() error {
	ln, err := net.Listen("tcp", ws.listenAddr)
	if err != nil {
		return err
	}
	ws.listener = ln

	for {
		conn, err := ws.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			break
		}
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())
		wsconn := NewWSConn(conn, false)
		go ws.handleConn(wsconn)
	}

	return nil
}

func (ws *WebSocketServer) handleConn(wsconn *WSConn) {
	defer wsconn.Close()

	if err := ws.handshake(wsconn); err != nil {
		fmt.Printf("Failed to complete handshake with %v: %v\n", wsconn.conn.RemoteAddr(), err)
		return
	}

	fmt.Printf("Connection established with %v\n", wsconn.conn.RemoteAddr())

	ws.handler(wsconn)
}

func (ws *WebSocketServer) handshake(wsconn *WSConn) error {
	reader := bufio.NewReader(wsconn.conn)

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	secKey, ok := headers["Sec-WebSocket-Key"]
	if !ok {
		return ErrServerHandshake
	}

	var handshake strings.Builder
	handshake.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	handshake.WriteString("Upgrade: websocket\r\n")
	handshake.WriteString("Connection: Upgrade\r\n")
	handshake.WriteString("Sec-WebSocket-Accept: ")
	handshake.WriteString(ws.genSecAccept(secKey))
	handshake.WriteString("\r\n\r\n")

	if _, err := wsconn.conn.Write([]byte(handshake.String())); err != nil {
		return err
	}

	return nil
}

func (ws *WebSocketServer) genSecAccept(secKey string) string {
	hash := sha1.Sum(fmt.Appendf(nil, "%s258EAFA5-E914-47DA-95CA-C5AB0DC85B11", secKey))
	return base64.StdEncoding.EncodeToString(hash[:])
}
