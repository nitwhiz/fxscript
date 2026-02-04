package fx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer_Comments(t *testing.T) {
	script := `
		# this is a 42.1337 # comment
		cmd1 arg1 # this is another comment with "stuff" in it
	`

	expectedTokens := []*Token{
		{IDENT, "cmd1"},
		{IDENT, "arg1"},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Ident(t *testing.T) {
	script := `
		cmd1 arg1
		cmd2
	`

	expectedTokens := []*Token{
		{IDENT, "cmd1"},
		{IDENT, "arg1"},
		{NEWLINE, ""},
		{IDENT, "cmd2"},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Numbers(t *testing.T) {
	script := "42 -15 +39.55 -42.0\n"

	expectedTokens := []*Token{
		{NUMBER, "42"},
		{SUB, "-"},
		{NUMBER, "15"},
		{ADD, "+"},
		{NUMBER, "39.55"},
		{SUB, "-"},
		{NUMBER, "42.0"},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_OperatorsAndNumbers(t *testing.T) {
	script := "+42 + 13 - 37 * -72 / 42\n"

	expectedTokens := []*Token{
		{ADD, "+"},
		{NUMBER, "42"},
		{ADD, "+"},
		{NUMBER, "13"},
		{SUB, "-"},
		{NUMBER, "37"},
		{MUL, "*"},
		{SUB, "-"},
		{NUMBER, "72"},
		{DIV, "/"},
		{NUMBER, "42"},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_OperatorsWithParens(t *testing.T) {
	script := "(+42 + 13) - (37 * (-72 / 42 ))\n"

	expectedTokens := []*Token{
		{LPAREN, ""},
		{ADD, "+"},
		{NUMBER, "42"},
		{ADD, "+"},
		{NUMBER, "13"},
		{RPAREN, ""},
		{SUB, "-"},
		{LPAREN, ""},
		{NUMBER, "37"},
		{MUL, "*"},
		{LPAREN, ""},
		{SUB, "-"},
		{NUMBER, "72"},
		{DIV, "/"},
		{NUMBER, "42"},
		{RPAREN, ""},
		{RPAREN, ""},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_InvOperator(t *testing.T) {
	script := "^42 ^-13\n"

	expectedTokens := []*Token{
		{INV, "^"},
		{NUMBER, "42"},
		{INV, "^"},
		{SUB, "-"},
		{NUMBER, "13"},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_AddrOfOperator(t *testing.T) {
	script := "&42 &13\n"

	expectedTokens := []*Token{
		{AND, "&"},
		{NUMBER, "42"},
		{AND, "&"},
		{NUMBER, "13"},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_And(t *testing.T) {
	script := "4 & 16\n"

	expectedTokens := []*Token{
		{NUMBER, "4"},
		{AND, "&"},
		{NUMBER, "16"},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Or(t *testing.T) {
	script := "4 | 16\n"

	expectedTokens := []*Token{
		{NUMBER, "4"},
		{OR, "|"},
		{NUMBER, "16"},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Labels(t *testing.T) {
	script := `
		some-label:
		%someLabel2:
	`

	expectedTokens := []*Token{
		{IDENT, "some-label"},
		{COLON, ""},
		{NEWLINE, ""},
		{PERCENT, "%"},
		{IDENT, "someLabel2"},
		{COLON, ""},
		{NEWLINE, ""},
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
		{MACRO, ""},
		{IDENT, "myMacro"},
		{NEWLINE, ""},
		{IDENT, "hello"},
		{IDENT, "world"},
		{NEWLINE, ""},
		{ENDMACRO, ""},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Defines(t *testing.T) {
	script := `
		def msgHello "Hello World!"
		def wordCount 2
	`

	expectedTokens := []*Token{
		{DEF, ""},
		{IDENT, "msgHello"},
		{STRING, "Hello World!"},
		{NEWLINE, ""},
		{DEF, ""},
		{IDENT, "wordCount"},
		{NUMBER, "2"},
		{NEWLINE, ""},
		{EOF, ""},
	}

	l := NewLexer([]byte(script))

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Variables(t *testing.T) {
	script := `
		var myVar
	`

	expectedTokens := []*Token{
		{VAR, ""},
		{IDENT, "myVar"},
		{NEWLINE, ""},
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
		{STRING, "Hello World!"},
		{NEWLINE, ""},
		{STRING, "Strings can .contain all @sorts of -42.1337 # characters"},
		{NEWLINE, ""},
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

	tok, err := l.NextToken()

	require.NoError(t, err)

	require.Equal(t, tokens[0], tok, "expected same EOF token to be returned")
}
