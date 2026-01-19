package fx

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
	args map[string]int
	body *TokenSlice
}

func newMacro(signature []string, body []*Token) *Macro {
	args := make(map[string]int)

	for i, arg := range signature {
		args[arg] = i
	}

	return &Macro{
		args: args,
		body: newTokenSlice(body),
	}
}

func (m *Macro) Body(args [][]*Token) (*TokenSlice, error) {
	tokens := make([]*Token, 0, len(m.body.tokens))

	var argName string
	var ok bool

	for i := 0; i < len(m.body.tokens); i++ {
		argName, ok, i = macroArgToken(m.body.tokens, i)

		if ok {
			argIdx, ok := m.args[argName]

			if !ok {
				return nil, &SyntaxError{&UnknownMacroArgumentError{argName}}
			}

			if len(args) <= argIdx {
				return nil, &SyntaxError{&MissingMacroArgumentError{argName}}
			}

			tokens = append(tokens, args[argIdx]...)
		} else {
			tokens = append(tokens, m.body.tokens[i])
		}
	}

	return newTokenSlice(tokens), nil
}
