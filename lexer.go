package buffer // import "github.com/tdewolff/buffer"

import "io"

type block struct {
	buf    []byte
	next   int // index in pool plus one
	active bool
}

type BufferPool struct {
	pool []block
	head int // index in pool plus one
	tail int // index in pool plus one

	pos int // byte pos in tail
}

func (z *BufferPool) swap(oldBuf []byte, size int) []byte {
	// find new buffer that can be reused
	swap := -1
	for i, _ := range z.pool {
		if !z.pool[i].active && size <= cap(z.pool[i].buf) {
			swap = i
			break
		}
	}
	if swap == -1 { // no free buffer found for reuse
		if z.tail == 0 && z.pos >= len(oldBuf) && size <= cap(oldBuf) { // but we can reuse the current buffer!
			z.pos -= len(oldBuf)
			return oldBuf[:0]
		} else { // allocate new
			z.pool = append(z.pool, block{make([]byte, 0, size), 0, true})
			swap = len(z.pool) - 1
		}
	}

	newBuf := z.pool[swap].buf

	// put current buffer into pool
	z.pool[swap] = block{oldBuf, 0, true}
	if z.head != 0 {
		z.pool[z.head-1].next = swap + 1
	}
	z.head = swap + 1
	if z.tail == 0 {
		z.tail = swap + 1
	}

	return newBuf[:0]
}

func (z *BufferPool) free(n int) {
	z.pos += n
	// move the tail over to next buffers
	for z.tail != 0 && z.pos >= len(z.pool[z.tail-1].buf) {
		z.pos -= len(z.pool[z.tail-1].buf)
		newTail := z.pool[z.tail-1].next
		z.pool[z.tail-1].active = false // after this, any thread may pick up the inactive buffer, so it can't be used anymore
		z.tail = newTail
	}
	if z.tail == 0 {
		z.head = 0
	}
}

// Lexer is a buffered reader that allows peeking forward and shifting, taking an io.Reader.
// It keeps data in-memory until Free, taking a byte length, is called to move beyond the data.
type Lexer struct {
	r   io.Reader
	err error

	pool BufferPool

	buf []byte
	pos int // index in buf
	end int // index in buf
}

// NewLexer returns a new Lexer for a given io.Reader with a 4kB estimated buffer size.
// If the io.Reader implements Bytes, that buffer is used instead.
func NewLexer(r io.Reader) *Lexer {
	return NewLexerSize(r, defaultBufSize)
}

// NewLexerSize returns a new Lexer for a given io.Reader and estimated required buffer size.
// If the io.Reader implements Bytes, that buffer is used instead.
func NewLexerSize(r io.Reader, size int) *Lexer {
	// if reader has the bytes in memory already, use that instead
	if buffer, ok := r.(interface {
		Bytes() []byte
	}); ok {
		return &Lexer{
			err: io.EOF,
			buf: buffer.Bytes(),
		}
	}
	z := &Lexer{
		r:   r,
		buf: make([]byte, 0, size),
	}
	z.Peek(0)
	return z
}

func (z *Lexer) read(end int) byte {
	if z.err != nil {
		return 0
	}

	// get new buffer
	size := cap(z.buf)
	d := len(z.buf) - z.pos
	if 2*d > c { // if the token is larger than half the buffer, increase buffer size
		size = 2*size + d
	}
	buf := z.pool.swap(z.buf[:z.pos], size)
	copy(buf[:d], z.buf[z.pos:]) // copy the left-overs (unfinished token) from the old buffer

	// read in new data for the rest of the buffer
	var n int
	n, z.err = z.r.Read(buf[d:cap(buf)])
	end -= z.pos
	z.end -= z.pos
	z.pos, z.buf = 0, buf[:d+n]
	if n == 0 {
		if z.err == nil {
			z.err = io.EOF
		}
		return 0
	}
	return z.buf[end]
}

func (z *Lexer) Free(n int) {
	z.pool.free(n)
}

// Err returns the error returned from io.Reader. It may still return valid bytes for a while though.
func (z *Lexer) Err() error {
	if z.err == io.EOF && z.end < len(z.buf) {
		return nil
	}
	return z.err
}

// Peek returns the ith byte relative to the end position and possibly does an allocation.
// Peek returns zero when an error has occurred, Err returns the error.
func (z *Lexer) Peek(end int) byte {
	end += z.end
	if end >= len(z.buf) {
		return z.read(end)
	}
	return z.buf[end]
}

// PeekRune returns the rune and rune length of the ith byte relative to the end position.
func (z *Lexer) PeekRune(i int) (rune, int) {
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
func (z *Lexer) Move(n int) {
	z.end += n
}

// MoveTo rewinds the position to the given mark.
func (z *Lexer) MoveTo(n int) {
	z.end = z.pos + n
}

// Pos returns a mark to which can be rewinded.
func (z *Lexer) Pos() int {
	return z.end - z.pos
}

// Bytes returns the bytes of the current selection.
func (z *Lexer) Bytes() []byte {
	return z.buf[z.pos:z.end]
}

// Shift returns the bytes of the current selection and collapses the position to the end.
func (z *Lexer) Shift() []byte {
	b := z.buf[z.pos:z.end]
	z.pos = z.end
	return b
}

// Skip collapses the position to the end.
func (z *Lexer) Skip() {
	z.pos = z.end
}
