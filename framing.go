package wsx

import "bytes"

type Opcode uint8

const (
	OPCODE_CONT  Opcode = 0x0
	OPCODE_TEXT  Opcode = 0x1
	OPCODE_BIN   Opcode = 0x2
	OPCODE_CLOSE Opcode = 0x8
	OPCODE_PING  Opcode = 0x9
	OPCODE_PONG  Opcode = 0xA
)

func (o Opcode) isControl() bool {
	return 0x8 <= o && o <= 0xF
}

type Frame struct {
	Fin     bool
	Opcode  Opcode
	Payload bytes.Buffer
}

type MessageChunk struct {
	Next    *MessageChunk
	Payload bytes.Buffer
}

type Message struct {
	Opcode Opcode
	Chunks *MessageChunk
}
