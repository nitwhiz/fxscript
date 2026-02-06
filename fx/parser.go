package fx

import (
	"strconv"
	"strings"
)

var eofToken = &Token{Type: EOF, Value: ""}

var _ TokenSource = (*Lexer)(nil)
var _ TokenSource = (*TokenIterator)(nil)
var _ TokenSource = (*TokenSlice)(nil)

type TokenSource interface {
	NextToken() (*Token, error)
	Filename() string
}

const (
	tokenPrefetch = 16
)

type ParserConfig struct {
	FS       *ParserFS
	LookupFn LookupFn

	CommandTypes CommandTypeTable
	Identifiers  IdentifierTable
	BufSize      int
}

type Parser struct {
	includedFiles map[string]bool
	src           *TokenIterator

	commandTypes CommandTypeTable
	identifiers  IdentifierTable

	done bool

	fs       *ParserFS
	lookupFn LookupFn
}

func NewParser(src TokenSource, c *ParserConfig) *Parser {
	bufSize := c.BufSize

	if bufSize == 0 {
		bufSize = 32
	}

	p := Parser{
		includedFiles: make(map[string]bool),
		src:           NewTokenIterator("main", src, bufSize),

		commandTypes: c.CommandTypes,
		identifiers:  c.Identifiers,

		fs:       c.FS,
		lookupFn: c.LookupFn,
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
	if tok, err = p.peek(); err != nil {
		return
	}

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

	addressNode := &AddressNode{
		Address:    0,
		SourceInfo: tok.SourceInfo,
	}

	script.addSymbol(labelName, addressNode)

	expr = addressNode

	return
}

func (p *Parser) resolveIdent(script *Script, tok *Token) (expr ExpressionNode, err error) {
	var ok bool

	if expr, ok = script.defines[tok.Value]; ok {
		return
	}

	var varIdent int

	if varIdent, ok = script.variables[tok.Value]; ok {
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

		expr = &FloatNode{
			Value:      val,
			SourceInfo: tok.SourceInfo,
		}

		return
	}

	var val int64

	if val, err = strconv.ParseInt(tok.Value, 10, 64); err != nil {
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

func (p *Parser) parseDefine(script *Script) (err error) {
	if _, err = p.advance(); err != nil {
		return
	}

	var nameIdent *Token

	if nameIdent, err = p.advance(); err != nil {
		return
	}

	script.defines[nameIdent.Value], err = p.parseExpression(script)

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

		if cmd.Type == CmdNone {
			cmd.SourceInfo = tok.SourceInfo
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
		p.src.SetPrefix(nameIdent.Value)
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
	case PREPROCESSOR:
		if err = p.parsePreprocessorDirective(); err != nil {
			return
		}
	case MACRO:
		if err = p.parseMacro(script); err != nil {
			return
		}
	case DEF:
		if err = p.parseDefine(script); err != nil {
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
		err = &SyntaxError{&UnexpectedTokenError{[]TokenType{end, MACRO, DEF, IDENT, NEWLINE}, tok}}
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
