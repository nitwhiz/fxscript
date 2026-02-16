package fx

import (
	"os"
	"path"
)

type ParserFS interface {
	ReadFile(name string) (string, []byte, error)
	WithBasePath(path string) ParserFS
	FilePath(name string) string
}

type ParserOsFS struct {
	basePath string
}

func (p *ParserOsFS) ReadFile(name string) (abs string, bs []byte, err error) {
	if path.IsAbs(name) {
		abs = name
	} else {
		abs = path.Join(p.basePath, name)
	}

	bs, err = os.ReadFile(abs)

	return
}

func (p *ParserOsFS) WithBasePath(basePath string) ParserFS {
	if !path.IsAbs(basePath) {
		basePath = path.Join(p.basePath, basePath)
	}

	return NewParserOsFS(basePath)
}

func (p *ParserOsFS) FilePath(name string) string {
	return path.Join(p.basePath, name)
}

func NewParserOsFS(basePath string) *ParserOsFS {
	if !path.IsAbs(basePath) {
		wd, _ := os.Getwd()
		basePath = path.Join(wd, basePath)
	}

	return &ParserOsFS{basePath}
}

func (p *Parser) parseFile(fileName string) (err error) {
	sfs := p.fs
	srcFilename := p.src.Filename()
	dirPath := path.Dir(fileName)

	if dirPath != "" && dirPath != "." {
		sfs = p.fs.WithBasePath(path.Join(path.Dir(srcFilename), dirPath))
	}

	fullPath := sfs.FilePath(path.Base(fileName))

	if _, ok := p.includedFiles[fullPath]; ok {
		return
	}

	p.includedFiles[fullPath] = true

	var scriptData []byte

	if _, scriptData, err = sfs.ReadFile(fullPath); err != nil {
		return
	}

	p.src.Insert("", NewLexer(scriptData, fullPath))

	return
}

func LoadScript(scriptData []byte, filename string, cfg *ParserConfig) (script *Script, err error) {
	return NewParser(NewLexer(scriptData, filename), cfg).Parse()
}

func LoadFile(fileName string, cfg *ParserConfig) (script *Script, err error) {
	fileName, bs, err := cfg.FS.ReadFile(fileName)
	return LoadScript(bs, fileName, cfg)
}
