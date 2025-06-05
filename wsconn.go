package wsx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"
)

var (
	ErrSocketError   = errors.New("socket error")
	ErrServerClose   = errors.New("server close error")
	ErrProtocolError = errors.New("protocol error")
)

type WSConn struct {
	conn     net.Conn
	isClient bool
}

func NewWSConn(conn net.Conn, isClient bool) *WSConn {
	return &WSConn{
		conn:     conn,
		isClient: isClient,
	}
}

func (w *WSConn) Addr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w *WSConn) Close(codes ...uint16) error {
	code := uint16(1000)
	if len(codes) > 0 {
		code = codes[0]
	}
	closePayload := make([]byte, 2)
	closePayload[0] = byte(code >> 8)
	closePayload[1] = byte(code & 0xFF)
	w.sendFrame(true, OPCODE_CLOSE, closePayload)
	return w.conn.Close()
}

func (w *WSConn) sendFrame(fin bool, opcode Opcode, payload []byte) error {
	var header byte
	if fin {
		header |= 0x80
	}
	header |= byte(opcode)

	if _, err := w.conn.Write([]byte{header}); err != nil {
		return err
	}

	payloadLen := len(payload)
	if payloadLen < 126 {
		lengthByte := byte(payloadLen)
		if w.isClient {
			lengthByte |= 0x80
		}
		if _, err := w.conn.Write([]byte{lengthByte}); err != nil {
			return err
		}
	} else if payloadLen <= 0xFFFF {
		lengthByte := byte(126)
		if w.isClient {
			lengthByte |= 0x80
		}
		if _, err := w.conn.Write([]byte{lengthByte}); err != nil {
			return err
		}
		if err := binary.Write(w.conn, binary.BigEndian, uint16(payloadLen)); err != nil {
			return err
		}
	} else {
		lengthByte := byte(127)
		if w.isClient {
			lengthByte |= 0x80
		}
		if _, err := w.conn.Write([]byte{lengthByte}); err != nil {
			return err
		}
		if err := binary.Write(w.conn, binary.BigEndian, uint64(payloadLen)); err != nil {
			return err
		}
	}

	var maskKey []byte
	if w.isClient {
		maskKey = make([]byte, 4)
		for i := range maskKey {
			maskKey[i] = byte(rand.Intn(256))
		}
		if _, err := w.conn.Write(maskKey); err != nil {
			return err
		}
		masked := make([]byte, payloadLen)
		for i := range payload {
			masked[i] = payload[i] ^ maskKey[i%4]
		}
		_, err := w.conn.Write(masked)
		return err
	}

	_, err := w.conn.Write(payload)
	return err
}

func (w *WSConn) readFrame() (*Frame, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(w.conn, header); err != nil {
		return nil, ErrSocketError
	}

	fin := header[0]&0x80 != 0
	rsv := header[0] & 0x70
	if rsv != 0 {
		w.Close(1002)
		return nil, ErrProtocolError
	}
	opcode := Opcode(header[0] & 0x0F)

	var payloadLen int
	switch header[1] & 0x7F {
	case 126:
		var extLen uint16
		if err := binary.Read(w.conn, binary.BigEndian, &extLen); err != nil {
			return nil, ErrSocketError
		}
		payloadLen = int(extLen)
	case 127:
		var extLen uint64
		if err := binary.Read(w.conn, binary.BigEndian, &extLen); err != nil {
			return nil, ErrSocketError
		}
		payloadLen = int(extLen)
	default:
		payloadLen = int(header[1] & 0x7F)
	}

	if opcode.isControl() && payloadLen > 125 {
		w.Close(1002)
		return nil, ErrProtocolError
	}

	var mask []byte
	if header[1]&0x80 != 0 {
		mask = make([]byte, 4)
		if _, err := io.ReadFull(w.conn, mask); err != nil {
			return nil, ErrSocketError
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(w.conn, payload); err != nil {
		return nil, ErrSocketError
	}

	if mask != nil {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}

	return &Frame{
		Fin:     fin,
		Opcode:  opcode,
		Payload: *bytes.NewBuffer(payload),
	}, nil
}

func (w *WSConn) SendMessage(opcode Opcode, payload []byte) error {
	return w.sendFrame(true, opcode, payload)
}

func (w *WSConn) ReadMessage() (*Message, error) {
	var msg Message
	var messagePayload bytes.Buffer

	for {
		frame, err := w.readFrame()
		if err != nil {
			return nil, err
		}

		if frame.Opcode.isControl() {
			switch frame.Opcode {
			case OPCODE_CLOSE:
				w.sendFrame(true, OPCODE_CLOSE, frame.Payload.Bytes())
				return nil, ErrServerClose
			case OPCODE_PING:
				if err := w.sendFrame(true, OPCODE_PONG, frame.Payload.Bytes()); err != nil {
					return nil, err
				}
			}
			continue
		}

		if messagePayload.Len() == 0 {
			msg.Opcode = frame.Opcode
		}
		messagePayload.Write(frame.Payload.Bytes())

		if frame.Fin {
			break
		}
	}

	msg.Payload = messagePayload
	return &msg, nil
}
