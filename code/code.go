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
	OpSetLocal
	OpGetLocal
	OpGetBuiltinFunction
	OpClosure
	OpGetFree

	OpSetVar
	OpGetVar
)

var definitions = map[OpCode]*Definition{
	OpConstant:           {"OpConstant", []int{2}},
	OpAdd:                {"OpAdd", []int{}},
	OpMultiply:           {"OpMultiply", []int{}},
	OpDivide:             {"OpDivide", []int{}},
	OpPop:                {"OpPop", []int{}},
	OpSubtract:           {"OpSubtract", []int{}},
	OpTrue:               {"OpTrue", []int{}},
	OpFalse:              {"OpFalse", []int{}},
	OpEquals:             {"OpEquals", []int{}},
	OpNotEquals:          {"OpNotEquals", []int{}},
	OpGreaterThan:        {"OpGreaterThan", []int{}},
	OpJump:               {"OpJump", []int{2}},
	OpJumpIfNotTrue:      {"OpJumpIfNotTrue", []int{2}},
	OpNull:               {"OpNull", []int{}},
	OpSetGlobal:          {"OpSetGlobal", []int{2}},
	OpGetGlobal:          {"OpGetGlobal", []int{2}},
	OpArray:              {"OpArray", []int{2}},
	OpHash:               {"OpHash", []int{2}},
	OpIndex:              {"OpHash", []int{}},
	OpCall:               {"OpCall", []int{1}},
	OpReturn:             {"OpReturn", []int{}},
	OpReturnValue:        {"OpReturnValue", []int{}},
	OpSetLocal:           {"OpSetLocal", []int{1}},
	OpGetLocal:           {"OpGetLocal", []int{1}},
	OpGetBuiltinFunction: {"OpGetBuiltinFunction", []int{1}},
	OpClosure:            {"OpClosure", []int{2, 1}},
	OpGetFree:            {"OpGetFree", []int{1}},

	OpSetVar: {"OpSetVar", []int{2}},
	OpGetVar: {"OpGetVar", []int{2}},
}
