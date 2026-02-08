package fx

import (
	"fmt"
)

type UnexpectedTokenError struct {
	Expected []TokenType
	Token    *Token
}

func (e *UnexpectedTokenError) Error() string {
	if len(e.Expected) == 0 {
		return fmt.Sprintf("unexpected token '%s'", e.Token)
	}

	return fmt.Sprintf("unexpected token '%s', expected %v", e.Token, e.Expected)
}

type UnknownCommandError struct {
	Command string
}

func (e *UnknownCommandError) Error() string {
	return fmt.Sprintf("unknown command: '%s'", e.Command)
}

type UnknownLabelError struct {
	Label string
}

func (e *UnknownLabelError) Error() string {
	return fmt.Sprintf("unknown label: '%s'", e.Label)
}

type UnknownPreprocessorDirectiveError struct {
	Directive string
}

func (e *UnknownPreprocessorDirectiveError) Error() string {
	return fmt.Sprintf("unknown preprocessor directive: '%s'", e.Directive)
}

type SyntaxError struct {
	*SourceInfo
	Err error
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error at %s: %s", e.SourceInfo, e.Err)
}

type UnknownOperatorError struct {
	TokenType TokenType
}

func (e *UnknownOperatorError) Error() string {
	return fmt.Sprintf("unknown operator: '%s'", e.TokenType)
}

type RuntimeError struct {
	*SourceInfo
	Err error
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("runtime error at %s: %s", e.SourceInfo, e.Err)
}

type UnexpectedBinaryOpError struct {
	Left  any
	Right any
}

func (e *UnexpectedBinaryOpError) Error() string {
	return fmt.Sprintf("unexpected binary operation with left operand '%v' and right operand '%v'", e.Left, e.Right)
}

type MisingMacroArgumentError struct {
	Name string
}

func (e *MisingMacroArgumentError) Error() string {
	return fmt.Sprintf("missing macro argument '%s'", e.Name)
}

type UnknownMacroArgumentError struct {
	Name string
}

func (e *UnknownMacroArgumentError) Error() string {
	return fmt.Sprintf("unknown macro argument '%s'", e.Name)
}

type MissingMacroArgumentError struct {
	Name string
}

func (e *MissingMacroArgumentError) Error() string {
	return fmt.Sprintf("missing macro argument '%s'", e.Name)
}

type UnresolvedSymbolError struct {
	Symbol string
}

func (e *UnresolvedSymbolError) Error() string {
	return fmt.Sprintf("unresolved symbol '%s'", e.Symbol)
}

type MissingLookupFnError struct {
	Lookup string
}

func (e *MissingLookupFnError) Error() string {
	return fmt.Sprintf("missing lookup function for '%s'", e.Lookup)
}

type LookupFnError struct {
	Err error
}

func (e *LookupFnError) Error() string {
	return fmt.Sprintf("lookup function error: %s", e.Err)
}

type InvalidPreprocessorValueError struct {
	Directive string
	Value     string
}

func (e *InvalidPreprocessorValueError) Error() string {
	return fmt.Sprintf("invalid preprocessor value for directive %s: %s", e.Directive, e.Value)
}

type ParseError struct {
	*SourceInfo
	Err error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at %s: %s", e.SourceInfo, e.Err)
}

type UnexpectedTypeError struct {
	TypeName string
}

func (e *UnexpectedTypeError) Error() string {
	return fmt.Sprintf("unexpected type '%s'", e.TypeName)
}
