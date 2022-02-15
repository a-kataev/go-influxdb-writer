package writer

import (
	"time"

	"github.com/a-kataev/go-influxdb-writer/internal/batch"
	"github.com/a-kataev/go-influxdb-writer/internal/client"
)

type Options struct {
	Client *client.Options
	Batch  *batch.Options
	Writer *writerOptions
	Logger Logger
}

func DefaultOptions() *Options {
	return &Options{
		Client: &client.Options{
			ServerURL:   "http://localhost:8086",
			AuthToken:   "admin:password",
			Bucket:      "test",
			Precision:   "ns",
			HTTPTimeout: 8 * time.Second,
		},
		Batch: &batch.Options{
			BufferSize:   1024 * 1024 * 3,
			EntriesLimit: 5000,
		},
		Writer: &writerOptions{
			SendInterval: 10 * time.Second,
			SendTimeout:  9 * time.Second,
		},
		Logger: &defaultLogger{},
	}
}

func (o *Options) SetLogger(logger Logger) *Options {
	o.Logger = logger
	return o
}

func (o *Options) SetSendInterval(interval time.Duration) *Options {
	o.Writer.SendInterval = interval
	return o
}

func (o *Options) SetSendTimeout(timeout time.Duration) *Options {
	o.Writer.SendTimeout = timeout
	return o
}

func (o *Options) SetServerURL(url string) *Options {
	o.Client.ServerURL = url
	return o
}

func (o *Options) SetAuthToken(token string) *Options {
	o.Client.AuthToken = token
	return o
}

func (o *Options) SetBucket(bucket string) *Options {
	o.Client.Bucket = bucket
	return o
}

func (o *Options) SetPrecision(precision string) *Options {
	o.Client.Precision = precision
	return o
}

func (o *Options) SetHTTPTimeout(timeout time.Duration) *Options {
	o.Client.HTTPTimeout = timeout
	return o
}

func (o *Options) SetBatchSize(size uint64) *Options {
	o.Batch.BufferSize = size
	return o
}

func (o *Options) SetEntriesLimit(limit uint64) *Options {
	o.Batch.EntriesLimit = limit
	return o
}
