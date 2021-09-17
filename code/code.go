package code

type Instructions []byte

type OpCode byte

const (
	OpConstant OpCode = iota
	OpAdd
	OpMultiply
	OpDivide
	OpPop
	OpSubtract
	OpTrue
	OpFalse
	OpEquals
	OpNotEquals
	OpGreaterThan
	OpJump
	OpJumpIfNotTrue
)

var definitions = map[OpCode]*Definition{
	OpConstant:      {"OpConstant", []int{2}},
	OpAdd:           {"OpAdd", []int{}},
	OpMultiply:      {"OpMultiply", []int{}},
	OpDivide:        {"OpDivide", []int{}},
	OpPop:           {"OpPop", []int{}},
	OpSubtract:      {"OpSubtract", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpEquals:        {"OpEquals", []int{}},
	OpNotEquals:     {"OpNotEquals", []int{}},
	OpGreaterThan:   {"OpGreaterThan", []int{}},
	OpJump:          {"OpJump", []int{2}},
	OpJumpIfNotTrue: {"OpJumpIfNotTrue", []int{2}},
}
