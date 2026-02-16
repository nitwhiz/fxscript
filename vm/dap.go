//go:build debug

package vm

import (
	"log/slog"
	"net"
	"os"

	"github.com/nitwhiz/fxscript/fx"
)

func initDAPHooks(r *Runtime) {
	if r.hooks == nil {
		r.hooks = NewHooks()
	}

	r.hooks.PreExecute.Add(-100, func(f *Frame, cmd *fx.CommandNode) {
		withSession(func(session *dapDebugSession) {
			cmdSrc := cmd.SourceInfo

			session.stackFrames.AddParents(cmdSrc)
			session.stackFrames.Update(cmdSrc)

			if session.step {
				session.doStep()
			} else {
				session.checkBreakpointHit()
			}

			session.stackFrames.RemoveParents(cmdSrc)
		})
	})

	r.hooks.PostExecute.Add(-100, func(f *Frame, cmd *fx.CommandNode, jumpPc int, jump bool) {
		withSession(func(session *dapDebugSession) {
			if jump == false {
				return
			}

			switch cmd.Type {
			case fx.CmdCall:
				session.stackFrames.Add(session.newStackFrame(cmd, "CALL "+session.labelName(jumpPc)))
			case fx.CmdGoto:
				session.stackFrames.Add(session.newStackFrame(cmd, "GOTO "+session.labelName(jumpPc)))
			case fx.CmdRet:
				session.stackFrames.Ret()
			default:
				// remove parent depth-1?
				break
			}
		})
	})

	r.hooks.PreStart.Add(-100, func(f *Frame, pc int) {
		withSession(func(session *dapDebugSession) {
			session.stackFrames.Set(session.newStackFrame(f.script.Commands()[pc], "CALL "+session.labelName(pc)))
			session.currentFrame = f
		})
	})
}

func startDAPListener(r *Runtime, identifiers fx.IdentifierTable) {
	listener, err := net.Listen("tcp", ":4711")

	if err != nil {
		panic(err)
	}

	defer listener.Close()

	slog.Info("waiting for dap connection")

	conn, err := listener.Accept()

	if err != nil {
		slog.Error("unable to accept connection", slog.Any("err", err))
		return
	}

	slog.Info("starting dap session")

	session := newSession(r, identifiers, conn)

	ds.setSession(session)

	session.run()

	ds.setSession(nil)

	slog.Info("dap session ended, exiting")

	os.Exit(0)
}
