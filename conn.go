package wsx

import (
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"
	"unicode/utf8"
)

var (
	ErrSocketError      = errors.New("socket error")
	ErrConnectionClosed = errors.New("connection closed by peer")
	ErrProtocolError    = errors.New("protocol error")
)

type Conn struct {
	conn     net.Conn
	isClient bool
}

func NewConn(conn net.Conn, isClient bool) *Conn {
	return &Conn{
		conn:     conn,
		isClient: isClient,
	}
}

func (ws *Conn) Addr() net.Addr {
	return ws.conn.RemoteAddr()
}

func (ws *Conn) Close(codes ...uint16) error {
	code := uint16(1000)
	if len(codes) > 0 {
		code = codes[0]
	}
	closePayload := make([]byte, 2)
	closePayload[0] = byte(code >> 8)
	closePayload[1] = byte(code & 0xFF)
	if err := ws.sendFrame(true, OpcodeClose, closePayload); err != nil {
		return err
	}
	return ws.conn.Close()
}

func (ws *Conn) Drop() error {
	return ws.conn.Close()
}

func (ws *Conn) SendMessage(opcode Opcode, payload []byte) error {
	return ws.sendFrame(true, opcode, payload)
}

func (ws *Conn) ReadMessage() (*Message, error) {
	var msg Message
	var messagePayload []byte
	var started bool

	for {
		frame, err := ws.readFrame()
		if err != nil {
			return nil, err
		}

		if frame.opcode.isControl() {
			switch frame.opcode {
			case OpcodeClose:
				closeCode := uint16(1000)
				payloadLen := len(frame.payload)

				if payloadLen == 1 {
					ws.Close(1002)
					return nil, ErrProtocolError
				}

				if payloadLen >= 2 {
					closeCode = uint16(frame.payload[0])<<8 | uint16(frame.payload[1])
					if payloadLen > 2 && !utf8.Valid(frame.payload[2:]) {
						ws.Close(1002)
						return nil, ErrProtocolError
					}
				}

				if !ws.isValidCloseCode(closeCode) {
					ws.Close(1002)
					return nil, ErrProtocolError
				}

				ws.Close(closeCode)
				return nil, ErrConnectionClosed
			case OpcodePing:
				if err := ws.sendFrame(true, OpcodePong, frame.payload); err != nil {
					return nil, err
				}
			}
			continue
		}

		if frame.opcode == OpcodeCont && !started {
			ws.Close(1002)
			return nil, ErrProtocolError
		}

		if frame.opcode != OpcodeCont && started {
			ws.Close(1002)
			return nil, ErrProtocolError
		}

		if !started {
			msg.Opcode = frame.opcode
			started = true
		}
		messagePayload = append(messagePayload, frame.payload...)

		if frame.fin {
			break
		}
	}

	if !utf8.Valid(messagePayload) && msg.Opcode != OpcodeBin {
		ws.Drop()
		return nil, ErrProtocolError
	}

	msg.Payload = messagePayload
	return &msg, nil
}

func (ws *Conn) sendFrame(fin bool, opcode Opcode, payload []byte) error {
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
		lengthByte := byte(payloadLen)
		if ws.isClient {
			lengthByte |= 0x80
		}
		if _, err := ws.conn.Write([]byte{lengthByte}); err != nil {
			return err
		}
	} else if payloadLen <= 0xFFFF {
		lengthByte := byte(126)
		if ws.isClient {
			lengthByte |= 0x80
		}
		if _, err := ws.conn.Write([]byte{lengthByte}); err != nil {
			return err
		}
		if err := binary.Write(ws.conn, binary.BigEndian, uint16(payloadLen)); err != nil {
			return err
		}
	} else {
		lengthByte := byte(127)
		if ws.isClient {
			lengthByte |= 0x80
		}
		if _, err := ws.conn.Write([]byte{lengthByte}); err != nil {
			return err
		}
		if err := binary.Write(ws.conn, binary.BigEndian, uint64(payloadLen)); err != nil {
			return err
		}
	}

	var maskKey []byte
	if ws.isClient {
		maskKey = make([]byte, 4)
		for i := range maskKey {
			maskKey[i] = byte(rand.Intn(256))
		}
		if _, err := ws.conn.Write(maskKey); err != nil {
			return err
		}
		masked := make([]byte, payloadLen)
		for i := range payload {
			masked[i] = payload[i] ^ maskKey[i%4]
		}
		_, err := ws.conn.Write(masked)
		return err
	}

	_, err := ws.conn.Write(payload)
	return err
}

func (ws *Conn) readFrame() (*frame, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(ws.conn, header); err != nil {
		return nil, ErrSocketError
	}

	fin := header[0]&0x80 != 0
	rsv := header[0] & 0x70
	if rsv != 0 {
		ws.Close(1002)
		return nil, ErrProtocolError
	}
	opcode := Opcode(header[0] & 0x0F)

	if opcode.isControl() && !fin {
		ws.Close(1002)
		return nil, ErrProtocolError
	}

	if opcode.isReserved() {
		ws.Drop()
		return nil, ErrProtocolError
	}

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

	if opcode.isControl() && payloadLen > 125 {
		ws.Close(1002)
		return nil, ErrProtocolError
	}

	var mask []byte
	if header[1]&0x80 != 0 {
		mask = make([]byte, 4)
		if _, err := io.ReadFull(ws.conn, mask); err != nil {
			return nil, ErrSocketError
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(ws.conn, payload); err != nil {
		return nil, ErrSocketError
	}

	if mask != nil {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}

	return &frame{
		fin:     fin,
		opcode:  opcode,
		payload: payload,
	}, nil
}

func (ws *Conn) isValidCloseCode(code uint16) bool {
	return (code >= 1000 && code <= 1003) ||
		(code >= 1007 && code <= 1011) ||
		(code >= 3000 && code <= 4999)
}
