package buffer // import "github.com/tdewolff/buffer"

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	r := NewWriter(make([]byte, 0, 3))

	n, _ := r.Write([]byte("abc"))
	assert.Equal(t, 3, n, "first write must write 3 characters")
	assert.Equal(t, []byte("abc"), r.Bytes(), "first write must match 'abc'")

	n, _ = r.Write([]byte("def"))
	assert.Equal(t, 3, n, "second write must write 3 characters")
	assert.Equal(t, []byte("abcdef"), r.Bytes(), "second write must match 'abcdef'")

	r.Reset()
	assert.Equal(t, []byte(""), r.Bytes(), "reset must match ''")

	n, _ = r.Write([]byte("ghijkl"))
	assert.Equal(t, 6, n, "third write must write 6 characters")
	assert.Equal(t, []byte("ghijkl"), r.Bytes(), "third write must match 'ghijkl'")
}

func ExampleNewWriter() {
	w := NewWriter(make([]byte, 0, 11)) // initial buffer length is 11
	w.Write([]byte("Lorem ipsum"))
	fmt.Println(string(w.Bytes()))
	// Output: Lorem ipsum
}

func ExampleWriter_Reset() {
	w := NewWriter(make([]byte, 0, 11))                 // initial buffer length is 10
	w.Write([]byte("garbage that will be overwritten")) // does reallocation
	w.Reset()
	w.Write([]byte("Lorem ipsum"))
	fmt.Println(string(w.Bytes()))
	// Output: Lorem ipsum
}
