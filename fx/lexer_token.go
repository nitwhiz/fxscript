package fx

import (
	"fmt"
	"strconv"
)

type TokenType uint

func (t TokenType) String() string {
	switch t {
	case EOF:
		return "EOF"
	case ILLEGAL:
		return "ILLEGAL"
	case NEWLINE:
		return "NEWLINE"
	case COMMA:
		return "COMMA"
	case COLON:
		return "COLON"
	case STRING:
		return "STRING"
	case IDENT:
		return "IDENT"
	case NUMBER:
		return "NUMBER"
	case CONST:
		return "CONST"
	case VAR:
		return "VAR"
	case MACRO:
		return "MACRO"
	case ENDMACRO:
		return "ENDMACRO"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case ADD:
		return "ADD"
	case SUB:
		return "SUB"
	case MUL:
		return "MUL"
	case DIV:
		return "DIV"
	case SHL:
		return "SHL"
	case SHR:
		return "SHR"
	case LT:
		return "LT"
	case GT:
		return "GT"
	case LTE:
		return "LTE"
	case GTE:
		return "GTE"
	case EQ:
		return "EQ"
	case NEQ:
		return "NEQ"
	case EXCL:
		return "EXCL"
	case INV:
		return "INV"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case DOLLAR:
		return "DOLLAR"
	default:
		return "MISSINGNAME(" + strconv.Itoa(int(t)) + ")"
	}
}

const (
	EOF TokenType = iota
	ILLEGAL

	NEWLINE
	COMMA
	COLON

	STRING
	IDENT
	NUMBER

	CONST
	VAR

	MACRO
	ENDMACRO

	LPAREN
	RPAREN

	ADD
	SUB
	MUL
	DIV

	SHL
	SHR

	LT
	GT
	LTE
	GTE

	EQ
	NEQ

	EXCL
	INV

	AND
	OR

	DOLLAR
	PERCENT
)

const (
	SynPlus      = "+"
	SynMinus     = "-"
	SynAsterisk  = "*"
	SynSlash     = "/"
	SynPercent   = "%"
	SynExcl      = "!"
	SynLower     = "<"
	SynGreater   = ">"
	SynEqual     = "="
	SynInv       = "^"
	SynAmpersand = "&"
	SynPipe      = "|"
)

type Token struct {
	Type  TokenType
	Value string
}

func (t *Token) String() string {
	return fmt.Sprintf("%s(%s)", t.Type, t.Value)
}

var identKeywords = map[string]TokenType{
	"var":      VAR,
	"const":    CONST,
	"macro":    MACRO,
	"endmacro": ENDMACRO,
}

func newToken(typ TokenType, value string) *Token {
	return &Token{
		Type:  typ,
		Value: value,
	}
}
