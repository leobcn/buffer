package buffer // import "github.com/tdewolff/buffer"

import (
	"io"
	"io/ioutil"
)

var nullBuffer = []byte{0}

// MemLexer is a buffered reader that allows peeking forward and shifting, taking an io.Reader.
// It keeps data in-memory until Free, taking a byte length, is called to move beyond the data.
type MemLexer struct {
	buf   []byte
	pos   int // index in buf
	start int // index in buf
	err   error

	restore func()
}

func NewMemLexer(r io.Reader) *MemLexer {
	var b []byte
	if r != nil {
		if buffer, ok := r.(interface {
			Bytes() []byte
		}); ok {
			b = buffer.Bytes()
		} else {
			var err error
			// TODO: ReadALL + 1 byte, so reallocation for appending NULL is not needed
			b, err = ioutil.ReadAll(r)
			if err != nil {
				return &MemLexer{
					err: err,
				}
			}
		}
	}
	return NewMemLexerBytes(b)
}

// NewMemLexer returns a new MemLexer for a given io.Reader with a 4kB estimated buffer size.
// If the io.Reader implements Bytes, that buffer is used instead.
func NewMemLexerBytes(b []byte) *MemLexer {
	z := &MemLexer{
		buf: b,
	}

	n := len(b)
	if n == 0 {
		z.buf = nullBuffer
	} else if b[n-1] != 0 {
		// Append NULL to buffer, but try to avoid reallocation
		if cap(b) > n {
			// Overwrite next byte but restore when done
			b = b[:n+1]
			c := b[n]
			b[n] = 0

			z.buf = b
			z.restore = func() {
				b[n] = c
			}
		} else {
			z.buf = append(b, 0)
		}
	}
	return z
}

func (z *MemLexer) Restore() {
	if z.restore != nil {
		z.restore()
		z.restore = nil
	}
}

// Err returns the error returned from io.Reader. It may still return valid bytes for a while though.
func (z *MemLexer) Err() error {
	if z.err != nil {
		return z.err
	} else if z.pos >= len(z.buf)-1 {
		z.Restore()
		return io.EOF
	}
	return nil
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
