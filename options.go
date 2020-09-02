package writer

import (
	"time"

	"github.com/a-kataev/go-influxdb-writer/internal/batch"
	"github.com/a-kataev/go-influxdb-writer/internal/client"
)

// Options
type Options struct {
	logger Logger
	writer *writerOptions
	client *client.Options
	batch  *batch.Options
}

// DefaultOptions
func DefaultOptions() *Options {
	return &Options{
		logger: &defaultLogger{},
		writer: &writerOptions{
			SendInterval: 10 * time.Second,
			SendTimeout:  9 * time.Second,
		},
		client: &client.Options{
			ServerURL:   "http://localhost:8086",
			AuthToken:   "admin:password",
			Bucket:      "test",
			Precision:   "ns",
			HTTPTimeout: 8 * time.Second,
		},
		batch: &batch.Options{
			BufferSize:   1024 * 1024 * 3,
			EntriesLimit: 5000,
		},
	}
}

// SetLogger
func (o *Options) SetLogger(logger Logger) *Options {
	o.logger = logger
	return o
}

// SetSendInterval
func (o *Options) SetSendInterval(interval time.Duration) *Options {
	o.writer.SendInterval = interval
	return o
}

// SetSendTimeout
func (o *Options) SetSendTimeout(timeout time.Duration) *Options {
	o.writer.SendTimeout = timeout
	return o
}

// SetServerURL
func (o *Options) SetServerURL(url string) *Options {
	o.client.ServerURL = url
	return o
}

// SetAuthToken
func (o *Options) SetAuthToken(token string) *Options {
	o.client.AuthToken = token
	return o
}

// SetBucket
func (o *Options) SetBucket(bucket string) *Options {
	o.client.Bucket = bucket
	return o
}

// SetPrecision
func (o *Options) SetPrecision(precision string) *Options {
	o.client.Precision = precision
	return o
}

// SetHTTPTimeout
func (o *Options) SetHTTPTimeout(timeout time.Duration) *Options {
	o.client.HTTPTimeout = timeout
	return o
}

// SetBatchSize
func (o *Options) SetBatchSize(size uint64) *Options {
	o.batch.BufferSize = size
	return o
}

// SetEntriesLimit
func (o *Options) SetEntriesLimit(limit uint64) *Options {
	o.batch.EntriesLimit = limit
	return o
}
