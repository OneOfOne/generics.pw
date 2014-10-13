package main

import (
	"bytes"
	"sync"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 256))
		},
	}
)

func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func putBuffers(buffers ...*bytes.Buffer) {
	for _, b := range buffers {
		b.Reset()
		bufferPool.Put(b)
	}
}
