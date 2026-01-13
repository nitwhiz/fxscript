package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

func handleNop(f *RuntimeFrame, _ []fx.ExpressionNode) (jumpTarget int, jump bool) {
	f.CmdNop()
	return
}

func handleHostCall(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	args := make([]any, len(cmdArgs))

	for i, arg := range cmdArgs {
		args[i] = f.eval(arg)
	}

	return f.HostCall(f, args)
}

func handleGoto(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		JumpTarget int `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		return f.CmdGoto(args.JumpTarget)
	})
}

func handleSet(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable fx.Variable `arg:""`
		Value    int         `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		f.CmdSet(args.Variable, args.Value)
		return
	})
}

func handleCopy(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		From fx.Variable `arg:""`
		To   fx.Variable `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		f.CmdCopy(args.From, args.To)
		return
	})
}

func handleAdd(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable fx.Variable `arg:""`
		Value    int         `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		f.CmdAdd(args.Variable, args.Value)
		return
	})
}

func handleCall(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Addr int `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		return f.CmdCall(args.Addr)
	})
}

func handleRet(f *RuntimeFrame, _ []fx.ExpressionNode) (jumpTarget int, jump bool) {
	return f.CmdRet()
}

func handleJumpIf(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable   fx.Variable `arg:""`
		Value      int         `arg:""`
		JumpTarget int         `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		return f.CmdJumpIf(args.Variable, args.Value, args.JumpTarget)
	})
}

func handleSetFlag(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable fx.Variable `arg:""`
		Flag     fx.Flag     `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		f.CmdSetFlag(args.Variable, args.Flag)
		return
	})
}

func handleClearFlag(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable fx.Variable `arg:""`
		Flag     fx.Flag     `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		f.CmdClearFlag(args.Variable, args.Flag)
		return
	})
}

func handleJumpIfFlag(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable   fx.Variable `arg:""`
		Flag       fx.Flag     `arg:""`
		JumpTarget int         `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		return f.CmdJumpIfFlag(args.Variable, args.Flag, args.JumpTarget)
	})
}

func handleJumpIfNotFlag(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
	type Args struct {
		Variable   fx.Variable `arg:""`
		Flag       fx.Flag     `arg:""`
		JumpTarget int         `arg:""`
	}

	return WithArgs(f, cmdArgs, func(f *RuntimeFrame, args *Args) (jumpTarget int, jump bool) {
		return f.CmdJumpIfNotFlag(args.Variable, args.Flag, args.JumpTarget)
	})
}
