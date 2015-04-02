package buffer // import "github.com/tdewolff/buffer"

import "io"

// Reader implements a reader over a byte slice.
type Reader struct {
	buf []byte
	pos int
}

// NewReader returns a new Reader for a given byte slice.
func NewReader(buf []byte) *Reader {
	return &Reader{
		buf: buf,
	}
}

// Read reads bytes into the given byte slice and returns the number of bytes read and an error if occurred.
func (r *Reader) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	if r.pos >= len(r.buf) {
		return 0, io.EOF
	}
	n := copy(b, r.buf[r.pos:])
	r.pos += n
	return n, nil
}

// Bytes returns the underlying byte slice.
func (r *Reader) Bytes() []byte {
	return r.buf
}
