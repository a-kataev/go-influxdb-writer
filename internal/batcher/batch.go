package batcher

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

type Batch interface {
	Write(line string) error
	Reader() (io.Reader, uint32, uint32)
	Reset()
}

type Options struct {
	LinesLimit uint32
	BatchSize  uint32
}

type batch struct {
	options     *Options
	lock        sync.RWMutex
	buffer      *bytes.Buffer
	lineCounter uint32
}

func New(options *Options) Batch {
	b := &batch{
		buffer:  bytes.NewBuffer([]byte{}),
		options: options,
	}

	b.buffer.Grow(int(b.options.BatchSize))

	return b
}

var ErrLimitExceeded = errors.New("limit is exceeded")

func (b *batch) Write(line string) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if uint32(b.buffer.Len()+len(line)+1) >= b.options.BatchSize {
		return ErrLimitExceeded
	}

	if b.lineCounter+1 >= b.options.LinesLimit {
		return ErrLimitExceeded
	}

	_, _ = b.buffer.WriteString(line + "\n")

	b.lineCounter++

	return nil
}

func (b *batch) Reader() (io.Reader, uint32, uint32) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.buffer.Len() == 0 {
		return nil, 0, 0
	}

	copyBuffer := make([]byte, b.buffer.Len())
	copy(copyBuffer, b.buffer.Bytes())

	return bytes.NewReader(copyBuffer), uint32(b.buffer.Len()), b.lineCounter
}

func (b *batch) Reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.buffer.Reset()
	b.lineCounter = 0
}
