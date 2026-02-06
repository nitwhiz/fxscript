package fx

func (p *Parser) parsePrimary(script *Script) (expr ExpressionNode, err error) {
	var tok *Token

	if tok, err = p.advance(); err != nil {
		return
	}

	switch tok.Type {
	case NEWLINE:
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
		return p.resolveIdent(script, tok)
	default:
		err = &SyntaxError{&UnexpectedTokenError{[]TokenType{NEWLINE, ADD, SUB, MUL, EXCL, INV, AND, LPAREN, NUMBER, STRING, IDENT}, tok}}
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
