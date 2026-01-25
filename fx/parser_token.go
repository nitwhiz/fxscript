package fx

import (
	"io"
	"strconv"
	"sync/atomic"

	"github.com/nitwhiz/ring-buffer"
)

type TokenIterator struct {
	prefix  string
	src     TokenSource
	drained bool
	buf     *ring.Buffer[*Token]
	bufSize int

	prev *TokenIterator

	lastInsertId *atomic.Int64
}

func NewTokenIterator(prefix string, source TokenSource, bufSize int) *TokenIterator {
	lastInsertId := atomic.Int64{}

	lastInsertId.Store(-1)

	return &TokenIterator{
		prefix:  prefix,
		src:     source,
		drained: false,
		buf:     ring.NewBuffer[*Token](bufSize),
		bufSize: bufSize,

		prev: nil,

		lastInsertId: &lastInsertId,
	}
}

func (i *TokenIterator) SetPrefix(prefix string) {
	i.prefix = prefix
}

func (i *TokenIterator) Prefixed(name string) string {
	return i.prefix + name
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
		if tok, err = i.src.NextToken(); err != nil {
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

func (i *TokenIterator) Insert(prefix string, src TokenSource) {
	i.prev = &TokenIterator{
		prefix:       i.prefix,
		src:          i.src,
		buf:          i.buf,
		bufSize:      i.bufSize,
		drained:      i.drained,
		prev:         i.prev,
		lastInsertId: i.lastInsertId,
	}

	i.prefix = prefix + "_" + strconv.Itoa(int(i.lastInsertId.Add(1)))
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

	if tok, err = i.buf.ReadOne(); err != nil {
		if err == io.EOF {
			tok = eofToken
			err = nil

			if i.prev == nil {
				i.prefix = ""
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
