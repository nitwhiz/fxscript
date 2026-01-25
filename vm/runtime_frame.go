package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

var _ Environment = (*RuntimeFrame)(nil)

type RuntimeFrame struct {
	Environment
	*Runtime

	pc int

	callStackPointer int
	callStack        []int

	operandStackPointer int
	operandStack        []int
}

func (f *RuntimeFrame) setValue(identifier fx.Identifier, value int) {
	if identifier >= fx.VariableOffset {
		f.setMemory(identifier, value)
		return
	}

	f.Environment.Set(identifier, value)
}

func (f *RuntimeFrame) getValue(identifier fx.Identifier) (value int) {
	if identifier >= fx.VariableOffset {
		return f.getMemory(identifier)
	}

	return f.Environment.Get(identifier)
}

func (f *RuntimeFrame) pushCallStack(v int) {
	f.callStack[f.callStackPointer] = v
	f.callStackPointer++
}

func (f *RuntimeFrame) popCallStack() (int, bool) {
	if f.callStackPointer == 0 {
		return 0, false
	}

	f.callStackPointer--
	return f.callStack[f.callStackPointer], true
}

func (f *RuntimeFrame) pushOperandStack(v int) {
	f.operandStack[f.operandStackPointer] = v
	f.operandStackPointer++
}

func (f *RuntimeFrame) popOperandStack() (int, bool) {
	if f.operandStackPointer == 0 {
		return 0, false
	}

	f.operandStackPointer--
	return f.operandStack[f.operandStackPointer], true
}

func (f *RuntimeFrame) ExecuteCommand(cmd *fx.CommandNode) (pc int, jump bool, err error) {
	pc, jump = f.handlers[cmd.Type](f, cmd.Args)
	return
}
