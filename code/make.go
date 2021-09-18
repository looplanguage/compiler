package code

import "encoding/binary"

func Make(op OpCode, operands ...int) []byte {
	def, ok := definitions[op]

	// If the opcode doesn't exist, return an empty byte array
	if !ok {
		return []byte{}
	}

	instructionLength := 1

	for _, operandWidth := range def.OperandWidths {
		instructionLength += operandWidth
	}

	instruction := make([]byte, instructionLength)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]

		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}

		offset += width
	}

	return instruction
}
