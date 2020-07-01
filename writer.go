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

	w.logger.Infof("started")

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

	if reader, size, count := w.batch.Reader(); reader != nil {
		err := w.client.Send(ctx, reader)

		if sendErr, ok := err.(*httpclient.SendError); ok {
			if len(sendErr.RequestID) > 0 {
				w.logger.Errorf("client.send: request_id: %s, status_code: %d, error: '%s'", sendErr.RequestID, sendErr.StatusCode, err)
			} else {
				w.logger.Errorf("client.send: status_code: %d, error: '%s'", sendErr.StatusCode, err)
			}
		} else {
			w.logger.Errorf("client.send: %s", err)
		}

		if err == nil {
			w.logger.Infof("send batch: size: %d, cound: %d", size, count)

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

	w.logger.Infof("stopped")
}
