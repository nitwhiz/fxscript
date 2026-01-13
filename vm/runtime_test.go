package vm

import (
	"testing"

	"github.com/nitwhiz/fxscript/fx"
	"github.com/stretchr/testify/require"
)

var _ Environment = (*TestEnv)(nil)

type testCommand struct {
	Typ     fx.CommandType
	Name    string
	Handler CommandHandler
}

const (
	CmdTest = fx.UserCommandOffset + iota
)

const (
	varA = iota
	varB
	varC
)

const (
	identAlpha = iota
)

type TestEnv struct {
	values map[fx.Variable]int

	testValue        bool
	lastHostCallArgs []any
	lastError        error
}

func (e *TestEnv) HostCall(_ *RuntimeFrame, args []any) (pc int, jump bool) {
	e.lastHostCallArgs = args
	return
}

func NewTestEnv(s string) (*Runtime, *TestEnv, error) {
	e := &TestEnv{
		values: make(map[fx.Variable]int),
	}

	r, err := e.Load(s)

	if err != nil {
		return nil, nil, err
	}

	return r, e, nil
}

func (e *TestEnv) Load(s string) (*Runtime, error) {
	commands := []*testCommand{
		{CmdTest, "test", func(f *RuntimeFrame, cmdArgs []fx.ExpressionNode) (jumpTarget int, jump bool) {
			e.testValue = true
			return
		}},
	}

	commandTypes := fx.CommandTypeTable{}

	for _, cmd := range commands {
		commandTypes[cmd.Name] = cmd.Typ
	}

	fxs, err := fx.LoadScript([]byte(s), &fx.ParserConfig{
		CommandTypes: commandTypes,
		Variables: fx.VariableTable{
			"a": varA,
			"b": varB,
			"c": varC,
		},
		Identifiers: fx.IdentifierTable{
			"alpha": identAlpha,
		},
	})

	if err != nil {
		return nil, err
	}

	r := NewRuntime(fxs)

	for _, cmd := range commands {
		r.registerCommand(&Command{cmd.Typ, cmd.Handler})
	}

	return r, nil
}

func (e *TestEnv) HandleError(err error) {
	e.lastError = err
}

func (e *TestEnv) Get(variable fx.Variable) (value int) {
	if val, ok := e.values[variable]; ok {
		return val
	}

	return 0
}

func (e *TestEnv) Set(variable fx.Variable, value int) {
	e.values[variable] = value
}

func runScript(t *testing.T, script string) *TestEnv {
	r, env, err := NewTestEnv(script)

	require.NoError(t, err)

	r.Start(0, env)

	return env
}

func TestCommandWithArgsRequiringComma(t *testing.T) {
	script := "hostCall \"hello\", -42.0, \"world\"\n"

	env := runScript(t, script)

	require.Len(t, env.lastHostCallArgs, 3)
}

func TestHostCallWithVariousArgs(t *testing.T) {
	script := `
		const name "marvin"
		set a 15
		doCall:
			hostCall (doCall - 1) "hello" ((alpha + 4) * a / 42.0 * 2) name
	`

	env := runScript(t, script)

	require.Len(t, env.lastHostCallArgs, 4)
}

func TestHandleError(t *testing.T) {
	script := "set\n"

	env := runScript(t, script)

	require.NotNil(t, env.lastError)
}

func TestCustomCommand(t *testing.T) {
	script := "test\n"

	env := runScript(t, script)

	require.True(t, env.testValue)
}

func TestAddInts(t *testing.T) {
	script := "set a (1 + 2)\n"

	env := runScript(t, script)

	require.Equal(t, 3, env.values[varA])
}

func TestAddFloats(t *testing.T) {
	script := "set a (1.2 + 2.7 + 0.3)\n"

	env := runScript(t, script)

	require.Equal(t, 4, env.values[varA])
}

func TestAddIntAndFloat(t *testing.T) {
	script := "set a (1 + 2.7 + 0.4)\n"

	env := runScript(t, script)

	require.Equal(t, 4, env.values[varA])
}

func TestSetAddWithVariableArgument(t *testing.T) {
	script := `
		set a (1 + 2)
		set (a + 1) (3 * 3)
	`

	r, env, err := NewTestEnv(script)

	require.NoError(t, err)

	r.Start(0, env)

	require.Equal(t, 3, env.values[varA])

	require.Equal(t, 9, env.values[varA+4])
}

func TestCallLabel(t *testing.T) {
	script := `
		goto end
		myLabel:
			set a 42
		end:
			nop
	`

	r, env, err := NewTestEnv(script)

	require.NoError(t, err)

	r.Start(0, env)

	require.Equal(t, 0, env.values[varA])

	ok := r.Call("myLabel", env)

	require.True(t, ok)
	require.Equal(t, 42, env.values[varA])
}
