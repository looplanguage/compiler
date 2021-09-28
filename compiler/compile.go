package compiler

import (
	"fmt"
	"github.com/looplanguage/compiler/code"
	"github.com/looplanguage/loop/models/ast"
	"github.com/looplanguage/loop/models/object"
	"sort"
)

var jumpReturns []*int

func (c *Compiler) Compile(node ast.Node, root, identifier, previous string) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt, root, identifier, previous)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression, root, "", previous)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.SuffixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right, root, "", previous)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left, root, "", previous)
			if err != nil {
				return err
			}

			c.emit(code.OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left, root, "", previous)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right, root, "", previous)
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
		err := c.Compile(node.Condition, root, "", previous)
		if err != nil {
			return err
		}

		jumpPos := c.emit(code.OpJumpIfNotTrue, 9999)

		err = c.Compile(node.Block, root, "", previous)
		if err != nil {
			return err
		}

		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpNull)
		}

		c.emit(code.OpJump, startPos)
		afterPos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterPos)

		if c.currentScope.Outer == nil {
			skipTo := len(c.currentInstructions())
			for _, jumpReturn := range jumpReturns {
				c.changeOperand(*jumpReturn, skipTo)
			}

			jumpReturns = []*int{}
		}
	case *ast.ConditionalStatement:
		err := c.Compile(node.Condition, root, "", previous)
		if err != nil {
			return err
		}

		jumpPos := c.emit(code.OpJumpIfNotTrue, 9999)

		err = c.Compile(node.Body, root, "", previous)
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
			err := c.Compile(node.ElseCondition, root, "", previous)
			if err != nil {
				return nil
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		} else if node.ElseStatement != nil {
			err := c.Compile(node.ElseStatement, root, "", previous)
			if err != nil {
				return nil
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpToEnd, afterAlternativePos)

		if c.currentScope.Outer == nil {
			skipTo := len(c.currentInstructions())
			for _, jumpReturn := range jumpReturns {
				c.changeOperand(*jumpReturn, skipTo)
			}

			jumpReturns = []*int{}
		}
	case *ast.BlockStatement:
		c.currentScope = c.deeperScope()
		for _, s := range node.Statements {
			err := c.Compile(s, root, "", previous)
			if err != nil {
				return err
			}
		}
		c.currentScope = c.currentScope.Outer

	case *ast.VariableDeclaration:
		index := c.variables

		name := node.Identifier.Value

		if root != "" {
			name = "_INTERNAL_" + root + "" + name
		}

		c.currentScope.Variables[index] = Variable{
			Name:   name,
			Index:  index,
			Object: &object.Null{},
		}

		c.variables++

		err := c.Compile(node.Value, root, "", previous)
		if err != nil {
			return err
		}

		c.emit(code.OpSetVar, index)
	case *ast.Assign:
		variable := c.currentScope.FindByName(node.Identifier.Value, root)

		if variable == nil {
			return fmt.Errorf("undefined variable %s", node.Identifier.Value)
		}

		err := c.Compile(node.Value, root, "", previous)
		if err != nil {
			return err
		}

		c.emit(code.OpSetVar, variable.Index)
	case *ast.Identifier:
		variable := c.currentScope.FindByName(node.Value, root)

		if variable != nil {
			c.emit(code.OpGetVar, variable.Index)
		} else {
			if symbol, ok := c.symbolTable.Resolve(node.Value); ok {
				c.loadSymbol(symbol)
			} else {
				return fmt.Errorf("undefined variable %s", node.Value)
			}
		}
	case *ast.Array:
		for _, element := range node.Elements {
			err := c.Compile(element, root, "", previous)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))
	case *ast.IndexExpression:
		err := c.Compile(node.Value, root, "", previous)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index, root, "", previous)
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
			err := c.Compile(k, root, "", previous)
			if err != nil {
				return err
			}
			err = c.Compile(node.Values[k], root, "", previous)
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

		err := c.Compile(node.Body, root, "", previous)
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
		if c.currentScope.Outer == nil {
			return fmt.Errorf("cannot have return statement in root scope")
		}

		err := c.Compile(node.Value, root, "", previous)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

		if c.scopeIndex == 0 {
			val := c.emit(code.OpJump, 9999)
			jumpReturns = append(jumpReturns, &val)
		}
	case *ast.CallExpression:
		err := c.Compile(node.Function, root, "", previous)
		if err != nil {
			return err
		}

		for _, arg := range node.Parameters {
			err := c.Compile(arg, root, "", previous)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Parameters))
	case *ast.Import:
		err := c.importPackage(root, node)

		if err != nil {
			return err
		}
	case *ast.Export:
		index := c.variables

		c.currentScope.Variables[index] = Variable{
			Name:   identifier,
			Index:  index,
			Object: &object.Null{},
		}

		c.variables++

		err := c.Compile(node.Expression, root, identifier, previous)
		if err != nil {
			return err
		}

		c.emit(code.OpSetVar, index)
	}

	return nil
}
