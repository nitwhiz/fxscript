package fx

import (
	"strconv"
	"strings"
)

func (p *Parser) parsePrimary(script *Script) (expr ExpressionNode, err error) {
	var tok *Token

	if tok, err = p.advance(); err != nil {
		return
	}

	switch tok.Type {
	case NEWLINE, RBRACKET:
		return
	case ADD, SUB, MUL, EXCL, INV, AND:
		return p.parseUnary(script, tok)
	case LPAREN:
		return p.parseExpressionInParens(script)
	case NUMBER:
		return p.parseNumber(tok)
	case STRING:
		return p.parseString(tok)
	case PERCENT:
		return p.parseLabel(script, tok)
	case IDENT:
		return p.parseExpressionIdent(script, tok)
	default:
		err = &SyntaxError{tok.SourceInfo, &UnexpectedTokenError{[]TokenType{NEWLINE, RBRACKET, ADD, SUB, MUL, EXCL, INV, AND, LPAREN, NUMBER, STRING, IDENT}, tok}}
		return
	}
}

func (p *Parser) parseExpression(script *Script) (ExpressionNode, error) {
	return p.parseEquality(script)
}

func (p *Parser) parseMultiplicative(script *Script) (expr ExpressionNode, err error) {
	return p.parseBinary(script, p.parsePrimary, MUL, DIV, PERCENT)
}

func (p *Parser) parseAdditive(script *Script) (expr ExpressionNode, err error) {
	return p.parseBinary(script, p.parseMultiplicative, ADD, SUB, AND, OR, INV)
}

func (p *Parser) parseEquality(script *Script) (ExpressionNode, error) {
	return p.parseBinary(script, p.parseComparison, EQ, NEQ)
}

func (p *Parser) parseComparison(script *Script) (expr ExpressionNode, err error) {
	return p.parseBinary(script, p.parseShift, LT, GT, LTE, GTE)
}

func (p *Parser) parseShift(script *Script) (expr ExpressionNode, err error) {
	return p.parseBinary(script, p.parseAdditive, SHL, SHR)
}

func (p *Parser) parseBinary(script *Script, next func(script *Script) (ExpressionNode, error), ops ...TokenType) (expr ExpressionNode, err error) {
	if expr, err = next(script); err != nil {
		return
	}

	var current *Token

	for {
		if current, err = p.peek(); err != nil {
			return
		}

		if !contains(ops, current.Type) {
			break
		}

		op := current

		if _, err = p.advance(); err != nil {
			return
		}

		var right ExpressionNode

		if right, err = next(script); err != nil {
			return
		}

		expr = &BinaryOpNode{
			Left:       expr,
			Operator:   op,
			Right:      right,
			SourceInfo: current.SourceInfo,
		}
	}

	return
}

func contains(ops []TokenType, tokType TokenType) bool {
	for _, op := range ops {
		if op == tokType {
			return true
		}
	}

	return false
}

func (p *Parser) parseExpressionIdent(script *Script, tok *Token) (expr ExpressionNode, err error) {
	var ok bool

	if expr, ok = script.defines[tok.Value]; ok {
		return
	}

	var varIdent int

	if varIdent, ok = script.variables[tok.Value]; ok {
		var nextToken *Token

		if nextToken, err = p.peek(); err != nil {
			return
		}

		if nextToken.Type == LBRACKET {
			expr, err = p.parseArrayAccess(script, tok, varIdent)
			return
		}

		expr = &IdentifierNode{
			Identifier: Identifier(varIdent),
			SourceInfo: tok.SourceInfo,
		}
		return
	}

	var identifier Identifier

	if identifier, ok = p.getIdentifier(tok.Value); ok {
		expr = &IdentifierNode{
			Identifier: identifier,
			SourceInfo: tok.SourceInfo,
		}
		return
	}

	return p.parseLabel(script, tok)
}

func (p *Parser) parseUnary(script *Script, tok *Token) (expr *UnaryOpNode, err error) {
	var operand ExpressionNode

	if operand, err = p.parsePrimary(script); err != nil {
		return
	}

	expr = &UnaryOpNode{
		Operator:   tok,
		Expr:       operand,
		SourceInfo: tok.SourceInfo,
	}
	return
}

func (p *Parser) parseExpressionInParens(script *Script) (expr ExpressionNode, err error) {
	if expr, err = p.parseExpression(script); err != nil {
		return
	}

	var tok *Token

	if tok, err = p.advance(); err != nil {
		return
	}

	if tok.Type != RPAREN {
		err = &SyntaxError{tok.SourceInfo, &UnexpectedTokenError{[]TokenType{RPAREN}, tok}}
		return
	}

	return
}

func (p *Parser) parseNumber(tok *Token) (expr ExpressionNode, err error) {
	if strings.Contains(tok.Value, ".") {
		var val float64

		if val, err = strconv.ParseFloat(tok.Value, 64); err != nil {
			return
		}

		expr = &FloatNode{
			Value:      val,
			SourceInfo: tok.SourceInfo,
		}

		return
	}

	var val int64

	intBase := 10
	intValue := tok.Value

	if len(intValue) > 2 {
		switch intValue[1] {
		case 'x':
			intBase = 16
		case 'o':
			intBase = 8
		case 'b':
			intBase = 2
		}
	}

	if intBase != 10 {
		intValue = intValue[2:]
	}

	if val, err = strconv.ParseInt(intValue, intBase, 64); err != nil {
		return
	}

	expr = &IntegerNode{
		Value:      int(val),
		SourceInfo: tok.SourceInfo,
	}

	return
}

func (p *Parser) parseString(tok *Token) (expr ExpressionNode, err error) {
	expr = &StringNode{
		Value:      tok.Value,
		SourceInfo: tok.SourceInfo,
	}
	return
}
