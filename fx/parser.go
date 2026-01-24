package fx

import (
	"strconv"
	"strings"
)

var eofToken = newToken(EOF, "")

type TokenSource interface {
	NextToken() (*Token, error)
}

const (
	tokenPrefetch = 16
)

type ParserConfig struct {
	CommandTypes CommandTypeTable
	Identifiers  IdentifierTable
	BufSize      int
}

type Parser struct {
	src *TokenIterator

	commandTypes CommandTypeTable
	identifiers  IdentifierTable

	done bool
}

func NewParser(src TokenSource, c *ParserConfig) *Parser {
	bufSize := c.BufSize

	if bufSize == 0 {
		bufSize = 32
	}

	p := Parser{
		src: NewTokenIterator("main", src, bufSize),

		commandTypes: c.CommandTypes,
		identifiers:  c.Identifiers,
	}

	return &p
}

func (p *Parser) peekAhead(n int) (*Token, error) {
	return p.src.Peek(n)
}

func (p *Parser) peek() (*Token, error) {
	return p.peekAhead(0)
}

func (p *Parser) advance() (tok *Token, err error) {
	return p.src.NextToken()
}

func (p *Parser) consumeUntil(end TokenType) (tokens []*Token, err error) {
	var tok *Token

	for {
		if tok, err = p.advance(); err != nil {
			return
		}

		if tok.Type == end || tok.Type == EOF {
			break
		}

		tokens = append(tokens, tok)
	}

	return
}

func (p *Parser) parseMacro(script *Script) (err error) {
	if _, err = p.advance(); err != nil {
		return
	}

	var ident *Token

	if ident, err = p.advance(); err != nil {
		return
	}

	var argTokens []*Token

	if argTokens, err = p.consumeUntil(NEWLINE); err != nil {
		return
	}

	args := make([]string, 0)

	var argName string
	var ok bool

	for i := 0; i < len(argTokens); i++ {
		argName, ok, i = macroArgToken(argTokens, i)

		if ok {
			args = append(args, argName)
		}
	}

	var macroTokens []*Token

	if macroTokens, err = p.consumeUntil(ENDMACRO); err != nil {
		return
	}

	script.macros[ident.Value] = newMacro(ident.Value, args, macroTokens)

	return
}

func (p *Parser) parseLabel(script *Script, tok *Token) (expr ExpressionNode, err error) {
	var labelName string

	if tok.Type == PERCENT {
		if tok, err = p.advance(); err != nil {
			return
		}

		if tok.Type == IDENT {
			labelName = p.src.Prefixed(tok.Value)
		}
	} else {
		labelName = tok.Value
	}

	addressNode := &AddressNode{0}

	script.addSymbol(labelName, addressNode)

	expr = addressNode

	return
}

func (p *Parser) resolveIdent(script *Script, tok *Token) (expr ExpressionNode, err error) {
	var ok bool

	if expr, ok = script.constants[tok.Value]; ok {
		return
	}

	var varIdent int

	if varIdent, ok = script.variables[tok.Value]; ok {
		expr = &IdentifierNode{Identifier(varIdent)}
		return
	}

	var identifier Identifier

	if identifier, ok = p.getIdentifier(tok.Value); ok {
		expr = &IdentifierNode{identifier}
		return
	}

	return p.parseLabel(script, tok)
}

func (p *Parser) parseUnary(script *Script, tok *Token) (expr *UnaryOpNode, err error) {
	var operand ExpressionNode

	if operand, err = p.parsePrimary(script); err != nil {
		return
	}

	expr = &UnaryOpNode{tok, operand}
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
		err = &SyntaxError{&UnexpectedTokenError{[]TokenType{RPAREN}, tok}}
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

		expr = &FloatNode{val}

		return
	}

	var val int64

	if val, err = strconv.ParseInt(tok.Value, 10, 32); err != nil {
		return
	}

	expr = &IntegerNode{int(val)}

	return
}

func (p *Parser) parseString(tok *Token) (expr ExpressionNode, err error) {
	expr = &StringNode{tok.Value}
	return
}

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
			Left:     expr,
			Operator: op,
			Right:    right,
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

func (p *Parser) parseConst(script *Script) (err error) {
	if _, err = p.advance(); err != nil {
		return
	}

	var nameIdent *Token

	if nameIdent, err = p.advance(); err != nil {
		return
	}

	script.constants[nameIdent.Value], err = p.parseExpression(script)

	return
}

func (p *Parser) parseVariable(script *Script) (err error) {
	if _, err = p.advance(); err != nil {
		return
	}

	var nameIdent *Token

	if nameIdent, err = p.advance(); err != nil {
		return
	}

	script.addVariable(nameIdent.Value)

	return
}

