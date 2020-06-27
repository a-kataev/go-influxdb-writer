package batcher

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

type Batch interface {
	Write(line string) error
	Reader() io.Reader
	Reset()
}

type Options struct {
	linesLimit uint32
	batchSize  uint32
}

func (o *Options) SetLinesLimit(count uint32) *Options {
	o.linesLimit = count
	return o
}

func (o *Options) SetBatchSize(size uint32) *Options {
	o.batchSize = size
	return o
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

	b.buffer.Grow(int(b.options.batchSize))

	return b
}

var ErrLimitExceeded = errors.New("limit is exceeded")

func (b *batch) Write(line string) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if uint32(len(line)+1+b.buffer.Len()) >= b.options.batchSize {
		return ErrLimitExceeded
	}

	if b.lineCounter+1 >= b.options.linesLimit {
		return ErrLimitExceeded
	}

	_, err := b.buffer.WriteString(line + "\n")
	if err == nil {
		b.lineCounter++
	}

	return err
}

func (b *batch) Reader() io.Reader {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.buffer.Len() == 0 {
		return nil
	}

	copyBuffer := make([]byte, b.buffer.Len())
	copy(copyBuffer, b.buffer.Bytes())

	return bytes.NewReader(copyBuffer)
}

func (b *batch) Reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.buffer.Reset()
	b.lineCounter = 0
}
