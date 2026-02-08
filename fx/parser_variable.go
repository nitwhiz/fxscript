package fx

import "fmt"

func (p *Parser) evalStaticVariableBracketExpression(script *Script) (v int, err error) {
	var expr ExpressionNode
	var firstTokenInBrackets *Token

	if expr, firstTokenInBrackets, err = p.parseVariableBracketExpression(script); err != nil {
		return
	}

	evalValue, evalErr := script.Eval(expr, func(identifier Identifier) any {
		err = &ParseError{firstTokenInBrackets.SourceInfo, &UnresolvedSymbolError{fmt.Sprintf("%d", identifier)}}
		return 0
	})

	if evalErr != nil {
		return 0, evalErr
	}

	var ok bool

	if v, ok = evalValue.(int); !ok {
		err = &ParseError{firstTokenInBrackets.SourceInfo, &UnexpectedTypeError{fmt.Sprintf("%T", evalValue)}}
	}

	return
}

func (p *Parser) parseVariableBracketExpression(script *Script) (expr ExpressionNode, ft *Token, err error) {
	var tok *Token

	if tok, err = p.peek(); err != nil {
		return
	}

	if tok.Type == LBRACKET {
		if tok, err = p.advance(); err != nil {
			return
		}
	}

	ft = tok

	expr, err = p.parseExpression(script)

	if tok, err = p.peek(); err != nil {
		return
	}

	if tok.Type != RBRACKET {
		err = &SyntaxError{tok.SourceInfo, &UnexpectedTokenError{[]TokenType{RBRACKET}, tok}}
		return
	}

	if _, err = p.advance(); err != nil {
		return
	}

	return
}

func (p *Parser) parseVariableDeclaration(script *Script) (err error) {
	if _, err = p.advance(); err != nil {
		return
	}

	var nameIdent *Token

	if nameIdent, err = p.advance(); err != nil {
		return
	}

	offset := script.addVariable(nameIdent.Value)

	next, err := p.peek()

	if err != nil {
		return
	}

	if next.Type == LBRACKET {
		var l int

		if l, err = p.evalStaticVariableBracketExpression(script); err != nil {
			return
		}

		// skip zero
		for i := range l - 1 {
			script.addVariableWithOffset(fmt.Sprintf("__%s_%d", nameIdent.Value, i+1), offset+i+1)
		}
	}

	return
}

func (p *Parser) parseArrayAccess(script *Script, identToken *Token, varIdent int) (expr ExpressionNode, err error) {
	expr, _, err = p.parseVariableBracketExpression(script)

	expr = &ArrayAccessNode{
		SourceInfo: identToken.SourceInfo,
		Variable:   Identifier(varIdent),
		Index:      expr,
	}

	return
}
