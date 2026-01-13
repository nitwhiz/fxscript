package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

type ErrorHandler func(err error)

type Environment interface {
	HandleError(err error)
	HostCall(f *RuntimeFrame, args []any) (pc int, jump bool)

	Get(variable fx.Variable) (value int)
	Set(variable fx.Variable, value int)
}

type Runtime struct {
	script   *fx.Script
	handlers []CommandHandler
}

func NewRuntime(s *fx.Script) *Runtime {
	r := Runtime{
		script:   s,
		handlers: make([]CommandHandler, 0, fx.UserCommandOffset),
	}

	r.registerBaseCommands()

	return &r
}

func (r *Runtime) NewFrame(pc int, env Environment) *RuntimeFrame {
	return &RuntimeFrame{
		Environment: env,
		Runtime:     r,
		pc:          pc,
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
