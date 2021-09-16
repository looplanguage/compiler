package compiler

import (
	"github.com/looplanguage/compiler/code"
	"github.com/looplanguage/loop/models/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func Create() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}
