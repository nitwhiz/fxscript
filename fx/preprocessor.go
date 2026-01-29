package fx

import (
	"io"
	"io/fs"
	"path"
)

type LookupFn func(value string) []byte

type Preprocessor struct {
	fs       fs.FS
	lookupFn LookupFn
}

func NewPreprocessor(cfg *ParserConfig) *Preprocessor {
	return &Preprocessor{cfg.FS, cfg.LookupFn}
}

func (p *Preprocessor) WithFS(fs fs.FS) *Preprocessor {
	return &Preprocessor{fs, p.lookupFn}
}

type PreprocessorDirective struct {
	Directive string
	Argument  string
	Begin     int
	Len       int
}

func concat(scriptData, insert []byte, directive *PreprocessorDirective) (resultScriptData []byte, size int, next int) {
	size = len(insert)

	resultScriptData = make([]byte, 0, len(scriptData)+size-directive.Len)

	resultScriptData = append(resultScriptData, scriptData[:directive.Begin]...)

	if size > 0 {
		resultScriptData = append(resultScriptData, insert...)
	}

	resultScriptData = append(resultScriptData, scriptData[(directive.Begin+directive.Len):]...)

	next = directive.Begin + size

	return
}

func (p *Preprocessor) include(scriptData []byte, directive *PreprocessorDirective) (resultScriptData []byte, next int, size int, err error) {
	var bs []byte

	if p.fs != nil {
		if bs, err = p.loadScriptFile(directive.Argument); err != nil {
			return
		}
	}

	resultScriptData, size, next = concat(scriptData, bs, directive)

	return
}

func (p *Preprocessor) constLookup(scriptData []byte, directive *PreprocessorDirective) (resultScriptData []byte, next int, size int, err error) {
	if p.lookupFn == nil {
		err = &SyntaxError{&MissingLookupFnError{directive.Directive}}
		return
	}

	var value []byte

	value = append(value, "const "...)
	value = append(value, p.lookupFn(directive.Argument)...)

	resultScriptData, size, next = concat(scriptData, value, directive)

	return
}

func (p *Preprocessor) handleDirective(scriptData []byte, directive *PreprocessorDirective) (resultScriptData []byte, next int, size int, err error) {
	switch directive.Directive {
	case "@include":
		if resultScriptData, next, size, err = p.include(scriptData, directive); err != nil {
			return
		}
		break
	case "@const":
		if resultScriptData, next, size, err = p.constLookup(scriptData, directive); err != nil {
			return
		}
		break
	default:
		err = &SyntaxError{&UnknownPreprocessorDirectiveError{directive.Directive}}
	}

	return
}

func (p *Preprocessor) process(scriptData []byte) (resultScriptData []byte, err error) {
	l := len(scriptData)

	resultScriptData = scriptData

	for i := 0; i < l; i++ {
		if resultScriptData[i] == '@' {
			directiveBegin := i

			var directiveValue string
			var argument string

			for i < l && resultScriptData[i] != ' ' && resultScriptData[i] != '\n' {
				directiveValue += string(resultScriptData[i])
				i++
			}

			if resultScriptData[i] != '\n' {
				i++

				for i < l && resultScriptData[i] != '\n' {
					argument += string(resultScriptData[i])
					i++
				}
			}

			var next int
			var size int

			directive := PreprocessorDirective{
				Directive: directiveValue,
				Argument:  argument,
				Begin:     directiveBegin,
				Len:       i - directiveBegin,
			}

			if resultScriptData, next, size, err = p.handleDirective(resultScriptData, &directive); err != nil {
				return
			}

			i = next
			l += size - directive.Len
		}
	}

	return
}

func (p *Preprocessor) loadScriptFile(fileName string) (scriptData []byte, err error) {
	sfs := p.fs

	dir := path.Dir(fileName)

	if dir != "" {
		if sfs, err = fs.Sub(sfs, dir); err != nil {
			return
		}
	}

	f, err := sfs.Open(path.Base(fileName))

	if err != nil {
		return
	}

	defer f.Close()

	if scriptData, err = io.ReadAll(f); err != nil {
		return
	}

	return p.WithFS(sfs).process(scriptData)
}

func (p *Preprocessor) LoadScript(source []byte, cfg *ParserConfig) (script *Script, err error) {
	if source, err = p.process(source); err != nil {
		return
	}

	lexer := NewLexer(source)
	parser := NewParser(lexer, cfg)

	script, err = parser.Parse()

	return
}

func (p *Preprocessor) Load(entryFileName string, cfg *ParserConfig) (script *Script, err error) {
	var source []byte

	if source, err = p.loadScriptFile(entryFileName); err != nil {
		return
	}

	return p.LoadScript(source, cfg)
}
