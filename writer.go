package writer

import (
	"context"
	"time"

	"github.com/a-kataev/go-influxdb-writer/internal/batch"
	"github.com/a-kataev/go-influxdb-writer/internal/client"
)

type Writer interface {
	WriteLine(line string)
	Write(b []byte)
	Close()
}

type writerOptions struct {
	SendInterval time.Duration
	SendTimeout  time.Duration
}

type writer struct {
	client       client.Client
	batch        batch.Batch
	write        chan []byte
	sendInterval time.Duration
	sendTimeout  time.Duration
	logger       Logger
}

func NewWriter(serverURL, authToken, bucket string) Writer {
	return NewWriterWithOptions(DefaultOptions().
		SetServerURL(serverURL).SetAuthToken(authToken).SetBucket(bucket))
}

func NewWriterWithOptions(options *Options) Writer {
	if options == nil {
		options = DefaultOptions()
	}
	w := &writer{
		client:       client.New(options.Client),
		batch:        batch.New(options.Batch),
		write:        make(chan []byte),
		sendInterval: options.Writer.SendInterval,
		sendTimeout:  options.Writer.SendTimeout,
		logger:       options.Logger,
	}

	go w.run()

	return w
}

func (w *writer) run() {
	w.logger.Infof("started")

	ticker := time.NewTicker(w.sendInterval)
	defer ticker.Stop()

	for {
		select {
		case b, ok := <-w.write:
			if !ok {
				return
			}

			if err := w.batch.Write(b); err == nil {
				w.send()

				if err := w.batch.Write(b); err != nil {
					w.logger.Errorf("batch.write: %s", err)
				}

				ticker.Stop()
				ticker = time.NewTicker(w.sendInterval)
			} else if err != nil {
				w.logger.Errorf("batch.write: %s", err)
			}
		case <-ticker.C:
			w.send()
		}
	}
}

func (w *writer) send() {
	ctx, cancel := context.WithTimeout(context.Background(), w.sendTimeout)
	defer cancel()

	defer w.batch.Reset()

	reader := w.batch.Reader()
	if reader.Size == 0 && reader.Entries == 0 {
		return
	}

	resp, err := w.client.Send(ctx, reader.Reader)
	if err != nil {
		w.logger.Errorf("client.send: %s", err)
		return
	}

	if resp.StatusCode == 204 {
		w.logger.Infof("send batch: size: %d, entries: %d",
			reader.Size, reader.Entries)
		return
	}

	if len(resp.ResponseError) > 0 {
		w.logger.Errorf("client.send: request_id: %s, status_code: %d, error: '%s'",
			resp.RequestID, resp.StatusCode, resp.ResponseError)
		return
	}

	w.logger.Errorf("client.send: request_id: %s, status_code: %d, response: '%s'",
		resp.RequestID, resp.StatusCode, resp.Response)
}

func (w *writer) WriteLine(line string) {
	w.Write([]byte(line))
}

func (w *writer) Write(b []byte) {
	w.write <- b
}

func (w *writer) Close() {
	close(w.write)

	w.send()

	w.logger.Infof("stopped")
}
