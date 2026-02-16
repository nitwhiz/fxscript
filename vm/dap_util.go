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

type FileLineList[T any] struct {
	elements map[string]T
}

func newFileLineList[T any]() *FileLineList[T] {
	return (&FileLineList[T]{}).ClearAll()
}

func (l *FileLineList[T]) String() string {
	return fmt.Sprintf("%v", l.elements)
}

func (l *FileLineList[T]) canonical(file string, line int) string {
	return file + ":" + strconv.Itoa(line)
}

func (l *FileLineList[T]) Add(file string, line int, value T) {
	l.elements[l.canonical(file, line)] = value
}

func (l *FileLineList[T]) Has(file string, line int) (ok bool) {
	_, ok = l.elements[l.canonical(file, line)]
	return
}

func (l *FileLineList[T]) Delete(file string) {
	for k := range l.elements {
		if strings.HasPrefix(k, file+":") {
			delete(l.elements, k)
		}
	}
}

func (l *FileLineList[T]) All() iter.Seq2[string, T] {
	return maps.All(l.elements)
}

func (l *FileLineList[T]) ClearAll() *FileLineList[T] {
	l.elements = make(map[string]T)
	return l
}

type BreakpointList struct {
	*FileLineList[struct{}]
}

func newBreakpointList() *BreakpointList {
	return &BreakpointList{newFileLineList[struct{}]()}
}

func (b *BreakpointList) Add(file string, line int) {
	b.FileLineList.Add(file, line, struct{}{})
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

func (s *StackFrames) IsCurrent(src *fx.SourceInfo) bool {
	if src == nil {
		return false
	}

	if len(s.frames) == 0 {
		return false
	}

	return s.frames[0].Column == src.Column && s.frames[0].Line == src.Line && s.frames[0].Source.Path == src.File.Name
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

func (s *StackFrames) AddParents(src *fx.SourceInfo) {
	if src.Parent != nil {
		s.AddParents(src.Parent)

		s.Update(src.Parent)

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

func (s *StackFrames) RemoveParents(src *fx.SourceInfo) {
	n := 0

	for src.Parent != nil {
		n++
		src = src.Parent
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

func (s *StackFrames) AllReverse() iter.Seq[dap.StackFrame] {
	return func(yield func(v dap.StackFrame) bool) {
		for i := len(s.frames) - 1; i >= 0; i-- {
			if !yield(s.frames[i]) {
				return
			}
		}
	}
}
