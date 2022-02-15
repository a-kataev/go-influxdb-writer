//go:generate mockery -name Batch -filename batch.go

package batch

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

type BatchReader struct {
	Reader  io.Reader
	Size    uint64
	Entries uint64
}

type Batch interface {
	Write(e []byte) error
	Reader() *BatchReader
	Reset()
}

type Options struct {
	BufferSize   uint64
	EntriesLimit uint64
}

type batch struct {
	lock         sync.RWMutex
	buffer       *bytes.Buffer
	entries      uint64
	bufferSize   uint64
	entriesLimit uint64
}

func New(options *Options) Batch {
	b := &batch{
		buffer:       bytes.NewBuffer([]byte{}),
		bufferSize:   options.BufferSize,
		entriesLimit: options.EntriesLimit,
	}

	b.buffer.Grow(int(b.bufferSize))

	return b
}

var (
	ErrLimitExceeded = errors.New("entries limit exceeded")
	ErrSizeExceeded  = errors.New("buffer size exceeded")
)

func (b *batch) Write(e []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if uint64(b.buffer.Len()+len(e)+1) >= b.bufferSize {
		return ErrSizeExceeded
	}

	if b.entries+1 >= b.entriesLimit {
		return ErrLimitExceeded
	}

	_, _ = b.buffer.Write(e)
	_, _ = b.buffer.WriteRune('\n')

	b.entries++

	return nil
}

func (b *batch) Reader() *BatchReader {
	b.lock.RLock()
	defer b.lock.RUnlock()

	copyBuffer := make([]byte, b.buffer.Len())
	copy(copyBuffer, b.buffer.Bytes())

	return &BatchReader{
		Reader:  bytes.NewReader(copyBuffer),
		Size:    uint64(b.buffer.Len()),
		Entries: b.entries,
	}
}

func (b *batch) Reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.buffer.Reset()
	b.entries = 0
}
