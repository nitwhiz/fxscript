package fx

type TokenType int

const (
	EOF TokenType = iota
	IllegalToken

	NewlineToken
	CommaToken
	IdentToken
	NumberToken
	ColonToken
	MacroToken
	EndMacroToken
	ConstToken
	StringToken
	OperatorToken
	LParenToken
	RParenToken
)

var keywords = map[string]TokenType{
	"const":    ConstToken,
	"macro":    MacroToken,
	"endmacro": EndMacroToken,
}

// todo: implement ~ or ^ as bitwise NOT

const (
	OpAdd = '+'
	OpSub = '-'
	OpMul = '*'
	OpDiv = '/'
	OpMod = '%'
)

type Token struct {
	Type  TokenType
	Value string
}

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
	return c == OpAdd || c == OpSub || c == OpMul || c == OpDiv || c == OpMod
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

func (l *Lexer) newToken(typ TokenType, value string) *Token {
	return &Token{
		Type:  typ,
		Value: value,
	}
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
		return l.newToken(tokTyp, "")
	}

	return l.newToken(IdentToken, ident)
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

	return l.newToken(NumberToken, l.substr(n))
}

func (l *Lexer) lexOperator() *Token {
	return l.newToken(OperatorToken, string(l.advance()))
}

func (l *Lexer) lexString() *Token {
	l.advance()

	n := 0

	for l.peekAhead(n) != '"' {
		n += 1
	}

	token := l.newToken(StringToken, l.substr(n))

	l.advance()

	return token
}

func (l *Lexer) lexNextToken() *Token {
	for l.skipWhitespace() {
	}

	c := l.peek()

	if c == 0 {
		l.done = true
		return l.newToken(EOF, "")
	}

	if c == ',' {
		l.advance()
		return l.newToken(CommaToken, "")
	}

	if c == '\n' {
		l.advance()

		tok := l.newToken(NewlineToken, "")

		l.line += 1

		return tok
	}

	if c == ':' {
		l.advance()
		return l.newToken(ColonToken, "")
	}

	if c == '(' {
		l.advance()
		return l.newToken(LParenToken, "")
	}

	if c == ')' {
		l.advance()
		return l.newToken(RParenToken, "")
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

	return l.newToken(IllegalToken, string(l.advance()))
}

func (l *Lexer) NextToken() *Token {
	if l.done {
		return l.lastToken
	}

lex:
	tok := l.lexNextToken()

	if (l.lastToken == nil && tok.Type == NewlineToken) || (l.lastToken != nil && l.lastToken.Type == NewlineToken && tok.Type == NewlineToken) {
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
