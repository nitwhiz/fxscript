//go:build debug

package vm

import "github.com/nitwhiz/fxscript/fx"

func initDap(r *Runtime, identifiers fx.IdentifierTable) {
	initDAPHooks(r)
	go startDAPListener(r, identifiers)
}
