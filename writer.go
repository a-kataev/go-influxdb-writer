package writer

import (
	"context"
	"errors"
	"time"

	"github.com/a-kataev/go-influxdb-writer/internal/batcher"
	"github.com/a-kataev/go-influxdb-writer/internal/httpclient"
)

type Writer interface {
	WriteLine(line string)
	Close()
}

type writerOptions struct {
	SendInterval time.Duration
	SendTimeout  time.Duration
}

type writer struct {
	client  httpclient.Client
	batch   batcher.Batch
	writeCh chan string
	logger  Logger
	options *writerOptions
}

func New(logger Logger, options *Options) Writer {
	if logger == nil {
		logger = &defaultLogger{}
	}

	w := &writer{
		client:  httpclient.New(options.client),
		batch:   batcher.New(options.batch),
		writeCh: make(chan string),
		logger:  logger,
		options: options.writer,
	}

	w.logger.Infof("writer: started")

	go w.run()

	return w
}

func (w *writer) run() {
	ticker := time.NewTicker(w.options.SendInterval)
	defer ticker.Stop()

	for {
		select {
		case line, ok := <-w.writeCh:
			if !ok {
				return
			}

			if err := w.batch.Write(line); errors.Is(err, batcher.ErrLimitExceeded) {
				w.send()

				if err := w.batch.Write(line); err != nil {
					w.logger.Errorf("batch.write: %s", err)
				}

				ticker.Stop()
				ticker = time.NewTicker(w.options.SendInterval)
			} else if err != nil {
				w.logger.Errorf("batch.write: %s", err)
			}
		case <-ticker.C:
			w.send()
		}
	}
}

func (w *writer) send() {
	ctx, cancel := context.WithTimeout(context.Background(), w.options.SendTimeout)
	defer cancel()

	if reader := w.batch.Reader(); reader != nil {
		if err := w.client.Send(ctx, reader); err != nil {
			w.logger.Errorf("client.send: %s", err)
		}
	}

	w.batch.Reset()
}

func (w *writer) WriteLine(line string) {
	w.writeCh <- line
}

func (w *writer) Close() {
	close(w.writeCh)

	w.send()

	w.logger.Infof("writer: stopped")
}
