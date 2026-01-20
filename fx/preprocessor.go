package fx

import (
	"io/fs"
	"path"
)

type PreprocessorDirective struct {
	Directive string
	Argument  string
	Begin     int
	Len       int
	Dir       string
	FS        fs.FS
}

func prepReplace(scriptData []byte, directive *PreprocessorDirective) (resultScriptData []byte, next int, size int, err error) {
	var bs []byte

	if bs, err = loadScriptSource(directive.FS, path.Join(directive.Dir, directive.Argument)); err != nil {
		return
	}

	size = len(bs)
	next = directive.Begin + size

	resultScriptData = make([]byte, 0, len(scriptData)+size-directive.Len)

	resultScriptData = append(resultScriptData, scriptData[:directive.Begin]...)
	resultScriptData = append(resultScriptData, bs...)
	resultScriptData = append(resultScriptData, scriptData[(directive.Begin+directive.Len):]...)

	return
}

func prepHandleDirective(scriptData []byte, directive *PreprocessorDirective) (resultScriptData []byte, next int, size int, err error) {
	switch directive.Directive {
	case "@include":
		if resultScriptData, next, size, err = prepReplace(scriptData, directive); err != nil {
			return
		}

		break
	default:
		err = &SyntaxError{&UnknownPreprocessorDirectiveError{directive.Directive}}
	}

	return
}
