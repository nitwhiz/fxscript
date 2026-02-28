//go:build debug

package vm

import (
	"log/slog"
	"sync"
)

type dapState struct {
	session     *dapDebugSession
	sessionCond *sync.Cond
}

var ds = dapState{
	session:     nil,
	sessionCond: sync.NewCond(&sync.Mutex{}),
}

func withSession(callback func(s *dapDebugSession)) {
	ds.sessionCond.L.Lock()
	defer ds.sessionCond.L.Unlock()

	for ds.session == nil {
		ds.sessionCond.Wait()
	}

	if ds.session.killed {
		slog.Warn("use of killed dap session")
		return
	}

	ds.session.waitForPause()

	callback(ds.session)
}

func (d *dapState) setSession(s *dapDebugSession) {
	d.sessionCond.L.Lock()
	defer d.sessionCond.L.Unlock()

	if s != nil && d.session != nil && !d.session.killed {
		// todo: panic is a bit dramatic
		panic("only one dap session can be active at a time")
	}

	d.session = s

	d.sessionCond.Signal()
}
