package compiler

import (
	"fmt"
	"github.com/looplanguage/compiler/code"
	"github.com/looplanguage/loop/models/ast"
	"github.com/looplanguage/loop/models/object"
)

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.SuffixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}

			c.emit(code.OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "*":
			c.emit(code.OpMultiply)
		case "/":
			c.emit(code.OpDivide)
		case "-":
			c.emit(code.OpSubtract)
		case "==":
			c.emit(code.OpEquals)
		case "!=":
			c.emit(code.OpNotEquals)
		case ">":
			c.emit(code.OpGreaterThan)
		default:
			return fmt.Errorf("unknown operator: %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.Boolean:
		switch node.Value {
		case true:
			c.emit(code.OpTrue)
		case false:
			c.emit(code.OpFalse)
		}
	case *ast.ConditionalStatement:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpPos := c.emit(code.OpJumpIfNotTrue, 9999)

		err = c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		if node.ElseCondition == nil && node.ElseStatement == nil {
			elsePos := len(c.instructions)
			c.changeOperand(jumpPos, elsePos)
		} else if node.ElseCondition != nil {
			elseJumpPos := c.emit(code.OpJump, 9999)

			afterElsePos := len(c.instructions)
			c.changeOperand(jumpPos, afterElsePos)

			err := c.Compile(node.ElseCondition)
			if err != nil {
				return nil
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}

			afterElsePos = len(c.instructions)
			c.changeOperand(elseJumpPos, afterElsePos)
		} else if node.ElseStatement != nil {
			afterElsePos := len(c.instructions)
			c.changeOperand(jumpPos, afterElsePos)

			err := c.Compile(node.ElseStatement)
			if err != nil {
				return nil
			}
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	fmt.Printf("%+v\n", obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.OpCode == code.OpPop
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) setLastInstruction(op code.OpCode, pos int) {
	previous := c.lastInstruction
	last := EmittedInstruction{OpCode: op, Position: pos}
	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) addInstruction(ins []byte) int {
	position := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return position
}
