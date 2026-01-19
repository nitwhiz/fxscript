package fx

import (
	"io"

	"github.com/nitwhiz/ring-buffer"
)

type TokenIterator struct {
	src     TokenSource
	drained bool
	buf     *ring.Buffer[*Token]
	bufSize int

	prev *TokenIterator
}

func NewTokenIterator(source TokenSource, bufSize int) *TokenIterator {
	return &TokenIterator{
		src:     source,
		drained: false,
		buf:     ring.NewBuffer[*Token](bufSize),
		bufSize: bufSize,

		prev: nil,
	}
}

func (i *TokenIterator) fillBuffer() (err error) {
	if i.drained {
		return
	}

	prefetch := tokenPrefetch - i.buf.Len()

	if prefetch <= 0 {
		return
	}

	var tok *Token

	for range prefetch {
		tok, err = i.src.NextToken()

		if err != nil {
			if err == io.EOF {
				err = nil
			}

			return
		}

		if tok.Type == EOF {
			i.drained = true
			return
		}

		if err = i.buf.WriteOne(tok); err != nil {
			return
		}
	}

	return
}

func (i *TokenIterator) Insert(src TokenSource) {
	i.prev = &TokenIterator{
		src:     i.src,
		buf:     i.buf,
		bufSize: i.bufSize,
		drained: i.drained,
		prev:    i.prev,
	}

	i.src = src
	i.buf = ring.NewBuffer[*Token](i.bufSize)
	i.drained = false
}

func (i *TokenIterator) Peek(n int) (tok *Token, err error) {
	if err = i.fillBuffer(); err != nil {
		return
	}

	unread := i.buf.Len()

	if unread <= n {
		if i.prev == nil {
			tok = eofToken
			return
		}

		return i.prev.Peek(n - unread)
	}

	tok, err = i.buf.Peek(n)

	if err == io.EOF {
		tok = eofToken
		err = nil
	}

	return
}

func (i *TokenIterator) NextToken() (tok *Token, err error) {
	if i.src == nil {
		tok = eofToken
		return
	}

	if err = i.fillBuffer(); err != nil {
		return
	}

	tok, err = i.buf.ReadOne()

	if err != nil {
		if err == io.EOF {
			tok = eofToken
			err = nil

			if i.prev == nil {
				i.src = nil
				i.buf = nil
				i.drained = true
			} else {
				*i = *i.prev
			}

			return i.NextToken()
		}

		return
	}

	return
}

type TokenSlice struct {
	tokens []*Token
	offset int
}

func newTokenSlice(tokens []*Token) *TokenSlice {
	return &TokenSlice{tokens: tokens}
}

func (s *TokenSlice) NextToken() (tok *Token, err error) {
	if s.offset >= len(s.tokens) {
		tok = eofToken
		return
	}

	tok = s.tokens[s.offset]

	s.offset++

	return
}
