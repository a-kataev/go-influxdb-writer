package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	version = "v0.1.1"
)

type Client interface {
	SendBatch(ctx context.Context, batch io.Reader) error
}

type Options struct {
	serverURL   string
	authToken   string
	bucket      string
	precision   string
	httpTimeout time.Duration
}

func (o *Options) SetServerURL(url string) *Options {
	o.serverURL = url
	return o
}

func (o *Options) SetAuthToken(token string) *Options {
	o.authToken = token
	return o
}

func (o *Options) SetBucket(bucket string) *Options {
	o.bucket = bucket
	return o
}

func (o *Options) SetPrecision(precision string) *Options {
	o.precision = precision
	return o
}

func (o *Options) SetHTTPTimeout(timeout time.Duration) *Options {
	o.httpTimeout = timeout
	return o
}

type client struct {
	http    *http.Client
	options *Options
}

func New(options *Options) Client {
	c := &client{
		http:    &http.Client{},
		options: options,
	}

	c.http.Timeout = c.options.httpTimeout

	return c
}

type responseError struct {
	Error string `json:"error"`
}

func (c *client) SendBatch(ctx context.Context, batch io.Reader) error {
	reqQuery := url.Values{}

	if len(c.options.bucket) > 0 {
		reqQuery.Add("bucket", c.options.bucket)
	}

	if len(c.options.precision) > 0 {
		reqQuery.Add("precision", c.options.precision)
	}

	reqURL := c.options.serverURL + "/api/v2/write"

	if len(reqQuery.Encode()) > 0 {
		reqURL += "?" + reqQuery.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, batch)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Token "+c.options.authToken)
	req.Header.Add("User-Agent", "go-influxdb-writer "+version)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return err
	}
	_, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 204 {
		respErr := &responseError{}
		if err := json.Unmarshal(body, respErr); err != nil {
			return fmt.Errorf("code: %d, body: '%s'", resp.StatusCode, string(body))
		}
		return errors.New(respErr.Error)
	}

	return nil
}
