//go:build debug

package vm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"slices"
	"strings"
	"sync"

	"github.com/google/go-dap"
	"github.com/nitwhiz/fxscript/fx"
)

type dapDebugSession struct {
	conn net.Conn
	rw   *bufio.ReadWriter

	breakpoints *BreakpointList

	runtime            *Runtime
	identifiers        fx.IdentifierTable
	commandSourceLines *FileLineList[*fx.CommandNode]
	macroCallLocations *FileLineList[*fx.CommandNode]

	stackFrames  *StackFrames
	currentFrame *Frame

	pauseCond *sync.Cond
	pause     bool

	step bool

	killed bool

	mu *sync.RWMutex
}

func newSession(r *Runtime, identifiers fx.IdentifierTable, conn net.Conn) *dapDebugSession {
	commandSourceLines := newFileLineList[*fx.CommandNode]()
	macroCallLocations := newFileLineList[*fx.CommandNode]()

	for _, cmd := range r.script.Commands() {
		commandSourceLines.Add(cmd.File.Name, cmd.Line, cmd)

		cmdSrc := cmd.SourceInfo

		for cmdSrc.Parent != nil {
			macroCallLocations.Add(cmdSrc.Parent.File.Name, cmdSrc.Parent.Line, cmd)
			cmdSrc = cmdSrc.Parent
		}
	}

	return &dapDebugSession{
		conn:               conn,
		rw:                 bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		breakpoints:        newBreakpointList(),
		runtime:            r,
		identifiers:        identifiers,
		commandSourceLines: commandSourceLines,
		macroCallLocations: macroCallLocations,

		stackFrames: newStackFrames(),

		pauseCond: sync.NewCond(&sync.Mutex{}),
		pause:     true,

		mu: &sync.RWMutex{},
	}
}

func (s *dapDebugSession) updateBreakpoints(file string, breakpoints []dap.SourceBreakpoint) []dap.Breakpoint {
	s.mu.Lock()
	defer s.mu.Unlock()

	bps := make([]dap.Breakpoint, 0)

	s.breakpoints.Delete(file)

	for _, sbp := range breakpoints {
		verified := s.commandSourceLines.Has(file, sbp.Line) || s.macroCallLocations.Has(file, sbp.Line)

		bps = append(bps, dap.Breakpoint{Verified: verified, Line: sbp.Line})

		if verified {
			s.breakpoints.Add(file, sbp.Line)
		}
	}

	slices.SortStableFunc(bps, func(a, b dap.Breakpoint) int { return a.Line - b.Line })

	slog.Info("breakpoints set", slog.Any("breakpoints", s.breakpoints))

	return bps
}

func (s *dapDebugSession) doStep() {
	s.setPause(true)
	s.step = false

	s.waitForPause()
}

func (s *dapDebugSession) checkBreakpointHit() (ok bool) {
	s.mu.RLock()

	sf := s.stackFrames.Current()

	if sf != nil {
		if s.breakpoints.Has(sf.Source.Path, sf.Line) {
			ok = true
		}
	}

	s.mu.RUnlock()

	if ok {
		s.sendStopped("breakpoint", 1)
		s.setPause(true)

		s.waitForPause()
	}

	return
}

func (s *dapDebugSession) kill() {
	if s.killed {
		return
	}

	s.setKilled(true)

	_ = s.conn.Close()
}

func (s *dapDebugSession) labelName(addr int) (result string) {
	for ln, labelAddress := range s.runtime.script.Labels() {
		if labelAddress == addr {
			result = ln
			break
		}
	}

	return
}

func (s *dapDebugSession) newStackFrame(cmd *fx.CommandNode, name string) dap.StackFrame {
	if name == "" {
		name = "<unknown>"
	}

	return dap.StackFrame{
		Name:   name,
		Line:   cmd.Line,
		Column: cmd.Column,
		Source: &dap.Source{
			Path: cmd.File.Name,
		},
	}
}

