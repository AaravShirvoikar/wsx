package wsx

type Opcode uint8

const (
	OpcodeCont  Opcode = 0x0
	OpcodeText  Opcode = 0x1
	OpcodeBin   Opcode = 0x2
	OpcodeClose Opcode = 0x8
	OpcodePing  Opcode = 0x9
	OpcodePong  Opcode = 0xA
)

func (o Opcode) isControl() bool {
	return 0x8 <= o && o <= 0xF
}

func (o Opcode) isReserved() bool {
	return 0x3 <= o && o <= 0x7 || 0xB <= o && o <= 0xF
}

type frame struct {
	fin     bool
	opcode  Opcode
	payload []byte
}

type Message struct {
	Opcode  Opcode
	Payload []byte
}
