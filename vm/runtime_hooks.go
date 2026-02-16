package vm

import (
	"maps"
	"slices"

	"github.com/nitwhiz/fxscript/fx"
)

type PreStartHookFunc func(f *Frame, pc int)

type PreExecuteHookFunc func(f *Frame, cmd *fx.CommandNode)

type PostExecuteHookFunc func(f *Frame, cmd *fx.CommandNode, jumpPc int, jump bool)

type PostUnmarshalArgsHookFunc func(args any)

type HookFunc interface {
	PreStartHookFunc | PreExecuteHookFunc | PostExecuteHookFunc | PostUnmarshalArgsHookFunc
}

type HookCollection[T HookFunc] map[int][]T

func (hc HookCollection[T]) Add(priority int, hook T) {
	if _, ok := hc[priority]; !ok {
		hc[priority] = []T{}
	}

	hc[priority] = append(hc[priority], hook)
}

type Hooks struct {
	PreStart          HookCollection[PreStartHookFunc]
	PreExecute        HookCollection[PreExecuteHookFunc]
	PostExecute       HookCollection[PostExecuteHookFunc]
	PostUnmarshalArgs HookCollection[PostUnmarshalArgsHookFunc]
}

func NewHooks() *Hooks {
	return &Hooks{
		PreStart:          HookCollection[PreStartHookFunc]{},
		PreExecute:        HookCollection[PreExecuteHookFunc]{},
		PostExecute:       HookCollection[PostExecuteHookFunc]{},
		PostUnmarshalArgs: HookCollection[PostUnmarshalArgsHookFunc]{},
	}
}

func runAllHooks[T HookFunc](hs HookCollection[T], callback func(h T)) {
	if len(hs) == 0 {
		return
	}

	hps := slices.Collect(maps.Keys(hs))

	slices.Sort(hps)

	for _, hp := range hps {
		for _, hook := range hs[hp] {
			callback(hook)
		}
	}
}

func (h *Hooks) RunPreStart(f *Frame, pc int) {
	runAllHooks(h.PreStart, func(hook PreStartHookFunc) { hook(f, pc) })
}

func (h *Hooks) RunPreExecute(f *Frame, cmd *fx.CommandNode) {
	runAllHooks(h.PreExecute, func(hook PreExecuteHookFunc) { hook(f, cmd) })
}

func (h *Hooks) RunPostExecute(f *Frame, cmd *fx.CommandNode, jumpPc int, jump bool) {
	runAllHooks(h.PostExecute, func(hook PostExecuteHookFunc) { hook(f, cmd, jumpPc, jump) })
}

func (h *Hooks) RunPostUnmarshalArgs(args any) {
	runAllHooks(h.PostUnmarshalArgs, func(hook PostUnmarshalArgsHookFunc) { hook(args) })
}
