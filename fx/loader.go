package fx

import (
	"io"
	"io/fs"
	"path"
)

func loadScriptSource(fs fs.FS, entryFileName string) (scriptData []byte, err error) {
	dir := path.Dir(entryFileName)

	f, err := fs.Open(entryFileName)

	if err != nil {
		return
	}

	defer f.Close()

	if scriptData, err = io.ReadAll(f); err != nil {
		return
	}

	l := len(scriptData)

	for i := 0; i < l; i++ {
		if scriptData[i] == '@' {
			directiveBegin := i

			var directiveValue string
			var argument string

			for i < l && scriptData[i] != ' ' && scriptData[i] != '\n' {
				directiveValue += string(scriptData[i])
				i++
			}

			if scriptData[i] != '\n' {
				i++

				for i < l && scriptData[i] != '\n' {
					argument += string(scriptData[i])
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
				Dir:       dir,
				FS:        fs,
			}

			if scriptData, next, size, err = prepHandleDirective(scriptData, &directive); err != nil {
				return
			}

			i = next
			l += size - directive.Len
		}
	}

	return
}

func LoadScript(source []byte, cfg *ParserConfig) (script *Script, err error) {
	l := NewLexer(source)
	p := NewParser(l, cfg)

	script, err = p.Parse()

	return
}

func LoadFS(fs fs.FS, entryFileName string, cfg *ParserConfig) (script *Script, err error) {
	var source []byte

	if source, err = loadScriptSource(fs, entryFileName); err != nil {
		return
	}

	return LoadScript(source, cfg)
}