func (p *Parser) parseCommand(script *Script) (err error) {
	cmd := CommandNode{
		Type: CmdNone,
	}

	var tok *Token

	var macro *Macro
	var macroArgs [][]*Token

	for {
		if tok, err = p.peek(); err != nil {
			return
		}

		if tok.Type == NEWLINE || tok.Type == EOF {
			_, err = p.advance()

			if cmd.Type != CmdNone {
				script.commands = append(script.commands, &cmd)
			} else if macro != nil {
				var tokSrc TokenSource

				if tokSrc, err = macro.Body(macroArgs); err != nil {
					return
				}

				p.src.Insert(macro.name, tokSrc)
			}

			return
		}

		if cmd.Type == CmdNone && macro == nil {
			if _, err = p.advance(); err != nil {
				return
			}

			cmdType, ok := p.getCommandType(tok.Value)

			if !ok {
				m, ok := script.macros[tok.Value]

				if ok {
					macro = m
				} else {
					err = &SyntaxError{&UnknownCommandError{tok.Value}}
					return
				}
			} else {
				cmd.Type = cmdType
			}
		} else {
			if macro != nil {
				var argTokens []*Token

				for {
					ok := true

					if tok.Type == NEWLINE || tok.Type == EOF {
						ok = false
					}

					if tok.Type == COMMA || !ok {
						if len(argTokens) > 0 {
							macroArgs = append(macroArgs, argTokens)
							argTokens = []*Token{}
						}
					} else if ok {
						argTokens = append(argTokens, tok)
					}

					if !ok {
						break
					}

					if _, err = p.advance(); err != nil {
						return
					}

					if tok, err = p.peek(); err != nil {
						return
					}
				}
			} else {
				var argNode ExpressionNode

				if argNode, err = p.parseExpression(script); err != nil {
					return
				}

				if tok, err = p.peek(); err != nil {
					return err
				}

				if tok.Type == COMMA {
					if _, err = p.advance(); err != nil {
						return
					}
				}

				if argNode == nil {
					return
				}

				cmd.Args = append(cmd.Args, argNode)
			}
		}
	}
}

func (p *Parser) parseLabelDeclaration(script *Script, prefixed bool) (err error) {
	var nameIdent *Token

	if nameIdent, err = p.advance(); err != nil {
		return
	}

	if _, err = p.advance(); err != nil {
		return
	}

	var name string

	if prefixed {
		name = p.src.Prefixed(nameIdent.Value)
	} else {
		name = nameIdent.Value
	}

	script.labels[name] = script.PC()

	return
}

func (p *Parser) parseIdent(script *Script, tok *Token) (err error) {
	var tok1 *Token

	if tok1, err = p.peekAhead(1); err != nil {
		return
	}

	if tok.Type == PERCENT && tok1.Type == IDENT {
		var tok2 *Token

		if tok2, err = p.peekAhead(2); err != nil {
			return
		}

		if tok2.Type == COLON {
			if _, err = p.advance(); err != nil {
				return
			}

			err = p.parseLabelDeclaration(script, true)
			return
		}
	} else if tok.Type == IDENT && tok1.Type == COLON {
		err = p.parseLabelDeclaration(script, false)
		return
	}

	return p.parseCommand(script)
}

func (p *Parser) parseNextNode(script *Script, end TokenType) (ok bool, err error) {
	ok = true

	var tok *Token

	if tok, err = p.peek(); err != nil {
		ok = false
		return
	}

	switch tok.Type {
	case end:
		ok = false

		if _, err = p.advance(); err != nil {
			return
		}
	case MACRO:
		if err = p.parseMacro(script); err != nil {
			return
		}
	case CONST:
		if err = p.parseConst(script); err != nil {
			return
		}
	case VAR:
		if err = p.parseVariable(script); err != nil {
			return
		}
	case PERCENT, IDENT:
		if err = p.parseIdent(script, tok); err != nil {
			return
		}
	case NEWLINE:
		if _, err = p.advance(); err != nil {
			return
		}
	default:
		err = &SyntaxError{&UnexpectedTokenError{[]TokenType{end, MACRO, CONST, IDENT, NEWLINE}, tok}}
		return
	}

	return
}

func augmentAddressNodes(script *Script) (err error) {
	for label, addrNodes := range script.symbols {
		pc, ok := script.labels[label]

		if !ok {
			return &SyntaxError{&UnknownLabelError{label}}
		}

		for _, addr := range addrNodes {
			addr.Address = pc
		}
	}

	return
}

func (p *Parser) Parse() (script *Script, err error) {
	script = newScript()

	ok := true

	for ok {
		if ok, err = p.parseNextNode(script, EOF); err != nil {
			return
		}
	}

	err = augmentAddressNodes(script)

	return
}
