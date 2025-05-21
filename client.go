package wsx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
)

var (
	ErrClientHandshake = errors.New("client handshake error")
	ErrSocketError     = errors.New("socket error")
	ErrServerClose     = errors.New("server close error")
)

type WebSocketClient struct {
	conn net.Conn
	host string
}

func NewWebSocketClient(host string) (*WebSocketClient, error) {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}
	return &WebSocketClient{conn: conn, host: host}, nil
}

func (ws *WebSocketClient) Close() error {
	return ws.conn.Close()
}

func (ws *WebSocketClient) Handshake() error {
	handshake := "GET / HTTP/1.1\r\n" +
		"Host: " + ws.host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"\r\n"

	_, err := ws.conn.Write([]byte(handshake))
	if err != nil {
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

func (ws *WebSocketClient) sendFrame(fin bool, opcode Opcode, payload []byte) error {
	var header byte
	if fin {
		header |= 0x80
	}
	header |= byte(opcode)

	if _, err := ws.conn.Write([]byte{header}); err != nil {
		return err
	}

	payloadLen := len(payload)
	if payloadLen < 126 {
		lengthByte := byte(0x80 | payloadLen)
		if _, err := ws.conn.Write([]byte{lengthByte}); err != nil {
			return err
		}
	} else if payloadLen <= 0xFFFF {
		lengthByte := byte(0x80 | 126)
		if _, err := ws.conn.Write([]byte{lengthByte}); err != nil {
			return err
		}
		if err := binary.Write(ws.conn, binary.BigEndian, uint16(payloadLen)); err != nil {
			return err
		}
	} else {
		lengthByte := byte(0x80 | 127)
		if _, err := ws.conn.Write([]byte{lengthByte}); err != nil {
			return err
		}
		if err := binary.Write(ws.conn, binary.BigEndian, uint64(payloadLen)); err != nil {
			return err
		}
	}

	mask := make([]byte, 4)
	for i := range mask {
		mask[i] = byte(rand.Intn(256))
	}
	if _, err := ws.conn.Write(mask); err != nil {
		return err
	}

	maskedPayload := make([]byte, payloadLen)
	for i := range payloadLen {
		maskedPayload[i] = payload[i] ^ mask[i%4]
	}

	_, err := ws.conn.Write(maskedPayload)
	return err
}

func (ws *WebSocketClient) SendMessage(opcode Opcode, payload []byte) error {
	return ws.sendFrame(true, opcode, payload)
}

func (ws *WebSocketClient) readFrame() (*Frame, error) {
	header := make([]byte, 2)
	if _, err := ws.conn.Read(header); err != nil {
		return nil, ErrSocketError
	}

	fin := header[0]&0x80 != 0
	opcode := Opcode(header[0] & 0x0F)

	var payloadLen int
	switch header[1] & 0x7F {
	case 126:
		var extLen uint16
		if err := binary.Read(ws.conn, binary.BigEndian, &extLen); err != nil {
			return nil, ErrSocketError
		}
		payloadLen = int(extLen)
	case 127:
		var extLen uint64
		if err := binary.Read(ws.conn, binary.BigEndian, &extLen); err != nil {
			return nil, ErrSocketError
		}
		payloadLen = int(extLen)
	default:
		payloadLen = int(header[1] & 0x7F)
	}

	mask := make([]byte, 4)
	if header[1]&0x80 != 0 {
		if _, err := ws.conn.Read(mask); err != nil {
			return nil, ErrSocketError
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := ws.conn.Read(payload); err != nil {
		return nil, ErrSocketError
	}

	if header[1]&0x80 != 0 {
		for i := range payloadLen {
			payload[i] ^= mask[i%4]
		}
	}

	return &Frame{Fin: fin, Opcode: opcode, Payload: *bytes.NewBuffer(payload)}, nil
}

func (ws *WebSocketClient) ReadMessage() (*Message, error) {
	var msg Message
	var lastChunk *MessageChunk

	for {
		frame, err := ws.readFrame()
		if err != nil {
			return nil, err
		}

		if frame.Opcode.isControl() {
			switch frame.Opcode {
			case OPCODE_CLOSE:
				return nil, ErrServerClose
			case OPCODE_PING:
				if err := ws.sendFrame(true, OPCODE_PONG, frame.Payload.Bytes()); err != nil {
					return nil, err
				}
			}
		} else {
			chunk := &MessageChunk{Payload: *bytes.NewBuffer(frame.Payload.Bytes())}
			if lastChunk == nil {
				msg.Chunks = chunk
				msg.Opcode = frame.Opcode
			} else {
				lastChunk.Next = chunk
			}
			lastChunk = chunk

			if frame.Fin {
				break
			}
		}
	}

	return &msg, nil
}
