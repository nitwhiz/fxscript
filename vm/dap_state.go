//go:build debug

package vm

import (
	"log/slog"
	"sync"
)

type dapState struct {
	session *dapDebugSession

	sessionCond *sync.Cond
	mu          *sync.RWMutex
}

var ds = dapState{
	session:     nil,
	sessionCond: sync.NewCond(&sync.Mutex{}),
	mu:          &sync.RWMutex{},
}

func withSession(callback func(s *dapDebugSession)) {
	ds.sessionCond.L.Lock()
	defer ds.sessionCond.L.Unlock()

	for ds.session == nil {
		ds.sessionCond.Wait()
	}

	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if ds.session == nil {
		return
	}

	if ds.session.killed {
		slog.Warn("use of killed dap session")
		return
	}

	ds.session.waitForPause()

	defer callback(ds.session)
}

func (d *dapState) setSession(s *dapDebugSession) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.sessionCond.L.Lock()

	if s != nil && d.session != nil && !d.session.killed {
		// todo: panic is a bit dramatic
		panic("only one dap session can be active at a time")
	}

	d.session = s

	d.sessionCond.L.Unlock()

	d.sessionCond.Signal()
}
