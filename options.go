package writer

import (
	"time"

	"github.com/a-kataev/go-influxdb-writer/internal/batcher"
	"github.com/a-kataev/go-influxdb-writer/internal/httpclient"
)

type Options struct {
	writer *writerOptions
	client *httpclient.Options
	batch  *batcher.Options
}

func DefaultOptions() *Options {
	return &Options{
		writer: &writerOptions{
			SendInterval: 10 * time.Second,
			SendTimeout:  9 * time.Second,
		},
		client: &httpclient.Options{
			ServerURL:   "http://localhost:8086",
			AuthToken:   "admin:password",
			Bucket:      "test",
			Precision:   "ns",
			HTTPTimeout: 8 * time.Second,
		},
		batch: &batcher.Options{
			LinesLimit: 5000,
			BatchSize:  1024 * 1024 * 3,
		},
	}
}

func (o *Options) SetSendInterval(interval time.Duration) *Options {
	o.writer.SendInterval = interval
	return o
}

func (o *Options) SetSendTimeout(timeout time.Duration) *Options {
	o.writer.SendTimeout = timeout
	return o
}

func (o *Options) SetServerURL(url string) *Options {
	o.client.ServerURL = url
	return o
}

func (o *Options) SetAuthToken(token string) *Options {
	o.client.AuthToken = token
	return o
}

func (o *Options) SetBucket(bucket string) *Options {
	o.client.Bucket = bucket
	return o
}

func (o *Options) SetPrecision(precision string) *Options {
	o.client.Precision = precision
	return o
}

func (o *Options) SetHTTPTimeout(timeout time.Duration) *Options {
	o.client.HTTPTimeout = timeout
	return o
}

func (o *Options) SetLinesLimit(count uint32) *Options {
	o.batch.LinesLimit = count
	return o
}

func (o *Options) SetBatchSize(size uint32) *Options {
	o.batch.BatchSize = size
	return o
}
