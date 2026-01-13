package fx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	cmdAccuracyCheck = UserCommandOffset + iota
	cmdHpUpdate
	cmdRecoil
	cmdPrint
)

const (
	attacker = iota
	recoilTypeMiss
)

const (
	multiHitCounter = iota
	upperMemoryStart
)

func NewTestParser(script string) *Parser {
	l := NewLexer([]byte(script))

	p := NewParser(l, &ParserConfig{
		CommandTypes: CommandTypeTable{
			"accuracyCheck": cmdAccuracyCheck,
			"hpUpdate":      cmdHpUpdate,
			"recoil":        cmdRecoil,
			"print":         cmdPrint,
		},
		Identifiers: IdentifierTable{
			"attacker":       attacker,
			"recoilTypeMiss": recoilTypeMiss,
		},
		Variables: VariableTable{
			"multiHitCounter":  multiHitCounter,
			"upperMemoryStart": upperMemoryStart,
		},
		Flags: nil,
	})

	return p
}

func TestParser_Const(t *testing.T) {
	script := `
		const msgHello "Hello World!"
		const wordCount 2
		const pi 3.14159265359
	`

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Empty(t, s.commands)

	c, ok := s.constants["msgHello"]

	require.True(t, ok, "expected constant msgHello to be defined")
	require.Equal(t, &StringNode{"Hello World!"}, c)

	c, ok = s.constants["wordCount"]

	require.True(t, ok, "expected constant wordCount to be defined")
	require.Equal(t, &IntegerNode{2}, c)

	c, ok = s.constants["pi"]

	require.True(t, ok, "expected constant pi to be defined")
	require.Equal(t, &FloatNode{3.14159265359}, c)
}

