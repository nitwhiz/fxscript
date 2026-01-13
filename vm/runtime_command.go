package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

type CommandHandler func(f *RuntimeFrame, args []fx.ExpressionNode) (jumpTarget int, jump bool)

type Command struct {
	Typ     fx.CommandType
	Handler CommandHandler
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

func (r *Runtime) registerBaseCommands() {
	r.RegisterCommands([]*Command{
		{fx.CmdNop, handleNop},
		{fx.CmdHostCall, handleHostCall},
		{fx.CmdGoto, handleGoto},
		{fx.CmdSet, handleSet},
		{fx.CmdAdd, handleAdd},
		{fx.CmdCall, handleCall},
		{fx.CmdRet, handleRet},
		{fx.CmdJumpIf, handleJumpIf},
		{fx.CmdSetFlag, handleSetFlag},
		{fx.CmdClearFlag, handleClearFlag},
		{fx.CmdJumpIfFlag, handleJumpIfFlag},
		{fx.CmdJumpIfNotFlag, handleJumpIfNotFlag},
		{fx.CmdCopy, handleCopy},
	})
}
