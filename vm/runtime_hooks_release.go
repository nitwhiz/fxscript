//go:build !debug

package vm

import "github.com/nitwhiz/fxscript/fx"

func (*Runtime) preStart(*Frame, int) {}

func (*Runtime) preExecute(*Frame, *fx.CommandNode) {}

func (*Runtime) postExecute(*Frame, *fx.CommandNode, int, bool) {}

func (*Runtime) postUnmarshalArgs(any) {}
