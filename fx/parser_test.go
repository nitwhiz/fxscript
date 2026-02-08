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
	l := NewLexer([]byte(script), "")
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

func sourceInfo(line, col int) *SourceInfo {
	return &SourceInfo{Line: line, Column: col, Filename: ""}
}

func TestParser_Ident(t *testing.T) {
	script := "myCmd A\n"

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{
			SourceInfo: sourceInfo(1, 1),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IdentifierNode{
					SourceInfo: sourceInfo(1, 7),
					Identifier: identA,
				},
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
		{
			SourceInfo: sourceInfo(1, 1),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IntegerNode{
					SourceInfo: sourceInfo(1, 7),
					Value:      42,
				},
				&UnaryOpNode{
					SourceInfo: sourceInfo(1, 11),
					Operator: &Token{
						SourceInfo: sourceInfo(1, 11),
						Type:       SUB,
						Value:      "-",
					},
					Expr: &IntegerNode{
						SourceInfo: sourceInfo(1, 12),
						Value:      42,
					},
				},
				&FloatNode{
					SourceInfo: sourceInfo(1, 16),
					Value:      13.37,
				},
				&UnaryOpNode{
					SourceInfo: sourceInfo(1, 23),
					Operator: &Token{
						SourceInfo: sourceInfo(1, 23),
						Type:       SUB,
						Value:      "-",
					},
					Expr: &FloatNode{
						SourceInfo: sourceInfo(1, 24),
						Value:      13.37,
					},
				},
				&UnaryOpNode{
					SourceInfo: sourceInfo(1, 31),
					Operator: &Token{
						SourceInfo: sourceInfo(1, 31),
						Type:       ADD,
						Value:      "+",
					},
					Expr: &IntegerNode{
						SourceInfo: sourceInfo(1, 32),
						Value:      13,
					},
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(1, 1),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&BinaryOpNode{
					SourceInfo: sourceInfo(1, 17),
					Left: &BinaryOpNode{
						SourceInfo: sourceInfo(1, 11),
						Left: &UnaryOpNode{
							SourceInfo: sourceInfo(1, 7),
							Operator: &Token{
								SourceInfo: sourceInfo(1, 7),
								Type:       ADD,
								Value:      "+",
							},
							Expr: &IntegerNode{
								SourceInfo: sourceInfo(1, 8),
								Value:      42,
							},
						},
						Operator: &Token{
							SourceInfo: sourceInfo(1, 11),
							Type:       ADD,
							Value:      "+",
						},
						Right: &UnaryOpNode{
							SourceInfo: sourceInfo(1, 13),
							Operator: &Token{
								SourceInfo: sourceInfo(1, 13),
								Type:       ADD,
								Value:      "+",
							},
							Expr: &IntegerNode{
								SourceInfo: sourceInfo(1, 14),
								Value:      13,
							},
						},
					},
					Operator: &Token{
						SourceInfo: sourceInfo(1, 17),
						Type:       SUB,
						Value:      "-",
					},
					Right: &BinaryOpNode{
						SourceInfo: sourceInfo(1, 28),
						Left: &BinaryOpNode{
							SourceInfo: sourceInfo(1, 22),
							Left: &IntegerNode{
								SourceInfo: sourceInfo(1, 19),
								Value:      37,
							},
							Operator: &Token{
								SourceInfo: sourceInfo(1, 22),
								Type:       MUL,
								Value:      "*",
							},
							Right: &UnaryOpNode{
								SourceInfo: sourceInfo(1, 24),
								Operator: &Token{
									SourceInfo: sourceInfo(1, 24),
									Type:       SUB,
									Value:      "-",
								},
								Expr: &IntegerNode{
									SourceInfo: sourceInfo(1, 25),
									Value:      72,
								},
							},
						},
						Operator: &Token{
							SourceInfo: sourceInfo(1, 28),
							Type:       DIV,
							Value:      "/",
						},
						Right: &IntegerNode{
							SourceInfo: sourceInfo(1, 30),
							Value:      42,
						},
					},
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(1, 1),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&BinaryOpNode{
					SourceInfo: sourceInfo(1, 18),
					Left: &BinaryOpNode{
						SourceInfo: sourceInfo(1, 12),
						Left: &UnaryOpNode{
							SourceInfo: sourceInfo(1, 8),
							Operator: &Token{
								SourceInfo: sourceInfo(1, 8),
								Type:       ADD,
								Value:      "+",
							},
							Expr: &IntegerNode{
								SourceInfo: sourceInfo(1, 9),
								Value:      42,
							},
						},
						Operator: &Token{
							SourceInfo: sourceInfo(1, 12),
							Type:       ADD,
							Value:      "+",
						},
						Right: &IntegerNode{
							SourceInfo: sourceInfo(1, 14),
							Value:      13,
						},
					},
					Operator: &Token{
						SourceInfo: sourceInfo(1, 18),
						Type:       SUB,
						Value:      "-",
					},
					Right: &BinaryOpNode{
						SourceInfo: sourceInfo(1, 24),
						Left: &IntegerNode{
							SourceInfo: sourceInfo(1, 21),
							Value:      37,
						},
						Operator: &Token{
							SourceInfo: sourceInfo(1, 24),
							Type:       MUL,
							Value:      "*",
						},
						Right: &BinaryOpNode{
							SourceInfo: sourceInfo(1, 31),
							Left: &UnaryOpNode{
								SourceInfo: sourceInfo(1, 27),
								Operator: &Token{
									SourceInfo: sourceInfo(1, 27),
									Type:       SUB,
									Value:      "-",
								},
								Expr: &IntegerNode{
									SourceInfo: sourceInfo(1, 28),
									Value:      72,
								},
							},
							Operator: &Token{
								SourceInfo: sourceInfo(1, 31),
								Type:       DIV,
								Value:      "/",
							},
							Right: &IntegerNode{
								SourceInfo: sourceInfo(1, 33),
								Value:      42,
							},
						},
					},
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(1, 1),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&UnaryOpNode{
					SourceInfo: sourceInfo(1, 7),
					Operator: &Token{
						SourceInfo: sourceInfo(1, 7),
						Type:       INV,
						Value:      "^",
					},
					Expr: &IntegerNode{
						SourceInfo: sourceInfo(1, 8),
						Value:      42,
					},
				},
				&UnaryOpNode{
					SourceInfo: sourceInfo(1, 12),
					Operator: &Token{
						SourceInfo: sourceInfo(1, 12),
						Type:       INV,
						Value:      "^",
					},
					Expr: &UnaryOpNode{
						SourceInfo: sourceInfo(1, 13),
						Operator: &Token{
							SourceInfo: sourceInfo(1, 13),
							Type:       SUB,
							Value:      "-",
						},
						Expr: &IntegerNode{
							SourceInfo: sourceInfo(1, 14),
							Value:      13,
						},
					},
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(1, 1),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&UnaryOpNode{
					SourceInfo: sourceInfo(1, 7),
					Operator: &Token{
						SourceInfo: sourceInfo(1, 7),
						Type:       AND,
						Value:      "&",
					},
					Expr: &IntegerNode{
						SourceInfo: sourceInfo(1, 8),
						Value:      42,
					},
				},
				&UnaryOpNode{
					SourceInfo: sourceInfo(1, 12),
					Operator: &Token{
						SourceInfo: sourceInfo(1, 12),
						Type:       AND,
						Value:      "&",
					},
					Expr: &UnaryOpNode{
						SourceInfo: sourceInfo(1, 13),
						Operator: &Token{
							SourceInfo: sourceInfo(1, 13),
							Type:       SUB,
							Value:      "-",
						},
						Expr: &IntegerNode{
							SourceInfo: sourceInfo(1, 14),
							Value:      13,
						},
					},
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(1, 1),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&BinaryOpNode{
					SourceInfo: sourceInfo(1, 9),
					Left: &IntegerNode{
						SourceInfo: sourceInfo(1, 7),
						Value:      4,
					},
					Operator: &Token{
						SourceInfo: sourceInfo(1, 9),
						Type:       AND,
						Value:      "&",
					},
					Right: &IntegerNode{
						SourceInfo: sourceInfo(1, 11),
						Value:      16,
					},
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(1, 1),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&BinaryOpNode{
					SourceInfo: sourceInfo(1, 9),
					Left: &IntegerNode{
						SourceInfo: sourceInfo(1, 7),
						Value:      4,
					},
					Operator: &Token{
						SourceInfo: sourceInfo(1, 9),
						Type:       OR,
						Value:      "|",
					},
					Right: &IntegerNode{
						SourceInfo: sourceInfo(1, 11),
						Value:      16,
					},
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(2, 3),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IntegerNode{
					SourceInfo: sourceInfo(2, 9),
					Value:      1,
				},
			},
		},
		{
			SourceInfo: sourceInfo(4, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IntegerNode{
					SourceInfo: sourceInfo(4, 10),
					Value:      2,
				},
			},
		},
		{
			SourceInfo: sourceInfo(6, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IntegerNode{
					SourceInfo: sourceInfo(6, 10),
					Value:      3,
				},
			},
		},
		{
			SourceInfo: sourceInfo(8, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IntegerNode{
					SourceInfo: sourceInfo(8, 10),
					Value:      4,
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(3, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IntegerNode{
					SourceInfo: sourceInfo(3, 10),
					Value:      1,
				},
			},
		},
		{
			SourceInfo: sourceInfo(3, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IntegerNode{
					SourceInfo: sourceInfo(3, 10),
					Value:      1,
				},
			},
		},
		{
			SourceInfo: sourceInfo(8, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IntegerNode{
					SourceInfo: sourceInfo(12, 6),
					Value:      2,
				},
			},
		},
		{
			SourceInfo: sourceInfo(13, 3),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IntegerNode{
					SourceInfo: sourceInfo(13, 9),
					Value:      3,
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(4, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IdentifierNode{
					SourceInfo: sourceInfo(4, 10),
					Identifier: identA,
				},
			},
		},
		{
			SourceInfo: sourceInfo(5, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&AddressNode{
					SourceInfo: sourceInfo(5, 11),
					Address:    0,
				},
			},
		},
		{
			SourceInfo: sourceInfo(4, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IdentifierNode{
					SourceInfo: sourceInfo(4, 10),
					Identifier: identA,
				},
			},
		},
		{
			SourceInfo: sourceInfo(5, 4),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&AddressNode{
					SourceInfo: sourceInfo(5, 11),
					Address:    2,
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(5, 3),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&StringNode{
					SourceInfo: sourceInfo(2, 17),
					Value:      "Hello World!",
				},
				&BinaryOpNode{
					SourceInfo: sourceInfo(3, 19),
					Left: &IntegerNode{
						SourceInfo: sourceInfo(3, 17),
						Value:      2,
					},
					Operator: &Token{
						SourceInfo: sourceInfo(3, 19),
						Type:       SHL,
						Value:      "<<",
					},
					Right: &IntegerNode{
						SourceInfo: sourceInfo(3, 22),
						Value:      1,
					},
				},
			},
		},
	}

	require.Equal(t, expectedCommands, commands)

	expectedDefines := map[string]ExpressionNode{
		"msgHello": &StringNode{SourceInfo: sourceInfo(2, 17), Value: "Hello World!"},
		"wordCount": &BinaryOpNode{
			SourceInfo: sourceInfo(3, 19),
			Left:       &IntegerNode{SourceInfo: sourceInfo(3, 17), Value: 2},
			Operator:   &Token{SourceInfo: sourceInfo(3, 19), Type: SHL, Value: "<<"},
			Right:      &IntegerNode{SourceInfo: sourceInfo(3, 22), Value: 1},
		},
	}

	require.Equal(t, expectedDefines, defines)

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
		{
			SourceInfo: sourceInfo(3, 3),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&IdentifierNode{
					SourceInfo: sourceInfo(3, 9),
					Identifier: VariableOffset,
				},
				&IntegerNode{
					SourceInfo: sourceInfo(3, 15),
					Value:      42,
				},
			},
		},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}

func TestParser_ArrayVariables(t *testing.T) {
	script := `
		var myArr[10]
		myCmd myArr[2] 42
	`

	commands, defines, labels, macros, err := parse(script)

	require.NoError(t, err)

	expectedCommands := []*CommandNode{
		{
			SourceInfo: sourceInfo(3, 3),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&ArrayAccessNode{
					SourceInfo: sourceInfo(3, 9),
					Variable:   VariableOffset,
					Index: &IntegerNode{
						SourceInfo: sourceInfo(3, 15),
						Value:      2,
					},
				},
				&IntegerNode{
					SourceInfo: sourceInfo(3, 18),
					Value:      42,
				},
			},
		},
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
		{
			SourceInfo: sourceInfo(2, 3),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&StringNode{
					SourceInfo: sourceInfo(2, 10),
					Value:      "Hello World!",
				},
			},
		},
		{
			SourceInfo: sourceInfo(3, 3),
			Type:       cmdMyCmd,
			Args: []ExpressionNode{
				&StringNode{
					SourceInfo: sourceInfo(3, 10),
					Value:      "Strings can .contain all @sorts of -42.1337 # characters",
				},
			},
		},
	}

	require.Equal(t, expectedCommands, commands)

	require.Empty(t, defines)
	require.Empty(t, labels)
	require.Empty(t, macros)
}
