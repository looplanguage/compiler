package compiler

import (
	"encoding/gob"
	"github.com/looplanguage/compiler/code"
	"github.com/looplanguage/loop/models/object"
)

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Variable struct {
	Name   string
	Index  int
	Scope  int
	Object object.Object
}

type VariableScope struct {
	Variables map[int]Variable
	Outer     *VariableScope
}

func (vs *VariableScope) FindByName(name, root string) *Variable {
	if root == "" {
		for _, v := range vs.Variables {
			if v.Name == name {
				return &v
			}
		}
	} else {
		for _, v := range vs.Variables {
			if v.Name == "_INTERNAL_"+root+name {
				return &v
			}
			if v.Name == name {
				return &v
			}
		}
	}

	if vs.Outer != nil {
		return vs.Outer.FindByName(name, root)
	}

	return nil
}

type Compiler struct {
	constants   []object.Object
	symbolTable *SymbolTable

	VariableScopes []VariableScope
	currentScope   *VariableScope

	scopes     []CompilationScope
	scopeIndex int

	variables int

	root string
}

type EmittedInstruction struct {
	OpCode   code.OpCode
	Position int
}

func Create() *Compiler {
	globalScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := CreateSymbolTable()

	for i, value := range object.Builtins {
		symbolTable.DefineBuiltin(i, value.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{globalScope},
		scopeIndex:  0,
		variables:   0,
		currentScope: &VariableScope{
			Variables: map[int]Variable{},
			Outer:     nil,
		},
	}
}

func (c *Compiler) deeperScope() *VariableScope {
	return &VariableScope{
		Variables: map[int]Variable{},
		Outer:     c.currentScope,
	}
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltinFunction, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	}
}

func (s *SymbolTable) GetAllVariables(current map[string]Symbol) map[string]Symbol {
	returnVal := current

	for k, v := range s.store {
		returnVal[k] = v
	}

	if s.Outer != nil {
		return s.Outer.GetAllVariables(returnVal)
	}

	return returnVal
}

func CreateWithState(s *SymbolTable, constants []object.Object) *Compiler {
	comp := Create()
	comp.symbolTable = s
	comp.constants = constants

	return comp
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) lastInstructionIs(op code.OpCode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.OpCode == op
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturn))

	c.scopes[c.scopeIndex].lastInstruction.OpCode = code.OpReturn
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	instructions := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = instructions
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) setLastInstruction(op code.OpCode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{OpCode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.OpCode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.symbolTable = CreateEnclosedSymbolTable(c.symbolTable)
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	c.symbolTable = c.symbolTable.Outer

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	return instructions
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	Variables    []VariableScope
}

func RegisterGobTypes() {
	gob.Register(&object.String{})
	gob.Register(&object.Integer{})
	gob.Register(&object.CompiledFunction{})
	gob.Register(&object.Null{})
}
