package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

type ErrorHandler func(err error)

type Environment interface {
	HandleError(err error)

	Get(identifier fx.Identifier) (value int)
	Set(identifier fx.Identifier, value int)
}

type Runtime struct {
	script    *fx.Script
	handlers  []CommandHandler
	stackSize int
}

func NewRuntime(s *fx.Script, cfg *RuntimeConfig) *Runtime {
	stackSize := cfg.StackSize

	if stackSize == 0 {
		stackSize = 16
	}

	r := Runtime{
		script:    s,
		handlers:  make([]CommandHandler, 0, fx.UserCommandOffset),
		stackSize: stackSize,
	}

	r.RegisterCommands(BaseCommands)
	r.RegisterCommands(cfg.UserCommands)

	return &r
}

func (r *Runtime) NewFrame(pc int, env Environment) *RuntimeFrame {
	return &RuntimeFrame{
		Environment: env,
		Runtime:     r,
		pc:          pc,
		stack:       make([]int, r.stackSize),
	}
}

// Start starts a new frame to run from a specific PC
func (r *Runtime) Start(pc int, env Environment) {
	f := r.NewFrame(pc, env)

	commands := f.script.Commands()

	for ; f.pc < len(commands); f.pc++ {
		jumpTarget, jump, _ := f.ExecuteCommand(commands[f.pc])

		if jump {
			f.pc = jumpTarget - 1
		}
	}
}

func (r *Runtime) Label(name string) (pc int, ok bool) {
	return r.script.Label(name)
}

func (r *Runtime) Call(label string, env Environment) bool {
	if pc, ok := r.Label(label); ok {
		r.Start(pc, env)
		return true
	}

	return false
}
