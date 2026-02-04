//go:build !debug

package vm

import "github.com/nitwhiz/fxscript/fx"

func (*Runtime) preExecute(*fx.CommandNode) {}

func (*Runtime) postExecute(*fx.CommandNode, int, bool) {}

func (*Runtime) postUnmarshalArgs(any) {}
