package buffer // import "github.com/tdewolff/buffer"

import "io"

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

// Err returns the error.
func (z *Shifter) Err() error {
	if z.eof && z.end < len(z.buf) {
		return nil
	}
	return z.err
}

// IsEOF returns true when it has encountered EOF and thus loaded the last buffer in memory.
func (z *Shifter) IsEOF() bool {
	return z.eof
}

// Peek returns the ith byte and possibly does an allocation.
func (z *Shifter) Peek(i int) byte {
	end := z.end + i
	if end >= len(z.buf) {
		if z.err != nil {
			return 0
		}

		// reallocate a new buffer (possibly larger)
		c := cap(z.buf)
		d := len(z.buf) - z.pos
		var buf []byte
		if 2*d > c {
			buf = make([]byte, d, 2*c)
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
			return 0
		}
	}
	return z.buf[end]
}

// PeekRune returns the rune of the ith byte.
func (z *Shifter) PeekRune(i int) rune {
	// from unicode/utf8
	c := z.Peek(i)
	if c < 0xC0 {
		return rune(c)
	} else if c < 0xE0 {
		return rune(c&0x1F)<<6 | rune(z.Peek(i+1)&0x3F)
	} else if c < 0xF0 {
		return rune(c&0x0F)<<12 | rune(z.Peek(i+1)&0x3F)<<6 | rune(z.Peek(i+2)&0x3F)
	} else {
		return rune(c&0x07)<<18 | rune(z.Peek(i+1)&0x3F)<<12 | rune(z.Peek(i+2)&0x3F)<<6 | rune(z.Peek(i+3)&0x3F)
	}
}

// Move advances the 0 position of Peek.
func (z *Shifter) Move(n int) {
	z.end += n
}

// MoveTo sets the 0 position of Peek.
func (z *Shifter) MoveTo(n int) {
	z.end = z.pos + n
}

// Pos returns the 0 position of Peek.
func (z *Shifter) Pos() int {
	return z.end - z.pos
}

// Bytes returns the bytes of the current selection.
func (z *Shifter) Bytes() []byte {
	return z.buf[z.pos:z.end]
}

// Shift returns the bytes of the current selection and collapses the position.
func (z *Shifter) Shift() []byte {
	b := z.buf[z.pos:z.end]
	z.pos = z.end
	return b
}

// Skip collapses the position.
func (z *Shifter) Skip() {
	z.pos = z.end
}
