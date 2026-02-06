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

	filename string
	line     int
	col      int

	lastToken *Token

	done bool
}

func NewLexer(source []byte, filename string) *Lexer {
	l := Lexer{
		source:    source,
		sourceLen: len(source),
		pos:       0,

		filename: filename,
		line:     1,
		col:      0,
	}

	return &l
}

func (l *Lexer) Filename() string {
	return l.filename
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

	if curr == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}

	l.pos++

	return curr
}

func (l *Lexer) skipAhead(n int) {
	var b byte

	for range n {
		if b = l.advance(); b == 0 {
			break
		}
	}
}

func (l *Lexer) substr(n int) string {
	var res string

	if l.pos+n >= l.sourceLen {
		res = string(l.source[l.pos:l.sourceLen])
	} else {
		res = string(l.source[l.pos:(l.pos + n)])
	}

	var b byte

	for range n {
		if b = l.advance(); b == 0 {
			break
		}
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

	if tokTyp, ok := identKeywords[ident]; ok {
		return l.newToken(tokTyp, "")
	}

	return l.newToken(IDENT, ident)
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

	return l.newToken(NUMBER, l.substr(n))
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
		tokType = PERCENT
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

	return l.newToken(tokType, opVal)
}

func (l *Lexer) lexString() (token *Token) {
	l.advance()

	n := 0

	for l.peekAhead(n) != '"' {
		n += 1
	}

	token = l.newToken(STRING, l.substr(n))

	l.advance()

	return
}

func (l *Lexer) lexPreprocessor() (token *Token) {
	l.advance()

	n := 0

	for l.peekAhead(n) != '\n' {
		n += 1
	}

	token = l.newToken(PREPROCESSOR, l.substr(n))

	return
}

func (l *Lexer) lexNextToken() *Token {
	for l.skipWhitespace() {
	}

	c := l.peek()

	switch c {
	case 0:
		l.done = true
		return eofToken
	case '@':
		return l.lexPreprocessor()
	case ',':
		l.advance()
		return l.newToken(COMMA, "")
	case '\n':
		l.advance()

		tok := l.newToken(NEWLINE, "")

		return tok
	case ':':
		l.advance()
		return l.newToken(COLON, "")
	case '(':
		l.advance()
		return l.newToken(LPAREN, "")
	case ')':
		l.advance()
		return l.newToken(RPAREN, "")
	case '$':
		l.advance()
		return l.newToken(DOLLAR, "")
	case '"':
		return l.lexString()
	case '#':
		return l.skipComment()
	}

	if isOperator(c) {
		return l.lexOperator()
	}

	if isDigit(c) {
		return l.lexNumber()
	}

	if isAlpha(c) {
		return l.lexIdent()
	}

	return l.newToken(ILLEGAL, string(l.advance()))
}

func (l *Lexer) NextToken() (*Token, error) {
	if l.done {
		return l.lastToken, nil
	}

lex:
	tok := l.lexNextToken()

	if (l.lastToken == nil && tok.Type == NEWLINE) || (l.lastToken != nil && l.lastToken.Type == NEWLINE && tok.Type == NEWLINE) {
		goto lex
	}

	l.lastToken = tok

	return tok, nil
}

func (l *Lexer) Lex() []*Token {
	var tokens []*Token

	for {
		tok, _ := l.NextToken()

		tokens = append(tokens, tok)

		if tok.Type == EOF {
			break
		}

	}

	return tokens
}
