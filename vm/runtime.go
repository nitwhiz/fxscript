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
	hooks            *Hooks
	script           *fx.Script
	handlers         []CommandHandler
	callStackSize    int
	operandStackSize int
	memory           []int
}

func NewRuntime(s *fx.Script, cfg *RuntimeConfig) *Runtime {
	callStackSize := cfg.CallStackSize

	if callStackSize == 0 {
		callStackSize = 32
	}

	operandStackSize := cfg.OperandStackSize

	if operandStackSize == 0 {
		operandStackSize = 64
	}

	r := Runtime{
		hooks:            cfg.Hooks,
		script:           s,
		handlers:         make([]CommandHandler, 0, fx.UserCommandOffset),
		callStackSize:    callStackSize,
		operandStackSize: operandStackSize,
		memory:           make([]int, len(s.Variables())),
	}

	r.RegisterCommands(BaseCommands)
	r.RegisterCommands(cfg.UserCommands)

	return &r
}

func (r *Runtime) setMemory(variable fx.Identifier, value int) {
	addr := int(variable - fx.VariableOffset)

	if addr > len(r.memory) {
		return
	}

	r.memory[addr] = value
}

func (r *Runtime) getMemory(variable fx.Identifier) (value int) {
	addr := int(variable - fx.VariableOffset)

	if addr >= len(r.memory) {
		value = 0
		return
	}

	value = r.memory[addr]
	return
}

func (r *Runtime) NewFrame(pc int, env Environment) *Frame {
	return &Frame{
		Environment:  env,
		Runtime:      r,
		pc:           pc,
		callStack:    make([]int, r.callStackSize),
		operandStack: make([]int, r.operandStackSize),
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
