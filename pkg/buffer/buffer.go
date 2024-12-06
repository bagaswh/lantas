package buffer

import "bytes"

const defaultMinSize = 1460  // TCP MSS
const defaultMaxSize = 65536 // 65 KiB, typical max TCP window size

func NewLimitedBuffer() *LimitedBuffer {
	b := make([]byte, defaultMinSize)
	buf := bytes.NewBuffer(b)
	return &LimitedBuffer{
		buf: buf,
	}
}

// LimitedBuffer is a buffer that can grow AND bounds the maximum buffer size.
type LimitedBuffer struct {
	buf     *bytes.Buffer
	b       []byte
	maxSize int
}

// func (b *LimitedBuffer) Write(p []byte) (int, error) {
// 	if len(p) > b.buf.Available() {

// 	}
// }

// func (b *LimitedBuffer) grow() (int, bool) {
// 	newCap := cap(b.b)
// }
