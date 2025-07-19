package generator

import "bytes"

type Buffer struct {
	buf bytes.Buffer
}

func (b *Buffer) WriteString(s string) (n int, err error) {
	defer b.buf.WriteString("\n---\n")
	return b.buf.WriteString(s)
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	defer b.buf.WriteString("\n---\n")
	return b.buf.Write(p)
}

func (b *Buffer) Bytes() []byte {
	return b.buf.Bytes()
}
