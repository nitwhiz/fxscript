package fx

import "fmt"

// todo: maybe strip this from release, too?

type SourceFile struct {
	Name string
}

type SourceInfo struct {
	Name   string
	File   *SourceFile
	Line   int
	Column int
	Parent *SourceInfo
}

func (s *SourceInfo) String() string {
	var fName string

	if s.File == nil {
		fName = "<script>"
	} else {
		fName = s.File.Name
	}

	return fmt.Sprintf("%s:%d:%d", fName, s.Line, s.Column)
}
