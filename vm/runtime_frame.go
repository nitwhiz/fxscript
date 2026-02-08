package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

var _ Environment = (*Frame)(nil)

type Frame struct {
	Environment
	*Runtime

	pc int

	callStackPointer int
	callStack        []int

	operandStackPointer int
	operandStack        []int
}

func (f *Frame) setValue(identifier fx.Identifier, value int) {
	if identifier >= fx.VariableOffset {
		f.setMemory(identifier, value)
		return
	}

	f.Environment.Set(identifier, value)
}

func (f *Frame) getValue(identifier fx.Identifier) (value int) {
	if identifier >= fx.VariableOffset {
		return f.getMemory(identifier)
	}

	return f.Environment.Get(identifier)
}

func (f *Frame) pushCallStack(v int) {
	f.callStack[f.callStackPointer] = v
	f.callStackPointer++
}

func (f *Frame) popCallStack() (int, bool) {
	if f.callStackPointer == 0 {
		return 0, false
	}

	f.callStackPointer--
	return f.callStack[f.callStackPointer], true
}

func (f *Frame) pushOperandStack(v int) {
	f.operandStack[f.operandStackPointer] = v
	f.operandStackPointer++
}

func (f *Frame) popOperandStack() (int, bool) {
	if f.operandStackPointer == 0 {
		return 0, false
	}

	f.operandStackPointer--
	return f.operandStack[f.operandStackPointer], true
}

func (f *Frame) ExecuteCommand(cmd *fx.CommandNode) (pc int, jump bool, err error) {
	f.preExecute(cmd)

	pc, jump = f.handlers[cmd.Type](f, cmd.Args)

	f.postExecute(cmd, pc, jump)

	return
}

func (f *Frame) resolveIdentifierValue(identifier fx.Identifier) (v any) {
	return f.getValue(identifier)

}

func (f *Frame) Eval(node fx.ExpressionNode) (v any, err error) {
	return f.script.Eval(node, f.resolveIdentifierValue)
}
