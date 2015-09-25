package buffer // import "github.com/tdewolff/buffer"

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Don't implement Bytes() to test for buffer exceeding.
type ReaderMockup struct {
	r io.Reader
}

func (r *ReaderMockup) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

////////////////////////////////////////////////////////////////

func TestShiftBuffer(t *testing.T) {
	var s = `Lorem ipsum dolor sit amet, consectetur adipiscing elit.`
	var b = NewShifter(bytes.NewBufferString(s))

	assert.Equal(t, true, b.IsEOF(), "buffer must be fully in memory")
	assert.Equal(t, 0, b.Pos(), "buffer must start at position 0")
	assert.Equal(t, byte('L'), b.Peek(0), "first character must be 'L'")
	assert.Equal(t, byte('o'), b.Peek(1), "second character must be 'o'")

	b.Move(1)
	assert.Equal(t, byte('o'), b.Peek(0), "must be 'o' at position 1")
	assert.Equal(t, byte('r'), b.Peek(1), "must be 'r' at position 1")
	b.MoveTo(6)
	assert.Equal(t, byte('i'), b.Peek(0), "must be 'i' at position 6")
	assert.Equal(t, byte('p'), b.Peek(1), "must be 'p' at position 7")

	assert.Equal(t, []byte("Lorem "), b.Bytes(), "buffered string must now read 'Lorem ' when at position 6")
	assert.Equal(t, []byte("Lorem "), b.Shift(), "shift must return the buffered string")
	assert.Equal(t, 0, b.Pos(), "after shifting position must be 0")
	assert.Equal(t, byte('i'), b.Peek(0), "must be 'i' at position 0 after shifting")
	assert.Equal(t, byte('p'), b.Peek(1), "must be 'p' at position 1 after shifting")
	assert.Nil(t, b.Err(), "error must be nil at this point")

	b.Move(len(s) - len("Lorem ") - 1)
	assert.Nil(t, b.Err(), "error must be nil just before the end of the buffer")
	b.Skip()
	assert.Equal(t, 0, b.Pos(), "after skipping position must be 0")
	b.Move(1)
	assert.Equal(t, io.EOF, b.Err(), "error must be EOF when past the buffer")
	b.Move(-1)
	assert.Nil(t, b.Err(), "error must be nil just before the end of the buffer, even when it has been past the buffer")
}

func TestShiftBufferSmall(t *testing.T) {
	MinBuf = 4
	s := `abcdefghi`
	b := NewShifter(&ReaderMockup{bytes.NewBufferString(s)})
	assert.Equal(t, byte('i'), b.Peek(8), "first character must be 'i' at position 8")
}

func TestShiftBufferRunes(t *testing.T) {
	var b = NewShifter(bytes.NewBufferString("aæ†\U00100000"))
	r, n := b.PeekRune(0)
	assert.Equal(t, 1, n, "first character must be length 1")
	assert.Equal(t, 'a', r, "first character must be rune 'a'")
	r, n = b.PeekRune(1)
	assert.Equal(t, 2, n, "second character must be length 2")
	assert.Equal(t, 'æ', r, "second character must be rune 'æ'")
	r, n = b.PeekRune(3)
	assert.Equal(t, 3, n, "fourth character must be length 3")
	assert.Equal(t, '†', r, "fourth character must be rune '†'")
	r, n = b.PeekRune(6)
	assert.Equal(t, 4, n, "seventh character must be length 4")
	assert.Equal(t, '\U00100000', r, "seventh character must be rune '\U00100000'")
}

func TestShiftBufferZeroLen(t *testing.T) {
	var b = NewShifter(&ReaderMockup{bytes.NewBufferString("")})
	assert.Equal(t, byte(0), b.Peek(0), "first character must yield error")
}

////////////////////////////////////////////////////////////////

func ExampleNewShifter() {
	b := bytes.NewBufferString("Lorem ipsum")
	r := NewShifter(b)
	for {
		c := r.Peek(0)
		if c == ' ' {
			break
		}
		r.Move(1)
	}
	fmt.Println(string(r.Shift()))
	// Output: Lorem
}

func ExampleShifter_PeekRune() {
	b := bytes.NewBufferString("† dagger") // † has a byte length of 3
	r := NewShifter(b)

	c, n := r.PeekRune(0)
	fmt.Println(string(c), n)
	// Output: † 3
}

func ExampleShifter_IsEOF() {
	b := bytes.NewBufferString("Lorem ipsum") // bytes.Buffer provides a Bytes function, NewShifter uses that and r.IsEOF() always returns true
	r := NewShifter(b)
	r.Move(5)

	lorem := r.Shift()
	if !r.IsEOF() { // required when io.Reader doesn't provide a Bytes function
		buf := make([]byte, len(lorem))
		copy(buf, lorem)
		lorem = buf
	}

	r.Peek(0) // might reallocate the internal buffer
	fmt.Println(string(lorem))
	// Output: Lorem
}

////////////////////////////////////////////////////////////////

func BenchmarkPeek(b *testing.B) {
	r := NewShifter(bytes.NewBufferString("Lorem ipsum"))
	for i := 0; i < b.N; i++ {
		j := i % 11
		r.Peek(j)
	}
}

var c = 0
var haystack = []byte("abcdefghijklmnopqrstuvwxyz")

func BenchmarkBytesEqual(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := i % (len(haystack) - 3)
		if bytes.Equal([]byte("wxyz"), haystack[j:j+4]) {
			c++
		}
	}
}

func BenchmarkBytesEqual2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := i % (len(haystack) - 3)
		if bytes.Equal([]byte{'w', 'x', 'y', 'z'}, haystack[j:j+4]) {
			c++
		}
	}
}

func BenchmarkBytesEqual3(b *testing.B) {
	match := []byte{'w', 'x', 'y', 'z'}
	for i := 0; i < b.N; i++ {
		j := i % (len(haystack) - 3)
		if bytes.Equal(match, haystack[j:j+4]) {
			c++
		}
	}
}

func BenchmarkBytesEqual4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := i % (len(haystack) - 3)
		if bytesEqual(haystack[j:j+4], 'w', 'x', 'y', 'z') {
			c++
		}
	}
}

func bytesEqual(stack []byte, match ...byte) bool {
	return bytes.Equal(stack, match)
}

func BenchmarkCharsEqual(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := i % (len(haystack) - 3)
		if haystack[j] == 'w' && haystack[j+1] == 'x' && haystack[j+2] == 'y' && haystack[j+3] == 'z' {
			c++
		}
	}
}

func BenchmarkCharsLoopEqual(b *testing.B) {
	match := []byte("wxyz")
	for i := 0; i < b.N; i++ {
		j := i % (len(haystack) - 3)
		equal := true
		for k := 0; k < 4; k++ {
			if haystack[j+k] != match[k] {
				equal = false
				break
			}
		}
		if equal {
			c++
		}
	}
}

func BenchmarkCharsFuncEqual(b *testing.B) {
	match := []byte("wxyz")
	for i := 0; i < b.N; i++ {
		j := i % (len(haystack) - 3)
		if at(match, haystack[j:]) {
			c++
		}
	}
}

func at(match []byte, stack []byte) bool {
	if len(stack) < len(match) {
		return false
	}
	for i, c := range match {
		if stack[i] != c {
			return false
		}
	}
	return true
}
