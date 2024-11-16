package httpclient

import "bytes"

const truncatedMsg = "... [truncated]"

type limitedBuffer struct {
	buf     bytes.Buffer
	maxSize int64
}

func newLimitedBuffer(maxSize int64) *limitedBuffer {
	return &limitedBuffer{maxSize: maxSize}
}

func (b *limitedBuffer) Write(p []byte) (n int, err error) {
	if int64(b.buf.Len())+int64(len(p)) > b.maxSize {
		remaining := int(b.maxSize) - b.buf.Len()
		if remaining > 0 {
			n, err = b.buf.Write(p[:remaining])
			b.buf.WriteString(truncatedMsg)
		}
		return len(p), nil
	}
	return b.buf.Write(p)
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}
