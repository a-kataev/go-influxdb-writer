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
	maxCountLines uint32
	maxBufferSize uint32
}

func (o *Options) SetMaxCountLines(count uint32) *Options {
	o.maxCountLines = count
	return o
}

func (o *Options) SetMaxBufferSize(size uint32) *Options {
	o.maxBufferSize = size
	return o
}

type batch struct {
	options *Options
	lock    sync.RWMutex
	buffer  *bytes.Buffer
	counter uint32
}

func New(options *Options) Batch {
	b := &batch{
		buffer:  bytes.NewBuffer([]byte{}),
		options: options,
	}

	b.buffer.Grow(int(b.options.maxBufferSize))

	return b
}

var ErrLimit = errors.New("limit")

func (b *batch) Write(line string) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if uint32(len(line)+1+b.buffer.Len()) >= b.options.maxBufferSize {
		return ErrLimit
	}

	if b.counter+1 >= b.options.maxCountLines {
		return ErrLimit
	}

	_, err := b.buffer.WriteString(line + "\n")
	if err == nil {
		b.counter++
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

	// log.Printf("DEDUG batch %s", string(copyBuffer))

	return bytes.NewReader(copyBuffer)
}

func (b *batch) Reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.buffer.Reset()
	b.counter = 0
}
