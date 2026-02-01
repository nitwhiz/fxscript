package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

type CommandHandler func(f *RuntimeFrame, args []fx.ExpressionNode) (jumpTarget int, jump bool)

type Command struct {
	Name    string
	Typ     fx.CommandType
	Handler CommandHandler
}

var BaseCommands = []*Command{
	{"nop", fx.CmdNop, handleNop},
	{"exit", fx.CmdExit, handleExit},
	{"push", fx.CmdPush, handlePush},
	{"pop", fx.CmdPop, handlePop},
	{"goto", fx.CmdGoto, handleGoto},
	{"set", fx.CmdSet, handleSet},
	{"call", fx.CmdCall, handleCall},
	{"ret", fx.CmdRet, handleRet},
	{"jumpIf", fx.CmdJumpIf, handleJumpIf},
}

func (r *Runtime) registerCommand(cmd *Command) {
	if int(cmd.Typ) >= len(r.handlers) {
		newHandlers := make([]CommandHandler, int(cmd.Typ)+1)
		copy(newHandlers, r.handlers)
		r.handlers = newHandlers
	}

	r.handlers[cmd.Typ] = cmd.Handler
}

func (r *Runtime) RegisterCommands(commands []*Command) {
	for _, cmd := range commands {
		r.registerCommand(cmd)
	}
}
