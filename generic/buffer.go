package generic

type T interface{}

type Buffer struct {
	buf []T
	pos int

	Read func([]T) int
}

// Peek returns the ith element and possibly does an allocation.
// Peeking past an error will panic.
func (z *Buffer) Peek(i int) *T {
	end := z.pos + i
	if end >= len(z.buf) {
		c := cap(z.buf)
		d := len(z.buf) - z.pos
		var buf []T
		if 2*d > c {
			buf = make([]T, d, 2*c)
		} else {
			buf = z.buf[:d]
		}
		copy(buf, z.buf[z.pos:])

		n := z.Read(buf[d:cap(buf)])
		end -= z.pos
		z.pos, z.buf = 0, buf[:d+n]
	}
	return &z.buf[end]
}

// Shift returns the first element and advances position.
func (z *Buffer) Shift() *T {
	t := z.Peek(0)
	z.pos++
	return t
}
