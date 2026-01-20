package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

type RuntimeConfig struct {
	UserCommands []*Command
	Identifiers  fx.IdentifierTable
	StackSize    int
}

func (r *RuntimeConfig) ParserConfig() *fx.ParserConfig {
	commandTypes := fx.CommandTypeTable{}

	for _, cmd := range BaseCommands {
		commandTypes[cmd.Name] = cmd.Typ
	}

	for _, cmd := range r.UserCommands {
		commandTypes[cmd.Name] = cmd.Typ
	}

	return &fx.ParserConfig{
		CommandTypes: commandTypes,
		Identifiers:  r.Identifiers,
	}
}
