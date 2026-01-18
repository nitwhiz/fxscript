package fx

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '-'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t'
}

func isOperator(c byte) bool {
	switch string(c) {
	case SynPlus, SynMinus,
		SynAsterisk, SynSlash, SynPercent,
		SynExcl, SynInv, SynAmpersand, SynPipe,
		SynLower, SynGreater, SynEqual:
		return true
	}

	return false
}

type Lexer struct {
	source []byte

	pos       int
	sourceLen int

	lastToken *Token

	line int
	done bool
}

func NewLexer(source []byte) *Lexer {
	l := Lexer{
		source:    source,
		sourceLen: len(source),
	}

	l.Rewind()

	return &l
}

func (l *Lexer) Rewind() {
	l.pos = 0
	l.line = 1
}

func (l *Lexer) peekAhead(n int) byte {
	if l.pos+n >= l.sourceLen {
		return 0
	}

	return l.source[l.pos+n]
}

func (l *Lexer) peek() byte {
	return l.peekAhead(0)
}

func (l *Lexer) advance() byte {
	curr := l.peek()
	l.pos++

	return curr
}

func (l *Lexer) skipAhead(n int) {
	if l.pos+n >= l.sourceLen {
		l.pos = l.sourceLen
		return
	}

	l.pos += n
}

func (l *Lexer) substr(n int) string {
	var res string

	if l.peekAhead(n) == 0 {
		res = string(l.source[l.pos:l.sourceLen])
		l.pos = l.sourceLen
	} else {
		res = string(l.source[l.pos:(l.pos + n)])
		l.pos += n
	}

	return res
}

func (l *Lexer) skipWhitespace() (skipped bool) {
	n := 0

	for isWhitespace(l.peekAhead(n)) {
		n += 1
	}

	l.skipAhead(n)

	return n > 0
}

func (l *Lexer) skipComment() *Token {
	n := 1

	for l.peekAhead(n) != '\n' {
		n += 1
	}

	l.skipAhead(n)

	return l.lexNextToken()
}

func (l *Lexer) lexIdent() *Token {
	n := 0

	for isAlpha(l.peekAhead(n)) || (n > 0 && isDigit(l.peekAhead(n))) {
		n += 1
	}

	ident := l.substr(n)

	if tokTyp, ok := keywords[ident]; ok {
		return newToken(tokTyp, "")
	}

	return newToken(IDENT, ident)
}

func (l *Lexer) lexNumber() *Token {
	n := 0

	for {
		next := l.peekAhead(n)

		if isDigit(next) || next == '.' {
			n++
		} else {
			break
		}
	}

	return newToken(NUMBER, l.substr(n))
}

func (l *Lexer) lexOperator() *Token {
	var opVal string
	tokType := ILLEGAL

	switch string(l.peekAhead(1)) {
	case SynEqual, SynLower, SynGreater:
		opVal = string(l.advance()) + string(l.advance())
		break
	default:
		opVal = string(l.advance())
		break
	}

	switch opVal {
	case SynAmpersand:
		tokType = AND
		break
	case SynPipe:
		tokType = OR
		break
	case SynExcl:
		tokType = EXCL
		break
	case SynInv:
		tokType = INV
		break
	case SynPlus:
		tokType = ADD
		break
	case SynMinus:
		tokType = SUB
		break
	case SynAsterisk:
		tokType = MUL
		break
	case SynSlash:
		tokType = DIV
		break
	case SynPercent:
		tokType = MOD
		break
	case SynLower:
		tokType = LT
		break
	case SynGreater:
		tokType = GT
		break
	case SynLower + SynLower:
		tokType = SHL
		break
	case SynGreater + SynGreater:
		tokType = SHR
		break
	case SynLower + SynEqual:
		tokType = LTE
		break
	case SynGreater + SynEqual:
		tokType = GTE
		break
	case SynEqual + SynEqual:
		tokType = EQ
		break
	case SynExcl + SynEqual:
		tokType = NEQ
		break
	default:
		break
	}

	return newToken(tokType, opVal)
}

func (l *Lexer) lexString() *Token {
	l.advance()

	n := 0

	for l.peekAhead(n) != '"' {
		n += 1
	}

	token := newToken(STRING, l.substr(n))

	l.advance()

	return token
}

func (l *Lexer) lexNextToken() *Token {
	for l.skipWhitespace() {
	}

	c := l.peek()

	if c == 0 {
		l.done = true
		return newToken(EOF, "")
	}

	if c == ',' {
		l.advance()
		return newToken(COMMA, "")
	}

	if c == '\n' {
		l.advance()

		tok := newToken(NEWLINE, "")

		l.line += 1

		return tok
	}

	if c == ':' {
		l.advance()
		return newToken(COLON, "")
	}

	if c == '(' {
		l.advance()
		return newToken(LPAREN, "")
	}

	if c == ')' {
		l.advance()
		return newToken(RPAREN, "")
	}

	if isOperator(c) {
		return l.lexOperator()
	}

	if isDigit(c) {
		return l.lexNumber()
	}

	if c == '"' {
		return l.lexString()
	}

	if c == '#' {
		return l.skipComment()
	}

	if isAlpha(c) {
		return l.lexIdent()
	}

	return newToken(ILLEGAL, string(l.advance()))
}

func (l *Lexer) NextToken() *Token {
	if l.done {
		return l.lastToken
	}

lex:
	tok := l.lexNextToken()

	if (l.lastToken == nil && tok.Type == NEWLINE) || (l.lastToken != nil && l.lastToken.Type == NEWLINE && tok.Type == NEWLINE) {
		goto lex
	}

	l.lastToken = tok

	return tok
}

func (l *Lexer) Lex() []*Token {
	l.Rewind()

	var tokens []*Token

	for {
		tok := l.NextToken()

		tokens = append(tokens, tok)

		if tok.Type == EOF {
			break
		}

	}

	return tokens
}