func (s *dapDebugSession) handleInitialize(r *dap.InitializeRequest) {
	s.send(&dap.InitializeResponse{
		Response: s.response(r.Request),
		Body: dap.Capabilities{
			SupportsConfigurationDoneRequest: true,
		},
	})

	s.send(&dap.InitializedEvent{
		Event: dap.Event{Event: "initialized"},
	})
}

func (s *dapDebugSession) handleDisconnect(r *dap.DisconnectRequest) {
	s.send(&dap.DisconnectResponse{Response: s.response(r.Request)})

	s.kill()
}

func (s *dapDebugSession) handleSetBreakpoints(r *dap.SetBreakpointsRequest) {
	bps := s.updateBreakpoints(r.Arguments.Source.Path, r.Arguments.Breakpoints)

	s.send(&dap.SetBreakpointsResponse{
		Response: s.response(r.Request),
		Body:     dap.SetBreakpointsResponseBody{Breakpoints: bps},
	})
}

func (s *dapDebugSession) handleConfigurationDone(r *dap.ConfigurationDoneRequest) {
	s.send(&dap.ConfigurationDoneResponse{Response: s.response(r.Request)})

	s.setPause(false)
}

func (s *dapDebugSession) handleAttach(r *dap.AttachRequest) {
	s.send(&dap.AttachResponse{Response: s.response(r.Request)})
}

func (s *dapDebugSession) handlePause(r *dap.PauseRequest) {
	s.send(&dap.PauseResponse{Response: s.response(r.Request)})

	s.setPause(true)

	s.send(&dap.StoppedEvent{
		Event: dap.Event{Event: "stopped"},
		Body:  dap.StoppedEventBody{Reason: "pause", ThreadId: 1},
	})
}

func (s *dapDebugSession) handleContinue(r *dap.ContinueRequest) {
	s.setPause(false)

	s.send(&dap.ContinueResponse{Response: s.response(r.Request)})
}

func (s *dapDebugSession) stepNext() {
	s.step = true
	s.setPause(false)
}

func (s *dapDebugSession) handleNext(r *dap.NextRequest) {
	s.stepNext()

	s.send(&dap.NextResponse{Response: s.response(r.Request)})
	s.send(&dap.StoppedEvent{
		Event: dap.Event{Event: "stopped"},
		Body:  dap.StoppedEventBody{Reason: "step", ThreadId: 1},
	})
}

func (s *dapDebugSession) handleStackTrace(r *dap.StackTraceRequest) {
	frames := slices.Collect(s.stackFrames.All())

	s.send(&dap.StackTraceResponse{
		Response: s.response(r.Request),
		Body: dap.StackTraceResponseBody{
			StackFrames: frames,
			TotalFrames: len(frames),
		},
	})
}

func (s *dapDebugSession) handleScopes(r *dap.ScopesRequest) {
	s.send(&dap.ScopesResponse{
		Response: s.response(r.Request),
		Body: dap.ScopesResponseBody{
			Scopes: []dap.Scope{
				{Name: "Identifiers", VariablesReference: 1},
				{Name: "Defines", VariablesReference: 2},
				{Name: "Variables", VariablesReference: 3},
				{Name: "Labels", VariablesReference: 4},
			},
		},
	})
}

