//go:build debug

package vm

import "github.com/nitwhiz/fxscript/fx"

func (r *Runtime) preExecute(cmd *fx.CommandNode) {
	if r.hooks != nil && r.hooks.PreExecute != nil {
		r.hooks.PreExecute(cmd)
	}
}

func (r *Runtime) postExecute(cmd *fx.CommandNode, jumpPc int, jump bool) {
	if r.hooks != nil && r.hooks.PostExecute != nil {
		r.hooks.PostExecute(cmd, jumpPc, jump)
	}
}

func (r *Runtime) postUnmarshalArgs(args any) {
	if r.hooks != nil && r.hooks.PostUnmarshalArgs != nil {
		r.hooks.PostUnmarshalArgs(args)
	}
}
