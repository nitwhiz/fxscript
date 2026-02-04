package fx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	identA = iota
)

const (
	cmdMyCmd = UserCommandOffset + iota
)

func parse(script string) (commands []*CommandNode, defines map[string]ExpressionNode, labels map[string]int, macros map[string]*Macro, err error) {
	l := NewLexer([]byte(script))
	p := NewParser(l, &ParserConfig{
		CommandTypes: CommandTypeTable{
			"myCmd": cmdMyCmd,
		},
		Identifiers: IdentifierTable{
			"A": identA,
		},
	})

	s, err := p.Parse()

	if err != nil {
		return
	}

	return s.Commands(), s.defines, s.labels, s.macros, nil
}

func TestParser_Ident(t *testing.T) {
	script := "myCmd A\n"

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{
			cmdMyCmd,
			[]ExpressionNode{
				&IdentifierNode{Identifier: identA},
			},
		},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_Numbers(t *testing.T) {
	script := "myCmd 42, -42, 13.37, -13.37, +13\n"

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{
			&IntegerNode{Value: 42},
			&UnaryOpNode{Operator: &Token{SUB, "-"}, Expr: &IntegerNode{Value: 42}},
			&FloatNode{Value: 13.37},
			&UnaryOpNode{Operator: &Token{SUB, "-"}, Expr: &FloatNode{Value: 13.37}},
			&UnaryOpNode{Operator: &Token{ADD, "+"}, Expr: &IntegerNode{Value: 13}},
		}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_OperatorsAndNumbers(t *testing.T) {
	script := "myCmd +42 + +13 - 37 * -72 / 42\n"

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{
			&BinaryOpNode{
				Left: &BinaryOpNode{
					Left: &UnaryOpNode{
						Operator: &Token{ADD, "+"},
						Expr:     &IntegerNode{42},
					},
					Operator: &Token{ADD, "+"},
					Right: &UnaryOpNode{
						Operator: &Token{ADD, "+"},
						Expr:     &IntegerNode{13},
					},
				},
				Operator: &Token{SUB, "-"},
				Right: &BinaryOpNode{
					Left: &BinaryOpNode{
						Left:     &IntegerNode{37},
						Operator: &Token{MUL, "*"},
						Right: &UnaryOpNode{
							Operator: &Token{SUB, "-"},
							Expr:     &IntegerNode{72},
						},
					},
					Operator: &Token{DIV, "/"},
					Right:    &IntegerNode{42},
				},
			},
		}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_OperatorsWithParens(t *testing.T) {
	script := "myCmd (+42 + 13) - (37 * (-72 / 42 ))\n"

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{
			&BinaryOpNode{
				Left: &BinaryOpNode{
					Left: &UnaryOpNode{
						Operator: &Token{ADD, "+"},
						Expr:     &IntegerNode{42},
					},
					Operator: &Token{ADD, "+"},
					Right:    &IntegerNode{13},
				},
				Operator: &Token{SUB, "-"},
				Right: &BinaryOpNode{
					Left:     &IntegerNode{37},
					Operator: &Token{MUL, "*"},
					Right: &BinaryOpNode{
						Left: &UnaryOpNode{
							Operator: &Token{SUB, "-"},
							Expr:     &IntegerNode{72},
						},
						Operator: &Token{DIV, "/"},
						Right:    &IntegerNode{42},
					},
				},
			},
		}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_InvOperator(t *testing.T) {
	script := "myCmd ^42, ^-13\n"

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{
			&UnaryOpNode{Operator: &Token{INV, "^"}, Expr: &IntegerNode{42}},
			&UnaryOpNode{Operator: &Token{INV, "^"}, Expr: &UnaryOpNode{Operator: &Token{SUB, "-"}, Expr: &IntegerNode{13}}},
		}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_AddrOfOperator(t *testing.T) {
	script := "myCmd &42, &-13\n"

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{
			&UnaryOpNode{Operator: &Token{AND, "&"}, Expr: &IntegerNode{42}},
			&UnaryOpNode{Operator: &Token{AND, "&"}, Expr: &UnaryOpNode{Operator: &Token{SUB, "-"}, Expr: &IntegerNode{13}}},
		}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_AndOperator(t *testing.T) {
	script := "myCmd 4 & 16\n"

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{&BinaryOpNode{Left: &IntegerNode{4}, Operator: &Token{AND, "&"}, Right: &IntegerNode{16}}}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_OrOperator(t *testing.T) {
	script := "myCmd 4 | 16\n"

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{&BinaryOpNode{Left: &IntegerNode{4}, Operator: &Token{OR, "|"}, Right: &IntegerNode{16}}}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_Labels(t *testing.T) {
	script := `
		myCmd 1
		loop:
			myCmd 2
		%_inner:
			myCmd 3
		end:
			myCmd 4
	`

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{&IntegerNode{1}}},
		{cmdMyCmd, []ExpressionNode{&IntegerNode{2}}},
		{cmdMyCmd, []ExpressionNode{&IntegerNode{3}}},
		{cmdMyCmd, []ExpressionNode{&IntegerNode{4}}},
	}

	require.Equal(t, expectedCommands, commands)

	expectedLabels := map[string]int{"loop": 1, "loop_inner": 2, "end": 3}

	require.Equal(t, expectedLabels, labels)

	require.Empty(t, defines)
	require.Empty(t, macros)
}

func TestParser_Macros(t *testing.T) {
	script := `
		macro m1
			myCmd 1
		endmacro

		macro m2 $value
			m1
			myCmd $value
		endmacro

		m1
		m2 2
		myCmd 3
	`

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{&IntegerNode{1}}},
		{cmdMyCmd, []ExpressionNode{&IntegerNode{1}}},
		{cmdMyCmd, []ExpressionNode{&IntegerNode{2}}},
		{cmdMyCmd, []ExpressionNode{&IntegerNode{3}}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Len(t, macros, 2)

	require.Empty(t, defines)
	require.Empty(t, labels)
}

func TestParser_MacrosWithLocalLabels(t *testing.T) {
	script := `
		macro mLoop
		%start:
			myCmd A
			myCmd %start
		endmacro

		mLoop
		mLoop
	`

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{&IdentifierNode{identA}}},
		{cmdMyCmd, []ExpressionNode{&AddressNode{0}}},
		{cmdMyCmd, []ExpressionNode{&IdentifierNode{identA}}},
		{cmdMyCmd, []ExpressionNode{&AddressNode{2}}},
	}

	require.Equal(t, expectedCommands, commands)
	require.Len(t, defines, 0)
	require.Len(t, labels, 2)
	require.Len(t, macros, 1)
}

func TestParser_Defines(t *testing.T) {
	script := `
		def msgHello "Hello World!"
		def wordCount 2 << 1

		myCmd msgHello, wordCount
	`

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{&StringNode{"Hello World!"}, &BinaryOpNode{&IntegerNode{2}, &Token{SHL, "<<"}, &IntegerNode{1}}}},
	}

	require.Equal(t, expectedCommands, commands)

	expecteddefines := map[string]ExpressionNode{
		"msgHello": &StringNode{"Hello World!"},
		"wordCount": &BinaryOpNode{
			Left:     &IntegerNode{2},
			Operator: &Token{SHL, "<<"},
			Right:    &IntegerNode{1},
		},
	}

	require.Equal(t, expecteddefines, defines)

	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_Variables(t *testing.T) {
	script := `
		var myVar
		myCmd myVar 42
	`

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{&IdentifierNode{VariableOffset}, &IntegerNode{42}}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_Strings(t *testing.T) {
	script := `
		myCmd "Hello World!"
		myCmd "Strings can .contain all @sorts of -42.1337 # characters"
	`

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{cmdMyCmd, []ExpressionNode{&StringNode{"Hello World!"}}},
		{cmdMyCmd, []ExpressionNode{&StringNode{"Strings can .contain all @sorts of -42.1337 # characters"}}},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}
