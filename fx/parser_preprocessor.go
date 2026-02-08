package fx

import (
	"bytes"
	"strings"
)

type LookupFn func(value string) ([]byte, error)

func (p *Parser) prepInclude(fileName string) error {
	return p.parseFile(fileName)
}

func (p *Parser) prepDefLookup(value string) (err error) {
	if p.lookupFn == nil {
		err = &MissingLookupFnError{value}
		return
	}

	lookupValue, err := p.lookupFn(value)

	if err != nil {
		return
	}

	result := &bytes.Buffer{}

	if _, err = result.WriteString("def "); err != nil {
		return
	}

	if _, err = result.Write(lookupValue); err != nil {
		return
	}

	p.src.Insert("", NewLexer(result.Bytes(), ""))

	return
}

func (p *Parser) parsePreprocessorDirective() (err error) {
	tok, err := p.advance()

	if err != nil {
		return
	}

	segments := strings.SplitN(tok.Value, " ", 2)

	if len(segments) == 0 {
		err = &SyntaxError{tok.SourceInfo, &InvalidPreprocessorValueError{"", tok.Value}}
		return
	}

	switch segments[0] {
	case "include":
		if len(segments) != 2 {
			err = &SyntaxError{tok.SourceInfo, &InvalidPreprocessorValueError{"include", tok.Value}}
			return
		}

		err = p.prepInclude(segments[1])

		return
	case "def":
		if len(segments) != 2 {
			err = &SyntaxError{tok.SourceInfo, &InvalidPreprocessorValueError{"def", tok.Value}}
			return
		}

		err = p.prepDefLookup(segments[1])

		return
	}

	return
}
