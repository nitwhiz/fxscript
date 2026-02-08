package fx

import (
	"fmt"
	"strconv"
)

type SourceInfo struct {
	Filename string
	Line     int
	Column   int
}

func (s *SourceInfo) String() string {
	var fName string

	if s.Filename == "" {
		fName = "<script>"
	} else {
		fName = s.Filename
	}

	return fmt.Sprintf("%s:%d:%d", fName, s.Line, s.Column)
}

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
	case DEF:
		return "DEF"
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
	case LBRACKET:
		return "LBRACKET"
	case RBRACKET:
		return "RBRACKET"
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
	case PREPROCESSOR:
		return "PREPROCESSOR"
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

	DEF
	VAR

	MACRO
	ENDMACRO

	LPAREN
	RPAREN
	LBRACKET
	RBRACKET

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

	PREPROCESSOR
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
	*SourceInfo
	Type  TokenType
	Value string
}

func (t *Token) String() string {
	return fmt.Sprintf("%s(%s)", t.Type, t.Value)
}

var identKeywords = map[string]TokenType{
	"var":      VAR,
	"def":      DEF,
	"macro":    MACRO,
	"endmacro": ENDMACRO,
}

func (l *Lexer) newToken(typ TokenType, value string) *Token {
	return &Token{
		Type:  typ,
		Value: value,
		SourceInfo: &SourceInfo{
			Filename: l.Filename(),
			Line:     l.line,
			Column:   l.col - len(value) + 1,
		},
	}
}
