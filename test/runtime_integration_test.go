package test

import (
	"bytes"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/nitwhiz/fxscript/fx"
	"github.com/nitwhiz/fxscript/vm"
	"github.com/stretchr/testify/require"
)

const (
	testFileExt = "fxt"
)

const (
	identA = iota
)

const (
	cmdEval = fx.UserCommandOffset + iota
	cmdBreakpoint
)

var commandNames = map[fx.CommandType]string{}

func init() {
	for _, cmd := range vm.BaseCommands {
		commandNames[cmd.Type] = cmd.Name
	}

	commandNames[cmdEval] = "eval"
	commandNames[cmdBreakpoint] = "break"
}

type TestEnv struct {
	t testing.TB

	values  map[fx.Identifier]int
	results []any
}

func NewTestEnv(t testing.TB) *TestEnv {
	return &TestEnv{
		t: t,

		values:  make(map[fx.Identifier]int),
		results: make([]any, 0),
	}
}

func (env *TestEnv) Get(identifier fx.Identifier) (value int) {
	return env.values[identifier]
}

func (env *TestEnv) Set(identifier fx.Identifier, value int) {
	env.values[identifier] = value
}

func (env *TestEnv) HandleError(err error) {
	env.t.Fatal(err)
}

func (env *TestEnv) handleEval(f *vm.Frame, args []fx.ExpressionNode) (jumpTarget int, jump bool) {
	values := make([]any, len(args))

	var err error

	for i, arg := range args {
		if values[i], err = f.Eval(arg); err != nil {
			env.HandleError(err)
		}
	}

	env.results = append(env.results, values...)

	return
}

func (env *TestEnv) handleBreak(f *vm.Frame, args []fx.ExpressionNode) (jumpTarget int, jump bool) {
	runtime.Breakpoint()
	return
}

func TestIntegration(t *testing.T) {
	testScripts, err := filepath.Glob("scripts/*." + testFileExt)

	require.NoError(t, err)

	slices.Sort(testScripts)

	identifiers := fx.IdentifierTable{
		"A": identA,
	}

	for _, scriptPath := range testScripts {
		t.Run(strings.TrimSuffix(path.Base(scriptPath), "."+testFileExt), func(t *testing.T) {
			data, err := os.ReadFile(scriptPath)

			require.NoError(t, err)

			segments := bytes.Split(data, []byte("--- EXPECT ---\n"))

			require.Len(t, segments, 2, "does the script have an EXPECT section?")

			e := NewTestEnv(t)

			rtCfg := &vm.RuntimeConfig{
				UserCommands: []*vm.Command{
					{"eval", cmdEval, e.handleEval},
					{"break", cmdBreakpoint, e.handleBreak},
				},
				Identifiers: identifiers,
				Hooks: &vm.Hooks{
					PreExecute: func(cmd *fx.CommandNode) {
						slog.Info("EXEC", slog.String("name", commandNames[cmd.Type]), slog.String("cmd", cmd.String()))
					},
					PostUnmarshalArgs: func(args any) {
						slog.Info("ARGS", slog.Any("args", args))
					},
				},
			}

			parserConfig := rtCfg.ParserConfig(
				fx.NewParserFS(os.DirFS("scripts/")),
				func(v string) ([]byte, error) {
					return []byte(v + " \"hello world!\""), nil
				},
			)

			fxs, err := fx.LoadScript(segments[0], path.Base(scriptPath), parserConfig)

			require.NoError(t, err)

			vm.NewRuntime(fxs, rtCfg).Start(0, e)

			expectLines := bytes.Split(segments[1], []byte("\n"))

			if len(expectLines) <= 1 {
				t.Skip("no EXPECT lines found")
			}

			expectLines = expectLines[:len(expectLines)-1]

			rPtr := 0
			expectLineOffset := len(bytes.Split(segments[0], []byte("\n")))

			for l, expectLine := range expectLines {
				if len(expectLine) == 0 {
					continue
				}

				firstChar := expectLine[0]

				if firstChar == '#' {
					continue
				}

				if len(e.results) <= rPtr {
					t.Fatal("unexpected end of results")
				}

				value := e.results[rPtr]

				currentLineInFile := expectLineOffset + l + 1

				switch firstChar {
				case '"':
					t.Log(value)

					require.IsType(t, "", value, "result expected to be a string at EXPECT line "+strconv.Itoa(currentLineInFile))

					expectedString := string(expectLine[1 : len(expectLine)-1])

					require.EqualValues(t, expectedString, value, "value mismatch at EXPECT line "+strconv.Itoa(currentLineInFile))
				default:
					if bytes.Contains(expectLine, []byte(".")) {
						t.Log(value)

						valueKind := reflect.ValueOf(value).Kind()

						if valueKind != reflect.Float64 {
							t.Fatal("result is expected to be a float64 at EXPECT line " + strconv.Itoa(currentLineInFile))
						}

						expectedFloat, err := strconv.ParseFloat(string(expectLine), 64)

						require.NoError(t, err, "unable to parse float64 at EXPECT line "+strconv.Itoa(currentLineInFile))
						require.EqualValues(t, expectedFloat, value, "value mismatch at EXPECT line "+strconv.Itoa(currentLineInFile))
					} else {
						t.Log(value)

						valueKind := reflect.ValueOf(value).Kind()

						if valueKind != reflect.Int {
							t.Fatal("result is expected to be an int at EXPECT line " + strconv.Itoa(currentLineInFile))
						}

						expectedInt, err := strconv.ParseInt(string(expectLine), 10, 64)

						require.NoError(t, err, "unable to parse int64 at EXPECT line "+strconv.Itoa(currentLineInFile))
						require.EqualValues(t, int(expectedInt), value, "value mismatch at EXPECT line "+strconv.Itoa(currentLineInFile))
					}
				}

				rPtr++
			}

			if rPtr < len(e.results) {
				missingChecks := len(e.results) - rPtr
				t.Fatal("not all results were checked: missing " + strconv.Itoa(missingChecks) + " EXPECT line(s)")
			}
		})
	}
}
