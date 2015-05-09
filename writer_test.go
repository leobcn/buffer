package buffer // import "github.com/tdewolff/buffer"

import (
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
