package code

import "fmt"

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[OpCode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[OpCode(op)]

	if !ok {
		return nil, fmt.Errorf("unknown opcode %d", op)
	}

	return def, nil
}
