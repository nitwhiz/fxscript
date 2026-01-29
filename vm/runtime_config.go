package vm

import (
	"io/fs"

	"github.com/nitwhiz/fxscript/fx"
)

type RuntimeConfig struct {
	UserCommands     []*Command
	Identifiers      fx.IdentifierTable
	CallStackSize    int
	OperandStackSize int
}

func (r *RuntimeConfig) ParserConfig(fs fs.FS, lookupFn fx.LookupFn) *fx.ParserConfig {
	commandTypes := fx.CommandTypeTable{}

	for _, cmd := range BaseCommands {
		commandTypes[cmd.Name] = cmd.Typ
	}

	for _, cmd := range r.UserCommands {
		commandTypes[cmd.Name] = cmd.Typ
	}

	return &fx.ParserConfig{
		FS:       fs,
		LookupFn: lookupFn,

		CommandTypes: commandTypes,
		Identifiers:  r.Identifiers,
	}
}
