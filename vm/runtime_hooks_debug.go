//go:build debug

package vm

import "github.com/nitwhiz/fxscript/fx"

func (r *Runtime) preStart(f *Frame, pc int) {
	if r.hooks != nil {
		r.hooks.RunPreStart(f, pc)
	}
}

func (r *Runtime) preExecute(f *Frame, cmd *fx.CommandNode) {
	if r.hooks != nil {
		r.hooks.RunPreExecute(f, cmd)
	}
}

func (r *Runtime) postExecute(f *Frame, cmd *fx.CommandNode, jumpPc int, jump bool) {
	if r.hooks != nil {
		r.hooks.RunPostExecute(f, cmd, jumpPc, jump)
	}
}

func (r *Runtime) postUnmarshalArgs(args any) {
	if r.hooks != nil {
		r.hooks.RunPostUnmarshalArgs(args)
	}
}
