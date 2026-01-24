package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

func handleNop(*RuntimeFrame, []fx.ExpressionNode) (jumpTarget int, jump bool) {
	return
}

func handlePush(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable fx.Identifier `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		f.pushStack(f.getValue(args.Variable))
		return
	})
}

func handlePop(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable fx.Identifier `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		if v, ok := f.popStack(); ok {
			f.setValue(args.Variable, v)
		}

		return
	})
}

func handleGoto(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		JumpTarget int `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		return args.JumpTarget, true
	})
}

func handleSet(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable fx.Identifier `arg:""`
		Value    int           `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		if args.Variable >= fx.VariableOffset {
			f.setMemory(args.Variable, args.Value)
			return
		}

		f.Set(args.Variable, args.Value)
		return
	})
}

func handleCall(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Addr int `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		if args.Addr == 0 {
			return f.script.EndOfScript(), true
		}

		f.pushStack(f.pc + 1)

		return args.Addr, true
	})
}

func handleRet(f *RuntimeFrame, _ []fx.ExpressionNode) (jumpTarget int, jump bool) {
	var ok bool

	jump = true
	jumpTarget, ok = f.popStack()

	if !ok {
		jumpTarget = f.script.EndOfScript()
	}

	return
}

func handleJumpIf(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Condition  int `arg:""`
		JumpTarget int `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		if args.Condition != 0 {
			jumpTarget = args.JumpTarget
			jump = true
		}

		return
	})
}
