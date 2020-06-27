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

	"github.com/a-kataev/go-influxdb-writer/internal/version"
)

type Client interface {
	Send(ctx context.Context, data io.Reader) error
}

type Options struct {
	ServerURL   string
	AuthToken   string
	Bucket      string
	Precision   string
	HTTPTimeout time.Duration
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

	c.http.Timeout = c.options.HTTPTimeout

	return c
}

type responseError struct {
	Error string `json:"error"`
}

func (c *client) Send(ctx context.Context, data io.Reader) error {
	reqQuery := url.Values{}

	if len(c.options.Bucket) > 0 {
		reqQuery.Add("bucket", c.options.Bucket)
	}

	if len(c.options.Precision) > 0 {
		reqQuery.Add("precision", c.options.Precision)
	}

	reqURL := c.options.ServerURL + "/api/v2/write"

	if len(reqQuery.Encode()) > 0 {
		reqURL += "?" + reqQuery.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, data)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Token "+c.options.AuthToken)
	req.Header.Add("User-Agent", "go-influxdb-writer "+version.Version)

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
