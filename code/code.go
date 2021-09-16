package code

type Instructions []byte

type OpCode byte

const (
	OpConstant OpCode = iota
	OpAdd
)

var definitions = map[OpCode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpAdd:      {"OpAdd", []int{}},
}
