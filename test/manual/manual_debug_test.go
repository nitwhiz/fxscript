//go:build debug

package test

import (
	"os"
	"path"
	"testing"

	"github.com/nitwhiz/fxscript/fx"
	"github.com/nitwhiz/fxscript/vm"
	"github.com/stretchr/testify/require"
)

type TestEnv struct {
	t testing.TB

	memory  map[fx.Identifier]int
	results []any
}

func NewTestEnv(t testing.TB) *TestEnv {
	return &TestEnv{
		t:      t,
		memory: make(map[fx.Identifier]int),
	}
}

func (env *TestEnv) Get(identifier fx.Identifier) (value int) {
	return env.memory[identifier]
}

func (env *TestEnv) Set(identifier fx.Identifier, value int) {
	env.memory[identifier] = value
}

func (env *TestEnv) HandleError(err error) {
	env.t.Fatal(err)
}

func TestManual(t *testing.T) {
	t.Skip("manual test")

	rtCfg := &vm.RuntimeConfig{
		Debug: true,
	}

	wd, err := os.Getwd()

	if err != nil {
		t.Fatal(err)
	}

	parserConfig := rtCfg.ParserConfig(
		fx.NewParserOsFS(path.Join(wd, "scripts/")),
		nil,
	)

	fxs, err := fx.LoadFile("main.fx", parserConfig)

	require.NoError(t, err)

	e := NewTestEnv(t)

	rt := vm.NewRuntime(fxs, rtCfg)

	mainPc, ok := fxs.Label("_main")

	require.True(t, ok)

	rt.Start(mainPc, e)
}
