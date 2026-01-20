package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

var _ Environment = (*RuntimeFrame)(nil)

type RuntimeFrame struct {
	Environment
	*Runtime

	pc int

	sp    int
	stack []int
}

func (f *RuntimeFrame) pushStack(v int) {
	f.stack[f.sp] = v
	f.sp++
}

func (f *RuntimeFrame) popStack() (int, bool) {
	if f.sp == 0 {
		return 0, false
	}

	f.sp--
	return f.stack[f.sp], true
}

func (f *RuntimeFrame) ExecuteCommand(cmd *fx.CommandNode) (pc int, jump bool, err error) {
	pc, jump = f.handlers[cmd.Type](f, cmd.Args)
	return
}
