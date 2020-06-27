package writer

import (
	"time"

	"github.com/a-kataev/go-influxdb-writer/internal/batcher"
	"github.com/a-kataev/go-influxdb-writer/internal/httpclient"
)

type Options struct {
	writer writerOptions
	client httpclient.Options
	batch  batcher.Options
}

func (o *Options) SetFlushInterval(interval time.Duration) *Options {
	o.writer.flushInterval = interval
	return o
}

func (o *Options) SetFlushTimeout(timeout time.Duration) *Options {
	o.writer.flushTimeout = timeout
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