func (s *dapDebugSession) handleVariables(r *dap.VariablesRequest) {
	var variables []dap.Variable

	switch r.Arguments.VariablesReference {
	case 1:
		if s.currentFrame == nil {
			break
		}

		for name, addr := range s.identifiers {
			variables = append(
				variables,
				dap.Variable{
					Name:            name,
					Value:           fmt.Sprintf("%d", s.currentFrame.getValue(addr)),
					MemoryReference: fmt.Sprintf("%d", addr),
				},
			)
		}
	case 2:
		if s.currentFrame == nil {
			break
		}

		for name, exprNode := range s.runtime.script.Defines() {
			v, err := s.currentFrame.Eval(exprNode)

			if err == nil {
				variables = append(
					variables,
					dap.Variable{
						Name:  name,
						Value: fmt.Sprintf("%d", v),
					},
				)
			}
		}
	case 3:
		if s.currentFrame == nil {
			break
		}

		for name, addr := range s.runtime.script.Variables() {
			variables = append(
				variables,
				dap.Variable{
					Name:            name,
					Value:           fmt.Sprintf("%d", s.currentFrame.getValue(fx.Identifier(addr))),
					MemoryReference: fmt.Sprintf("%d", addr),
				},
			)
		}
	case 4:
		for name, addr := range s.runtime.script.Labels() {
			variables = append(
				variables,
				dap.Variable{
					Name:            name,
					Value:           fmt.Sprintf("%d", addr),
					MemoryReference: fmt.Sprintf("%d", addr),
				},
			)
		}
	}

	slices.SortStableFunc(variables, func(a, b dap.Variable) int { return strings.Compare(a.Name, b.Name) })

	s.send(&dap.VariablesResponse{
		Response: s.response(r.Request),
		Body: dap.VariablesResponseBody{
			Variables: variables,
		},
	})
}

func (s *dapDebugSession) run() {
	for {
		msg, err := dap.ReadProtocolMessage(s.rw.Reader)

		if err != nil {
			slog.Info("dap read error", slog.Any("err", err))

			if err == io.EOF {
				s.kill()
				return
			}

			if errors.Is(err, net.ErrClosed) {
				s.kill()
				return
			}

			continue
		}

		slog.Info("dap read message", slog.Any("msg", msg))

		go s.handle(msg)
	}
}

func (s *dapDebugSession) response(req dap.Request) dap.Response {
	return dap.Response{
		ProtocolMessage: dap.ProtocolMessage{Type: "response"},
		RequestSeq:      req.Seq,
		Success:         true,
		Command:         req.Command,
	}
}

func (s *dapDebugSession) send(msg dap.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dap.WriteProtocolMessage(s.rw.Writer, msg); err != nil {
		slog.Error("unable to write message", slog.Any("err", err))
	}

	_ = s.rw.Writer.Flush()

	slog.Info("dap write message", slog.Any("msg", msg))
}

func (s *dapDebugSession) sendStopped(reason string, threadID int) {
	s.send(&dap.StoppedEvent{
		Event: dap.Event{Event: "stopped"},
		Body:  dap.StoppedEventBody{Reason: reason, ThreadId: threadID},
	})
}

func (s *dapDebugSession) handle(msg dap.Message) {
	switch r := msg.(type) {
	case *dap.InitializeRequest:
		s.handleInitialize(r)
	case *dap.DisconnectRequest:
		s.handleDisconnect(r)
	case *dap.SetBreakpointsRequest:
		s.handleSetBreakpoints(r)
	case *dap.ConfigurationDoneRequest:
		s.handleConfigurationDone(r)
	case *dap.AttachRequest:
		s.handleAttach(r)
	case *dap.PauseRequest:
		s.handlePause(r)
	case *dap.ContinueRequest:
		s.handleContinue(r)
	case *dap.NextRequest:
		s.handleNext(r)
	case *dap.StackTraceRequest:
		s.handleStackTrace(r)
	case *dap.ScopesRequest:
		s.handleScopes(r)
	case *dap.VariablesRequest:
		s.handleVariables(r)
	}
}

func (s *dapDebugSession) setPause(pause bool) {
	s.pauseCond.L.Lock()
	s.pause = pause
	s.pauseCond.L.Unlock()

	s.pauseCond.Signal()
}

func (s *dapDebugSession) SetStep(step bool) {
	s.step = step
	s.setPause(false)
}

func (s *dapDebugSession) setKilled(killed bool) {
	s.pauseCond.L.Lock()

	s.killed = killed
	s.pause = false

	s.pauseCond.L.Unlock()

	s.pauseCond.Signal()
}

func (s *dapDebugSession) waitForPause() {
	s.pauseCond.L.Lock()
	defer s.pauseCond.L.Unlock()

	for s.pause && !s.killed {
		s.pauseCond.Wait()
	}
}
