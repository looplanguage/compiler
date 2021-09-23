package compiler

import (
	"fmt"
	"github.com/looplanguage/compiler/code"
	"github.com/looplanguage/loop/models/ast"
	"github.com/looplanguage/loop/models/object"
	"sort"
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
	case *ast.String:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.Boolean:
		switch node.Value {
		case true:
			c.emit(code.OpTrue)
		case false:
			c.emit(code.OpFalse)
		}
	case *ast.While:
		startPos := len(c.currentInstructions())
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpPos := c.emit(code.OpJumpIfNotTrue, 9999)

		err = c.Compile(node.Block)
		if err != nil {
			return err
		}

		c.emit(code.OpJump, startPos)
		c.changeOperand(jumpPos, len(c.currentInstructions()))
		c.emit(code.OpNull)
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

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		jumpToEnd := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterConsequencePos)

		if node.ElseCondition == nil && node.ElseStatement == nil {
			c.emit(code.OpNull)
		} else if node.ElseCondition != nil {
			err := c.Compile(node.ElseCondition)
			if err != nil {
				return nil
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		} else if node.ElseStatement != nil {
			err := c.Compile(node.ElseStatement)
			if err != nil {
				return nil
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpToEnd, afterAlternativePos)
	case *ast.BlockStatement:
		c.VariableScope = append(c.VariableScope,
			VariableScope{
				Variables: []Variable{},
				Outer:     &c.VariableScope[c.variableScopeIndex],
			})

		c.variableScopeIndex++
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
		c.variableScopeIndex--
	case *ast.VariableDeclaration:
		if len(c.VariableScope) < c.variableScopeIndex+1 {
			if c.variableScopeIndex != 0 && len(c.VariableScope) > c.variableScopeIndex-1 {
				c.VariableScope = append(c.VariableScope, VariableScope{
					Variables: []Variable{{Name: node.Identifier.Value, Object: &object.Null{}, Index: 0, Scope: c.variableScopeIndex}},
					Outer:     &c.VariableScope[c.variableScopeIndex-1],
				})
			} else {
				c.VariableScope = append(c.VariableScope, VariableScope{
					Variables: []Variable{{Name: node.Identifier.Value, Object: &object.Null{}, Index: 0, Scope: c.variableScopeIndex}},
					Outer:     nil,
				})
			}
		} else {
			c.VariableScope[c.variableScopeIndex].Variables = append(c.VariableScope[c.variableScopeIndex].Variables, Variable{Name: node.Identifier.Value, Object: &object.Null{}, Index: len(c.VariableScope[c.variableScopeIndex].Variables), Scope: c.variableScopeIndex})
		}

		index := len(c.VariableScope[c.variableScopeIndex].Variables) - 1

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		c.emit(code.OpSetVar, c.variableScopeIndex, index)
	case *ast.Assign:
		variable := c.VariableScope[c.variableScopeIndex-1].FindByName(node.Identifier.Value)

		if variable == nil {
			return fmt.Errorf("undefined variable %s", node.Identifier.Value)
		}

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		c.emit(code.OpSetVar, 0, variable.Index)
	case *ast.Identifier:
		fmt.Println(c.variableScopeIndex)
		if len(c.VariableScope)-1 > c.variableScopeIndex {
			variable := c.VariableScope[c.variableScopeIndex].FindByName(node.Value)

			if variable == nil {
				return fmt.Errorf("(scope) undefined variable %s", node.Value)
			}

			c.emit(code.OpGetVar, c.variableScopeIndex, variable.Index)
		} else {
			variable := c.VariableScope[c.variableScopeIndex].FindByName(node.Value)

			if variable == nil {
				return fmt.Errorf("(root) undefined variable %s", node.Value)
			}

			c.emit(code.OpGetVar, c.variableScopeIndex, variable.Index)
		}

		/*
			symbol, ok := c.symbolTable.Resolve(node.Value)

			if !ok {
				return fmt.Errorf("undefined variable %s", node.Value)
			}

			c.loadSymbol(symbol)
		*/
	case *ast.Array:
		for _, element := range node.Elements {
			err := c.Compile(element)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))
	case *ast.IndexExpression:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)
	case *ast.Hashmap:
		keys := []ast.Expression{}
		for k := range node.Values {
			keys = append(keys, k)
		}

		// TODO: Change this for the tests, this affects compiler performance
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Values[k])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Values)*2)
	case *ast.Function:
		c.enterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}

		if !c.lastInstructionIs(code.OpReturnValue) && !c.lastInstructionIs(code.OpReturn) {
			c.emit(code.OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		fmt.Println(c.Bytecode().Instructions.String())
		instructions := c.leaveScope()

		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}

		compiledFunc := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}

		c.emit(code.OpClosure, c.addConstant(compiledFunc), len(freeSymbols))
	case *ast.Return:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, arg := range node.Parameters {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Parameters))
	}

	return nil
}
