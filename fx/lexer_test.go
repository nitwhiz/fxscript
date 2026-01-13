package fx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer_Ident(t *testing.T) {
	script := `
		cmd1 arg1
		cmd2
	`

	expectedTokens := []*Token{
		{IdentToken, "cmd1"},
		{IdentToken, "arg1"},
		{NewlineToken, ""},
		{IdentToken, "cmd2"},
		{NewlineToken, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Numbers(t *testing.T) {
	script := "42 -15 +39.55 -42.0\n"

	expectedTokens := []*Token{
		{NumberToken, "42"},
		{OperatorToken, "-"},
		{NumberToken, "15"},
		{OperatorToken, "+"},
		{NumberToken, "39.55"},
		{OperatorToken, "-"},
		{NumberToken, "42.0"},
		{NewlineToken, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_OperatorsAndNumbers(t *testing.T) {
	script := "+42 + 13 - 37 * -72 / 42\n"

	expectedTokens := []*Token{
		{OperatorToken, "+"},
		{NumberToken, "42"},
		{OperatorToken, "+"},
		{NumberToken, "13"},
		{OperatorToken, "-"},
		{NumberToken, "37"},
		{OperatorToken, "*"},
		{OperatorToken, "-"},
		{NumberToken, "72"},
		{OperatorToken, "/"},
		{NumberToken, "42"},
		{NewlineToken, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_OperatorsWithParens(t *testing.T) {
	script := "(+42 + 13) - (37 * (-72 / 42 ))\n"

	expectedTokens := []*Token{
		{LParenToken, ""},
		{OperatorToken, "+"},
		{NumberToken, "42"},
		{OperatorToken, "+"},
		{NumberToken, "13"},
		{RParenToken, ""},
		{OperatorToken, "-"},
		{LParenToken, ""},
		{NumberToken, "37"},
		{OperatorToken, "*"},
		{LParenToken, ""},
		{OperatorToken, "-"},
		{NumberToken, "72"},
		{OperatorToken, "/"},
		{NumberToken, "42"},
		{RParenToken, ""},
		{RParenToken, ""},
		{NewlineToken, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Labels(t *testing.T) {
	script := `
		some-label:
		someLabel2:
	`

	expectedTokens := []*Token{
		{IdentToken, "some-label"},
		{ColonToken, ""},
		{NewlineToken, ""},
		{IdentToken, "someLabel2"},
		{ColonToken, ""},
		{NewlineToken, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Macro(t *testing.T) {
	script := `
		macro myMacro
			hello world
		endmacro
	`

	expectedTokens := []*Token{
		{MacroToken, ""},
		{IdentToken, "myMacro"},
		{NewlineToken, ""},
		{IdentToken, "hello"},
		{IdentToken, "world"},
		{NewlineToken, ""},
		{EndMacroToken, ""},
		{NewlineToken, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Constants(t *testing.T) {
	script := `
		const msgHello "Hello World!"
		const wordCount 2
	`

	expectedTokens := []*Token{
		{ConstToken, ""},
		{IdentToken, "msgHello"},
		{StringToken, "Hello World!"},
		{NewlineToken, ""},
		{ConstToken, ""},
		{IdentToken, "wordCount"},
		{NumberToken, "2"},
		{NewlineToken, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Strings(t *testing.T) {
	script := `
		"Hello World!"
		"Strings can .contain all @sorts of -42.1337 # characters"
	`

	expectedTokens := []*Token{
		{StringToken, "Hello World!"},
		{NewlineToken, ""},
		{StringToken, "Strings can .contain all @sorts of -42.1337 # characters"},
		{NewlineToken, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_StringAndNumber(t *testing.T) {
	script := "\"Hello World!\" -42.0"

	expectedTokens := []*Token{
		{StringToken, "Hello World!"},
		{OperatorToken, "-"},
		{NumberToken, "42.0"},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_EOF(t *testing.T) {
	l := NewLexer([]byte{})

	tokens := l.Lex()

	require.Equal(t, []*Token{{EOF, ""}}, tokens)

	tok := l.NextToken()

	require.Equal(t, tokens[0], tok, "expected same EOF token to be returned")
}
