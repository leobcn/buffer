package buffer

import "errors"

// MinBuf and MaxBuf are the initial and maximal internal buffer size.
var MinBuf = 1024
var MaxBuf = 1048576 // upper limit 1MB

// ErrExceeded is returned when the internal buffer exceeds 4096 bytes, a string or comment must thus be smaller than 4kB!
var ErrExceeded = errors.New("max buffer exceeded")
