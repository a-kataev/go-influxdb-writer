package writer

import (
	"context"
	"errors"
	"go-influxdb-writer/internal/batcher"
	"go-influxdb-writer/internal/httpclient"
	"time"
)

type Writer interface {
	WriteLine(line string)
	Close()
}

type Config struct {
	ServerURL     string
	AuthToken     string
	Bucket        string
	Precision     string
	HTTPTimeout   time.Duration
	LinesLimit    uint32
	BatchSize     uint32
	FlushInterval time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		ServerURL:     "http://localhost:8086",
		AuthToken:     "admin:password",
		Bucket:        "test",
		Precision:     "ns",
		HTTPTimeout:   10 * time.Second,
		LinesLimit:    5000,
		BatchSize:     1024 * 1024 * 3,
		FlushInterval: 10 * time.Second,
	}
}

type Options struct {
	flushInterval time.Duration
	flushTimeout  time.Duration
}

func (o *Options) SetFlushInterval(interval time.Duration) *Options {
	o.flushInterval = interval
	return o
}

func (o *Options) SetFlushTimeout(timeout time.Duration) *Options {
	o.flushTimeout = timeout
	return o
}

type writer struct {
	client  httpclient.Client
	batch   batcher.Batch
	writeCh chan string
	logger  Logger
	options *Options
}

func New(logger Logger, config *Config) Writer {
	clientOptions := new(httpclient.Options).
		SetServerURL(config.ServerURL).
		SetAuthToken(config.AuthToken).
		SetBucket(config.Bucket).
		SetPrecision(config.Precision).
		SetHTTPTimeout(config.HTTPTimeout)

	client := httpclient.New(clientOptions)

	batchOptions := new(batcher.Options).
		SetLinesLimit(config.LinesLimit).
		SetBatchSize(config.BatchSize)

	batch := batcher.New(batchOptions)

	writerOptions := new(Options).
		SetFlushInterval(config.FlushInterval).
		SetFlushTimeout(config.HTTPTimeout)

	if logger == nil {
		logger = &defaultLogger{}
	}

	return NewWriter(client, batch, logger, writerOptions)
}

func NewWriter(client httpclient.Client, batch batcher.Batch, logger Logger, options *Options) Writer {
	w := &writer{
		client:  client,
		batch:   batch,
		writeCh: make(chan string),
		logger:  logger,
		options: options,
	}

	w.logger.Infof("writer: started")

	go w.handler()

	return w
}

func (w *writer) handler() {
	ticker := time.NewTicker(w.options.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case line, ok := <-w.writeCh:
			if !ok {
				return
			}

			if err := w.batch.Write(line); errors.Is(err, batcher.ErrLimitExceeded) {
				w.flush()

				if err := w.batch.Write(line); err != nil {
					w.logger.Errorf("batch.write: %s", err)
				}

				ticker.Stop()
				ticker = time.NewTicker(w.options.flushInterval)
			}
		case <-ticker.C:
			w.flush()
		}
	}
}

func (w *writer) flush() {
	ctx, cancel := context.WithTimeout(context.Background(), w.options.flushTimeout)
	defer cancel()

	if reader := w.batch.Reader(); reader != nil {
		if err := w.client.SendBatch(ctx, reader); err != nil {
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
	w.flush()
	w.logger.Infof("writer: stopped")
}
