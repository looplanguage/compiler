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
	OpNull
	OpSetGlobal
	OpGetGlobal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturn
	OpReturnValue
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
	OpNull:          {"OpNull", []int{}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}},
	OpHash:          {"OpHash", []int{2}},
	OpIndex:         {"OpHash", []int{}},
	OpCall:          {"OpCall", []int{}},
	OpReturn:        {"OpReturn", []int{}},
	OpReturnValue:   {"OpReturn", []int{}},
}
