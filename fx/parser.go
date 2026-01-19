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
		src: NewTokenIterator(src, bufSize),

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
		tok, err = p.advance()

		if err != nil {
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

	ident, err := p.advance()

	if err != nil {
		return
	}

	argTokens, err := p.consumeUntil(NEWLINE)

	if err != nil {
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

	macroTokens, err := p.consumeUntil(ENDMACRO)

	if err != nil {
		return
	}

	script.macros[ident.Value] = newMacro(args, macroTokens)

	return
}

func (p *Parser) resolveIdent(script *Script, tok *Token) ExpressionNode {
	if expr, ok := script.constants[tok.Value]; ok {
		return expr
	}

	if identifier, ok := p.getIdentifier(tok.Value); ok {
		return &IdentifierNode{identifier}
	}

	addressNode := &AddressNode{0}

	script.symbols[tok.Value] = addressNode

	return addressNode
}

func (p *Parser) parsePrimary(script *Script) (expr ExpressionNode, err error) {
	tok, err := p.advance()

	if err != nil {
		return
	}

	switch tok.Type {
	case NEWLINE:
		return
	case ADD, SUB, MUL, EXCL, INV, AND:
		var operand ExpressionNode

		operand, err = p.parsePrimary(script)

		if err != nil {
			return
		}

		expr = &UnaryOpNode{tok, operand}
		return
	case LPAREN:
		expr, err = p.parseExpression(script)

		if err != nil {
			return
		}

		tok, err = p.advance()

		if err != nil {
			return
		}

		if tok.Type != RPAREN {
			err = &SyntaxError{&UnexpectedTokenError{[]TokenType{RPAREN}, tok}}
			return
		}

		return
	case NUMBER:
		if strings.Contains(tok.Value, ".") {
			var val float64

			val, err = strconv.ParseFloat(tok.Value, 64)

			if err != nil {
				return
			}

			expr = &FloatNode{val}
			return
		}

		var val int64

		val, err = strconv.ParseInt(tok.Value, 10, 32)

		if err != nil {
			return
		}

		expr = &IntegerNode{int(val)}
		return
	case STRING:
		expr = &StringNode{tok.Value}
		return
	case IDENT:
		expr = p.resolveIdent(script, tok)
		return
	default:
		err = &SyntaxError{&UnexpectedTokenError{[]TokenType{NEWLINE, ADD, SUB, MUL, EXCL, INV, AND, LPAREN, NUMBER, STRING, IDENT}, tok}}
		return
	}
}

func (p *Parser) parseExpression(script *Script) (ExpressionNode, error) {
	return p.parseEquality(script)
}

func (p *Parser) parseMultiplicative(script *Script) (expr ExpressionNode, err error) {
	return p.parseBinary(script, p.parsePrimary, MUL, DIV, MOD)
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
	expr, err = next(script)

	if err != nil {
		return
	}

	var current *Token

	for {
		current, err = p.peek()

		if err != nil {
			return
		}

		if !contains(ops, current.Type) {
			break
		}

		op := current

		_, err = p.advance()

		if err != nil {
			return
		}

		var right ExpressionNode

		right, err = next(script)

		if err != nil {
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

	nameIdent, err := p.advance()

	if err != nil {
		return
	}

	script.constants[nameIdent.Value], err = p.parseExpression(script)

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
		tok, err = p.peek()

		if err != nil {
			return
		}

		if tok.Type == NEWLINE || tok.Type == EOF {
			_, err = p.advance()

			if cmd.Type != CmdNone {
				script.commands = append(script.commands, &cmd)
			} else if macro != nil {
				var tokSrc TokenSource

				tokSrc, err = macro.Body(macroArgs)

				if err != nil {
					return
				}

				p.src.Insert(tokSrc)
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

					_, err = p.advance()

					if err != nil {
						return
					}

					tok, err = p.peek()

					if err != nil {
						return
					}
				}
			} else {
				var argNode ExpressionNode

				argNode, err = p.parseExpression(script)

				if err != nil {
					return
				}

				tok, err = p.peek()

				if err != nil {
					return
				}

				if tok.Type == COMMA {
					_, err = p.advance()

					if err != nil {
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

func (p *Parser) parseLabelDeclaration(script *Script) (err error) {
	nameIdent, err := p.advance()

	if err != nil {
		return
	}

	if _, err = p.advance(); err != nil {
		return
	}

	script.labels[nameIdent.Value] = script.PC()

	return
}

func (p *Parser) parseIdent(script *Script) (err error) {
	tok, err := p.peekAhead(1)

	if err != nil {
		return
	}

	if tok.Type == COLON {
		err = p.parseLabelDeclaration(script)
		return
	}

	return p.parseCommand(script)
}

func (p *Parser) parseNextNode(script *Script, end TokenType) (ok bool, err error) {
	ok = true

	tok, err := p.peek()

	if err != nil {
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
	case IDENT:
		if err = p.parseIdent(script); err != nil {
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
	for label, addr := range script.symbols {
		pc, ok := script.labels[label]

		if !ok {
			return &SyntaxError{&UnknownLabelError{label}}
		}

		addr.Address = pc
	}

	return
}

func (p *Parser) Parse() (script *Script, err error) {
	script = newScript()

	ok := true

	for ok {
		ok, err = p.parseNextNode(script, EOF)

		if err != nil {
			return nil, err
		}
	}

	err = augmentAddressNodes(script)

	return
}
