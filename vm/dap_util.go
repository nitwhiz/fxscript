//go:build debug

package vm

import (
	"fmt"
	"iter"
	"maps"
	"strconv"
	"strings"

	"github.com/google/go-dap"
	"github.com/nitwhiz/fxscript/fx"
)

const (
	scopeIdentifiers = 1
	scopeDefines     = 2
	scopeVariables   = 3
	scopeLabels      = 4
)

type FileLineList struct {
	elements map[string]*fx.SourceInfo
}

func newFileLineList() *FileLineList {
	return (&FileLineList{}).ClearAll()
}

func (l *FileLineList) String() string {
	return fmt.Sprintf("%v", l.elements)
}

func (l *FileLineList) canonical(file string, line int) string {
	return file + ":" + strconv.Itoa(line)
}

func (l *FileLineList) GetAt(file string, line int) (s *fx.SourceInfo, ok bool) {
	if s, ok = l.elements[l.canonical(file, line)]; ok {
		return
	}

	return nil, false
}

func (l *FileLineList) Add(s *fx.SourceInfo) {
	l.elements[l.canonical(s.File.Name, s.Line)] = s
}

func (l *FileLineList) AddAt(file string, line int, s *fx.SourceInfo) {
	l.elements[l.canonical(file, line)] = s
}

func (l *FileLineList) Has(s *fx.SourceInfo) (ok bool) {
	if s == nil {
		return
	}

	_, ok = l.elements[l.canonical(s.File.Name, s.Line)]
	return
}

func (l *FileLineList) HasAt(file string, line int) (ok bool) {
	_, ok = l.elements[l.canonical(file, line)]
	return
}

func (l *FileLineList) Delete(file string) {
	for k := range l.elements {
		if strings.HasPrefix(k, file+":") {
			delete(l.elements, k)
		}
	}
}

func (l *FileLineList) All() iter.Seq2[string, *fx.SourceInfo] {
	return maps.All(l.elements)
}

func (l *FileLineList) ClearAll() *FileLineList {
	l.elements = make(map[string]*fx.SourceInfo)
	return l
}

type StackFrames struct {
	frames []dap.StackFrame
}

func newStackFrames() *StackFrames {
	return &StackFrames{
		frames: make([]dap.StackFrame, 0, 8),
	}
}

func (s *StackFrames) Current() *dap.StackFrame {
	if len(s.frames) == 0 {
		return nil
	}

	return &s.frames[0]
}

func (s *StackFrames) Update(cmdSrc *fx.SourceInfo) {
	if len(s.frames) > 0 {

		s.frames[0].Line = cmdSrc.Line
		s.frames[0].Column = cmdSrc.Column
		s.frames[0].Source.Path = cmdSrc.File.Name
	}
}

func (s *StackFrames) Set(frames ...dap.StackFrame) {
	s.frames = frames
}

func (s *StackFrames) Add(frame dap.StackFrame) {
	s.frames = append(
		[]dap.StackFrame{
			frame,
		},
		s.frames...,
	)
}

func (s *StackFrames) AddParent(src *fx.SourceInfo) {
	if src.Parent != nil {
		s.Add(dap.StackFrame{
			Name:   "MACRO " + src.Name,
			Line:   src.Parent.Line,
			Column: src.Parent.Column,
			Source: &dap.Source{
				Path: src.Parent.File.Name,
			},
		})
	}
}

func (s *StackFrames) RemoveMacros() {
	n := 0

	for _, f := range s.frames {
		if !strings.HasPrefix(f.Name, "MACRO ") {
			break
		}
		n++
	}

	s.Remove(n)
}

func (s *StackFrames) Remove(n int) {
	if n == 0 {
		return
	}

	if n >= len(s.frames) {
		s.Set()
		return
	}

	s.frames = s.frames[n:]
}

func (s *StackFrames) Ret() {
	// remove all until there is a call

	start := 0

	for i := start + 1; i < len(s.frames); i++ {
		if strings.HasPrefix(s.frames[i].Name, "CALL ") {
			start = i + 1
			break
		}
	}

	if len(s.frames) > 0 {
		s.frames = s.frames[start:len(s.frames)]
	}
}

func (s *StackFrames) All() iter.Seq[dap.StackFrame] {
	return func(yield func(v dap.StackFrame) bool) {
		for _, frame := range s.frames {
			if !yield(frame) {
				return
			}
		}
	}
}