func TestParser_Command(t *testing.T) {
	script := `
		nop
	`

	expectedNodes := []*CommandNode{
		{CmdNop, nil},
	}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_CommandWithArgs(t *testing.T) {
	script := `
		nop -42.0 attacker "hello world" 33
	`

	expectedNodes := []*CommandNode{
		{CmdNop, []ExpressionNode{
			&UnaryOpNode{OpSub, &FloatNode{42.0}},
			&IdentifierNode{attacker},
			&StringNode{"hello world"},
			&IntegerNode{33},
		}},
	}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_CommandWithArgsWithConstants(t *testing.T) {
	script := `
		const msgHello "Hello World!"
		nop msgHello
	`

	expectedNodes := []*CommandNode{
		{CmdNop, []ExpressionNode{
			&StringNode{"Hello World!"},
		}},
	}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_Label(t *testing.T) {
	script := `
		effectHit:
			accuracyCheck
			hpUpdate
			goto end

		effectHitAndRecoil:
			accuracyCheck recoilMiss
			hpUpdate
			goto end

		secondLabel:
		recoilMiss:
			recoil recoilTypeMiss
			goto end

		goto secondLabel

		end:
			nop
			set multiHitCounter 3
	`

	expectedNodes := []*CommandNode{
		{cmdAccuracyCheck, nil},
		{cmdHpUpdate, nil},
		{CmdGoto, []ExpressionNode{&AddressNode{9}}},
		{cmdAccuracyCheck, []ExpressionNode{&AddressNode{6}}},
		{cmdHpUpdate, nil},
		{CmdGoto, []ExpressionNode{&AddressNode{9}}},
		{cmdRecoil, []ExpressionNode{&IdentifierNode{recoilTypeMiss}}},
		{CmdGoto, []ExpressionNode{&AddressNode{9}}},
		{CmdGoto, []ExpressionNode{&AddressNode{6}}},
		{CmdNop, nil},
		{CmdSet, []ExpressionNode{&VariableNode{multiHitCounter}, &IntegerNode{3}}},
	}

	expectedLabelPtrs := map[string]int{
		"effectHit":          0,
		"effectHitAndRecoil": 3,
		"secondLabel":        6,
		"recoilMiss":         6,
		"end":                9,
	}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)

	for k, expected := range expectedLabelPtrs {
		ptr, ok := s.Label(k)

		require.True(t, ok, "expected label %s to be defined", k)
		require.Equal(t, expected, ptr)
	}
}

func TestParser_Macro(t *testing.T) {
	script := `
		macro one
			accuracyCheck failed
			hpUpdate
		endmacro

		macro two
			one
			accuracyCheck
			hpUpdate
		endmacro

		two
		failed:
			one
	`

	expectedNodes := []*CommandNode{
		{cmdAccuracyCheck, []ExpressionNode{&AddressNode{4}}},
		{cmdHpUpdate, nil},
		{cmdAccuracyCheck, nil},
		{cmdHpUpdate, nil},
		{cmdAccuracyCheck, []ExpressionNode{&AddressNode{4}}},
		{cmdHpUpdate, nil},
	}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArgStringAndNumber(t *testing.T) {
	script := "hostCall \"hello\" -42\n"

	expectedNodes := []*CommandNode{{CmdHostCall, []ExpressionNode{&BinaryOpNode{&StringNode{"hello"}, OpSub, &IntegerNode{42}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg1(t *testing.T) {
	script := "accuracyCheck 2 + 4\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg2(t *testing.T) {
	script := "accuracyCheck (2 + 4)\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg3(t *testing.T) {
	script := "accuracyCheck (2 - -4)\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpSub, &UnaryOpNode{OpSub, &IntegerNode{4}}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg4(t *testing.T) {
	script := "accuracyCheck -42\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&UnaryOpNode{OpSub, &IntegerNode{42}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg5(t *testing.T) {
	script := "accuracyCheck (2 + 4) -42\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}}, OpSub, &IntegerNode{42}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg6(t *testing.T) {
	script := "accuracyCheck 2 * 4\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpMul, &IntegerNode{4}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg7(t *testing.T) {
	script := "accuracyCheck 2 + 4 * 8\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &BinaryOpNode{&IntegerNode{4}, OpMul, &IntegerNode{8}}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg8(t *testing.T) {
	script := "accuracyCheck (2 + 4) * 8\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}}, OpMul, &IntegerNode{8}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg9(t *testing.T) {
	script := "accuracyCheck 2 * 4 + 8\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpMul, &IntegerNode{4}}, OpAdd, &IntegerNode{8}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsCommandArg10(t *testing.T) {
	script := "accuracyCheck 2 * -4 + 8\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpMul, &UnaryOpNode{OpSub, &IntegerNode{4}}}, OpAdd, &IntegerNode{8}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg1(t *testing.T) {
	script := "accuracyCheck 2 +4\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg2(t *testing.T) {
	script := "accuracyCheck (2+4)\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg3(t *testing.T) {
	script := "accuracyCheck (2--4)\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpSub, &UnaryOpNode{OpSub, &IntegerNode{4}}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg4(t *testing.T) {
	script := "accuracyCheck -42\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&UnaryOpNode{OpSub, &IntegerNode{42}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg5(t *testing.T) {
	script := "accuracyCheck (2+ 4)-42\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}}, OpSub, &IntegerNode{42}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg6(t *testing.T) {
	script := "accuracyCheck 2*4\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpMul, &IntegerNode{4}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg7(t *testing.T) {
	script := "accuracyCheck 2+ 4 *8\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &BinaryOpNode{&IntegerNode{4}, OpMul, &IntegerNode{8}}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg8(t *testing.T) {
	script := "accuracyCheck (2+4)*8\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}}, OpMul, &IntegerNode{8}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg9(t *testing.T) {
	script := "accuracyCheck 2*4+8\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpMul, &IntegerNode{4}}, OpAdd, &IntegerNode{8}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_WeirdlyFormattedExpressionAsCommandArg10(t *testing.T) {
	script := "accuracyCheck 2*-4+8\n"

	expectedNodes := []*CommandNode{{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpMul, &UnaryOpNode{OpSub, &IntegerNode{4}}}, OpAdd, &IntegerNode{8}}}}}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}

func TestParser_ExpressionAsConstValue(t *testing.T) {
	script := `
		const val1 2 + 4
		const val2 2 * 4
		const val3 2 + 4 * 8
		const val4 (2 + 4) * 8
		const val5 2 * 4 + 8
		const val6 2 * -4 + 8
	`

	expectedConstants := map[string]ExpressionNode{
		"val1": &BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}},
		"val2": &BinaryOpNode{&IntegerNode{2}, OpMul, &IntegerNode{4}},
		"val3": &BinaryOpNode{&IntegerNode{2}, OpAdd, &BinaryOpNode{&IntegerNode{4}, OpMul, &IntegerNode{8}}},
		"val4": &BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpAdd, &IntegerNode{4}}, OpMul, &IntegerNode{8}},
		"val5": &BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpMul, &IntegerNode{4}}, OpAdd, &IntegerNode{8}},
		"val6": &BinaryOpNode{&BinaryOpNode{&IntegerNode{2}, OpMul, &UnaryOpNode{OpSub, &IntegerNode{4}}}, OpAdd, &IntegerNode{8}},
	}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Len(t, s.commands, 0)
	require.Equal(t, expectedConstants, s.constants)
}

func TestParser_ExpressionWithLabelAndIdentifier(t *testing.T) {
	script := `
		accuracyCheck test + 2
		accuracyCheck (test * (upperMemoryStart + 3))
		test:
			print "hey!"
	`

	expectedNodes := []*CommandNode{
		{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&AddressNode{2}, OpAdd, &IntegerNode{2}}}},
		{cmdAccuracyCheck, []ExpressionNode{&BinaryOpNode{&AddressNode{2}, OpMul, &BinaryOpNode{&VariableNode{upperMemoryStart}, OpAdd, &IntegerNode{3}}}}},
		{cmdPrint, []ExpressionNode{&StringNode{"hey!"}}},
	}

	p := NewTestParser(script)

	s, err := p.Parse()

	require.NoError(t, err)
	require.Equal(t, expectedNodes, s.commands)
}
