package fx

import (
	"io"
	"io/fs"
	"path"
)

type ParserFS struct {
	fs.FS
	path string
}

func (p *ParserFS) Open(name string) (fs.File, error) {
	return p.FS.Open(path.Join(p.path, name))
}

func (p *ParserFS) WithPath(path string) *ParserFS {
	return &ParserFS{p.FS, path}
}

func NewParserFS(fs fs.FS) *ParserFS {
	return &ParserFS{fs, "."}
}

func (p *Parser) parseFile(fileName string) (err error) {
	sfs := p.fs
	srcFilename := p.src.Filename()

	if srcFilename == "" {
		srcFilename = ""
	}

	dirPath := path.Join(path.Dir(srcFilename), path.Dir(fileName))

	if dirPath != "" {
		sfs = p.fs.WithPath(dirPath)
	}

	fullPath := path.Join(sfs.path, path.Base(fileName))

	if _, ok := p.includedFiles[fullPath]; ok {
		return
	}

	p.includedFiles[fullPath] = true

	f, err := p.fs.Open(fullPath)

	if err != nil {
		return
	}

	defer f.Close()

	var scriptData []byte

	if scriptData, err = io.ReadAll(f); err != nil {
		return
	}

	p.src.Insert("", NewLexer(scriptData, fullPath))

	return
}

func LoadScript(scriptData []byte, cfg *ParserConfig) (script *Script, err error) {
	return NewParser(NewLexer(scriptData, ""), cfg).Parse()
}

func LoadFile(fileName string, cfg *ParserConfig) (script *Script, err error) {
	f, err := cfg.FS.Open(fileName)

	if err != nil {
		return
	}

	bs, err := io.ReadAll(f)

	if err != nil {
		return
	}

	defer f.Close()

	return NewParser(NewLexer(bs, fileName), cfg).Parse()
}
