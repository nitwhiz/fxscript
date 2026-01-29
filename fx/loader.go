package fx

func LoadFS(entryFileName string, cfg *ParserConfig) (script *Script, err error) {
	return NewPreprocessor(cfg).Load(entryFileName, cfg)
}

func LoadScript(bs []byte, cfg *ParserConfig) (script *Script, err error) {
	return NewPreprocessor(cfg).LoadScript(bs, cfg)
}
