package vm

import "github.com/nitwhiz/fxscript/fx"

type Hooks struct {
	PreExecute        func(cmd *fx.CommandNode)
	PostExecute       func(cmd *fx.CommandNode, jumpPc int, jump bool)
	PostUnmarshalArgs func(args any)
}
