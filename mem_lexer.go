package buffer // import "github.com/tdewolff/buffer"

import (
	"io"
	"io/ioutil"
)

// MemLexer is a buffered reader that allows peeking forward and shifting, taking an io.Reader.
// It keeps data in-memory until Free, taking a byte length, is called to move beyond the data.
type MemLexer struct {
	buf       []byte
	pos       int // index in buf
	start     int // index in buf
	prevStart int
}

// NewMemLexer returns a new MemLexer for a given io.Reader with a 4kB estimated buffer size.
// If the io.Reader implements Bytes, that buffer is used instead.
func NewMemLexer(r io.Reader) *MemLexer {
	var b []byte

	// if reader has the bytes in memory already, use that instead
	if buffer, ok := r.(interface {
		Bytes() []byte
	}); ok {
		b = buffer.Bytes()
	} else {
		var err error
		b, err = ioutil.ReadAll(r)
		if err != nil {
		}
	}

	// append NULL to buffer if it isn't already there
	if len(b) > 0 && b[len(b)-1] != 0 {
		b = append(b, 0)
	}
	return &MemLexer{
		buf: b,
	}
}

// Err returns the error returned from io.Reader. It may still return valid bytes for a while though.
func (z *MemLexer) Err() error {
	if z.pos >= len(z.buf)-1 {
		return io.EOF
	}
	return nil
}

// Free frees up bytes of length n from previously shifted tokens.
// Each call to Shift should at one point be followed by a call to Free with a length returned by ShiftLen.
func (z *MemLexer) Free(n int) {
}

// Peek returns the ith byte relative to the end position and possibly does an allocation.
// Peek returns zero when an error has occurred, Err returns the error.
func (z *MemLexer) Peek(pos int) byte {
	pos += z.pos
	return z.buf[pos]
}

// PeekRune returns the rune and rune length of the ith byte relative to the end position.
func (z *MemLexer) PeekRune(pos int) (rune, int) {
	// from unicode/utf8
	c := z.Peek(pos)
	if c < 0xC0 {
		return rune(c), 1
	} else if c < 0xE0 {
		return rune(c&0x1F)<<6 | rune(z.Peek(pos+1)&0x3F), 2
	} else if c < 0xF0 {
		return rune(c&0x0F)<<12 | rune(z.Peek(pos+1)&0x3F)<<6 | rune(z.Peek(pos+2)&0x3F), 3
	}
	return rune(c&0x07)<<18 | rune(z.Peek(pos+1)&0x3F)<<12 | rune(z.Peek(pos+2)&0x3F)<<6 | rune(z.Peek(pos+3)&0x3F), 4
}

// Move advances the position.
func (z *MemLexer) Move(n int) {
	z.pos += n
}

// Pos returns a mark to which can be rewinded.
func (z *MemLexer) Pos() int {
	return z.pos - z.start
}

// Rewind rewinds the position to the given position.
func (z *MemLexer) Rewind(pos int) {
	z.pos = z.start + pos
}

// Lexeme returns the bytes of the current selection.
func (z *MemLexer) Lexeme() []byte {
	return z.buf[z.start:z.pos]
}

// Skip collapses the position to the end of the selection.
func (z *MemLexer) Skip() {
	z.start = z.pos
}

// Shift returns the bytes of the current selection and collapses the position to the end of the selection.
// It also returns the number of bytes we moved since the last call to Shift. This can be used in calls to Free.
func (z *MemLexer) Shift() []byte {
	b := z.buf[z.start:z.pos]
	z.start = z.pos
	return b
}
