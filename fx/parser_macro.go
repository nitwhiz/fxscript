package fx

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

func macroArgToken(argTokens []*Token, offset int) (argName string, ok bool, nextOffset int) {
	argTokensLen := len(argTokens)
	tok := argTokens[offset]

	nextOffset = offset

	if tok.Type == DOLLAR && argTokensLen > offset+1 && argTokens[offset+1].Type == IDENT {
		argName = argTokens[offset+1].Value
		ok = true
		nextOffset++
	}

	return
}

type Macro struct {
	name string
	args map[string]int
	body *TokenSlice
}

func newMacro(name string, argumentNames []string, body []*Token) *Macro {
	args := make(map[string]int)

	for i, arg := range argumentNames {
		args[arg] = i
	}

	return &Macro{
		name: name,
		args: args,
		body: newTokenSlice(body),
	}
}

func tokWithParent(tok *Token, parent *SourceInfo, name string) *Token {
	return &Token{
		Type:  tok.Type,
		Value: tok.Value,
		SourceInfo: &SourceInfo{
			Name:   name,
			File:   tok.File,
			Line:   tok.Line,
			Column: tok.Column,
			Parent: parent,
		},
	}
}

func (m *Macro) Body(args [][]*Token, parent *Token) (*TokenSlice, error) {
	tokens := make([]*Token, 0, len(m.body.tokens))

	var argName string
	var ok bool

	for i := 0; i < len(m.body.tokens); i++ {
		argName, ok, i = macroArgToken(m.body.tokens, i)

		if ok {
			argIdx, ok := m.args[argName]

			if !ok {
				return nil, &SyntaxError{m.body.tokens[i].SourceInfo, &UnknownMacroArgumentError{argName}}
			}

			if len(args) <= argIdx {
				return nil, &SyntaxError{m.body.tokens[i].SourceInfo, &MissingMacroArgumentError{argName}}
			}

			for _, argTok := range args[argIdx] {
				tokens = append(tokens, tokWithParent(argTok, parent.SourceInfo, m.name))
			}
		} else {
			tokens = append(tokens, tokWithParent(m.body.tokens[i], parent.SourceInfo, m.name))
		}
	}

	return newTokenSlice(tokens), nil
}
