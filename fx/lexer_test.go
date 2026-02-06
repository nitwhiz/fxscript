package fx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func tok(line, col int, typ TokenType, val string) *Token {
	return &Token{
		SourceInfo: &SourceInfo{
			Filename: "test.fx",
			Line:     line,
			Column:   col,
		},
		Type:  typ,
		Value: val,
	}
}

func TestLexer_Comments(t *testing.T) {
	script := `
		# this is a 42.1337 # comment
		cmd1 arg1 # this is another comment with "stuff" in it
	`

	expectedTokens := []*Token{
		tok(3, 3, IDENT, "cmd1"),
		tok(3, 8, IDENT, "arg1"),
		tok(4, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Ident(t *testing.T) {
	script := `
		cmd1 arg1
		cmd2
	`

	expectedTokens := []*Token{
		tok(2, 3, IDENT, "cmd1"),
		tok(2, 8, IDENT, "arg1"),
		tok(3, 1, NEWLINE, ""),
		tok(3, 3, IDENT, "cmd2"),
		tok(4, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Numbers(t *testing.T) {
	script := "42 -15 +39.55 -42.0\n"

	expectedTokens := []*Token{
		tok(1, 1, NUMBER, "42"),
		tok(1, 4, SUB, "-"),
		tok(1, 5, NUMBER, "15"),
		tok(1, 8, ADD, "+"),
		tok(1, 9, NUMBER, "39.55"),
		tok(1, 15, SUB, "-"),
		tok(1, 16, NUMBER, "42.0"),
		tok(2, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_OperatorsAndNumbers(t *testing.T) {
	script := "+42 + 13 - 37 * -72 / 42\n"

	expectedTokens := []*Token{
		tok(1, 1, ADD, "+"),
		tok(1, 2, NUMBER, "42"),
		tok(1, 5, ADD, "+"),
		tok(1, 7, NUMBER, "13"),
		tok(1, 10, SUB, "-"),
		tok(1, 12, NUMBER, "37"),
		tok(1, 15, MUL, "*"),
		tok(1, 17, SUB, "-"),
		tok(1, 18, NUMBER, "72"),
		tok(1, 21, DIV, "/"),
		tok(1, 23, NUMBER, "42"),
		tok(2, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_OperatorsWithParens(t *testing.T) {
	script := "(+42 + 13) - (37 * (-72 / 42 ))\n"

	expectedTokens := []*Token{
		tok(1, 2, LPAREN, ""),
		tok(1, 2, ADD, "+"),
		tok(1, 3, NUMBER, "42"),
		tok(1, 6, ADD, "+"),
		tok(1, 8, NUMBER, "13"),
		tok(1, 11, RPAREN, ""),
		tok(1, 12, SUB, "-"),
		tok(1, 15, LPAREN, ""),
		tok(1, 15, NUMBER, "37"),
		tok(1, 18, MUL, "*"),
		tok(1, 21, LPAREN, ""),
		tok(1, 21, SUB, "-"),
		tok(1, 22, NUMBER, "72"),
		tok(1, 25, DIV, "/"),
		tok(1, 27, NUMBER, "42"),
		tok(1, 31, RPAREN, ""),
		tok(1, 32, RPAREN, ""),
		tok(2, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_InvOperator(t *testing.T) {
	script := "^42 ^-13\n"

	expectedTokens := []*Token{
		tok(1, 1, INV, "^"),
		tok(1, 2, NUMBER, "42"),
		tok(1, 5, INV, "^"),
		tok(1, 6, SUB, "-"),
		tok(1, 7, NUMBER, "13"),
		tok(2, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_AddrOfOperator(t *testing.T) {
	script := "&42 &13\n"

	expectedTokens := []*Token{
		tok(1, 1, AND, "&"),
		tok(1, 2, NUMBER, "42"),
		tok(1, 5, AND, "&"),
		tok(1, 6, NUMBER, "13"),
		tok(2, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_And(t *testing.T) {
	script := "4 & 16\n"

	expectedTokens := []*Token{
		tok(1, 1, NUMBER, "4"),
		tok(1, 3, AND, "&"),
		tok(1, 5, NUMBER, "16"),
		tok(2, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Or(t *testing.T) {
	script := "4 | 16\n"

	expectedTokens := []*Token{
		tok(1, 1, NUMBER, "4"),
		tok(1, 3, OR, "|"),
		tok(1, 5, NUMBER, "16"),
		tok(2, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Labels(t *testing.T) {
	script := `
		some-label:
		%someLabel2:
	`

	expectedTokens := []*Token{
		tok(2, 3, IDENT, "some-label"),
		tok(2, 14, COLON, ""),
		tok(3, 1, NEWLINE, ""),
		tok(3, 3, PERCENT, "%"),
		tok(3, 4, IDENT, "someLabel2"),
		tok(3, 15, COLON, ""),
		tok(4, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

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
		tok(2, 8, MACRO, ""),
		tok(2, 9, IDENT, "myMacro"),
		tok(3, 1, NEWLINE, ""),
		tok(3, 4, IDENT, "hello"),
		tok(3, 10, IDENT, "world"),
		tok(4, 1, NEWLINE, ""),
		tok(4, 11, ENDMACRO, ""),
		tok(5, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Defines(t *testing.T) {
	script := `
		def msgHello "Hello World!"
		def wordCount 2
	`

	expectedTokens := []*Token{
		tok(2, 6, DEF, ""),
		tok(2, 7, IDENT, "msgHello"),
		tok(2, 17, STRING, "Hello World!"),
		tok(3, 1, NEWLINE, ""),
		tok(3, 6, DEF, ""),
		tok(3, 7, IDENT, "wordCount"),
		tok(3, 17, NUMBER, "2"),
		tok(4, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Variables(t *testing.T) {
	script := `
		var myVar
	`

	expectedTokens := []*Token{
		tok(2, 6, VAR, ""),
		tok(2, 7, IDENT, "myVar"),
		tok(3, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_Strings(t *testing.T) {
	script := `
		"Hello World!"
		"Strings can .contain all @sorts of -42.1337 # characters"
	`

	expectedTokens := []*Token{
		tok(2, 4, STRING, "Hello World!"),
		tok(3, 1, NEWLINE, ""),
		tok(3, 4, STRING, "Strings can .contain all @sorts of -42.1337 # characters"),
		tok(4, 1, NEWLINE, ""),
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}

	l := NewLexer([]byte(script), "test.fx")

	tokens := l.Lex()

	require.Equal(t, expectedTokens, tokens)
}

func TestLexer_EOF(t *testing.T) {
	l := NewLexer([]byte{}, "test.fx")

	tokens := l.Lex()

	require.Equal(t, []*Token{
		{
			SourceInfo: nil,
			Type:       EOF,
			Value:      "",
		},
	}, tokens)

	tok, err := l.NextToken()

	require.NoError(t, err)

	require.Equal(t, tokens[0], tok, "expected same EOF token to be returned")
}
