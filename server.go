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

type Handler func(wsconn *Conn)

type Server struct {
	listenAddr string
	listener   net.Listener
	handler    Handler
}

func NewServer(listenAddr string, handler Handler) *Server {
	return &Server{
		listenAddr: listenAddr,
		handler:    handler,
	}
}

func (ws *Server) ListenAndServe() error {
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
		wsconn := NewConn(conn, false)
		go ws.handleConn(wsconn)
	}

	return nil
}

func (ws *Server) handleConn(wsconn *Conn) {
	defer wsconn.Close()

	if err := ws.handshake(wsconn); err != nil {
		fmt.Printf("Failed to complete handshake with %v: %v\n", wsconn.conn.RemoteAddr(), err)
		return
	}

	fmt.Printf("Connection established with %v\n", wsconn.conn.RemoteAddr())

	ws.handler(wsconn)
}

func (ws *Server) handshake(wsconn *Conn) error {
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
	handshake.WriteString("Sec-WebSocket-Accept: " + ws.genSecAccept(secKey) + "\r\n")
	handshake.WriteString("\r\n")

	if _, err := wsconn.conn.Write([]byte(handshake.String())); err != nil {
		return err
	}

	return nil
}

func (ws *Server) genSecAccept(secKey string) string {
	guid := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	hash := sha1.Sum(fmt.Appendf(nil, "%s%s", secKey, guid))
	return base64.StdEncoding.EncodeToString(hash[:])
}
