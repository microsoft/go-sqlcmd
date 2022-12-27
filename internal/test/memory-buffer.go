package test

import "bytes"

// MemoryBuffer has both Write and Close methods for use as io.WriteCloser
// when testing (instead of os.Stdout), so tests can assert.Equal results etc.
type MemoryBuffer struct {
	buf *bytes.Buffer
}

func (b *MemoryBuffer) Write(p []byte) (n int, err error) {
	return b.buf.Write(p)
}

func (b *MemoryBuffer) Close() error {
	return nil
}

func (b *MemoryBuffer) String() string {
	return b.buf.String()
}

func NewMemoryBuffer() *MemoryBuffer {
	return &MemoryBuffer{buf: new(bytes.Buffer)}
}
