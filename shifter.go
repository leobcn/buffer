/*
Package buffer contains buffer and wrapper types for byte slices. It is useful for writing lexers or other high-performance byte slice handling.

The `Reader` and `Writer` types implement the `io.Reader` and `io.Writer` respectively and provide a thinner and faster interface than `bytes.Buffer`.
The `Shifter` type is useful for building lexers because it keeps track of the start and end position of a byte selection, and shifts the bytes whenever a valid token is found.
*/
package buffer // import "github.com/tdewolff/buffer"

import "io"

// MinBuf specified the initial length of the internal shifter buffer.
var MinBuf = 4096

// Shifter is a buffered reader that allows peeking forward and shifting, taking an io.Reader.
type Shifter struct {
	r   io.Reader
	err error
	eof bool

	buf []byte
	pos int
	end int
}

// NewShifter returns a new Shifter for a given io.Reader.
func NewShifter(r io.Reader) *Shifter {
	// If reader has the bytes in memory already, use that instead!
	if buffer, ok := r.(interface {
		Bytes() []byte
	}); ok {
		return &Shifter{
			err: io.EOF,
			eof: true,
			buf: buffer.Bytes(),
		}
	}
	z := &Shifter{
		r:   r,
		buf: make([]byte, 0, MinBuf),
	}
	z.Peek(0)
	return z
}

// Err returns the error returned from io.Reader. It may still return valid bytes for a while though.
func (z *Shifter) Err() error {
	if z.eof && z.end < len(z.buf) {
		return nil
	}
	return z.err
}

// IsEOF returns true when it has encountered EOF meaning that it has loaded the last data in memory (ie. previously returned byte slice will not be overwritten by Peek).
// Calling IsEOF is faster than checking Err() == io.EOF.
func (z *Shifter) IsEOF() bool {
	return z.eof
}

func (z *Shifter) read(end int) byte {
	if z.err != nil {
		return 0
	}

	// reallocate a new buffer (possibly larger)
	c := cap(z.buf)
	d := len(z.buf) - z.pos
	var buf []byte
	if 2*d > c {
		buf = make([]byte, d, 2*c+end-z.pos)
	} else {
		buf = z.buf[:d]
	}
	copy(buf, z.buf[z.pos:])

	// read in to fill the buffer till capacity
	var n int
	n, z.err = z.r.Read(buf[d:cap(buf)])
	z.eof = (z.err == io.EOF)
	end -= z.pos
	z.end -= z.pos
	z.pos, z.buf = 0, buf[:d+n]
	if n == 0 {
		if z.err == nil {
			z.err = io.EOF
			z.eof = true
		}
		return 0
	}
	return z.buf[end]
}

// Peek returns the ith byte relative to the end position and possibly does an allocation. Calling Peek may invalidate previous returned byte slices by Bytes or Shift, unless IsEOF returns true.
// Peek returns zero when an error has occurred, Err returns the error.
func (z *Shifter) Peek(end int) byte {
	end += z.end
	if end >= len(z.buf) {
		return z.read(end)
	}
	return z.buf[end]
}

// PeekRune returns the rune and rune length of the ith byte relative to the end position.
func (z *Shifter) PeekRune(i int) (rune, int) {
	// from unicode/utf8
	c := z.Peek(i)
	if c < 0xC0 {
		return rune(c), 1
	} else if c < 0xE0 {
		return rune(c&0x1F)<<6 | rune(z.Peek(i+1)&0x3F), 2
	} else if c < 0xF0 {
		return rune(c&0x0F)<<12 | rune(z.Peek(i+1)&0x3F)<<6 | rune(z.Peek(i+2)&0x3F), 3
	} else {
		return rune(c&0x07)<<18 | rune(z.Peek(i+1)&0x3F)<<12 | rune(z.Peek(i+2)&0x3F)<<6 | rune(z.Peek(i+3)&0x3F), 4
	}
}

// Move advances the end position.
func (z *Shifter) Move(n int) {
	z.end += n
}

// MoveTo sets the end position.
func (z *Shifter) MoveTo(n int) {
	z.end = z.pos + n
}

// Pos returns the end position.
func (z *Shifter) Pos() int {
	return z.end - z.pos
}

// Bytes returns the bytes of the current selection.
func (z *Shifter) Bytes() []byte {
	return z.buf[z.pos:z.end]
}

// Shift returns the bytes of the current selection and collapses the position to the end.
func (z *Shifter) Shift() []byte {
	b := z.buf[z.pos:z.end]
	z.pos = z.end
	return b
}

// Skip collapses the position to the end.
func (z *Shifter) Skip() {
	z.pos = z.end
}
