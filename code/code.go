package code

type Instructions []byte

type OpCode byte

const (
	OpConstant OpCode = iota
)
