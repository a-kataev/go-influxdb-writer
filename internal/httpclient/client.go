package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/a-kataev/go-influxdb-writer/internal/version"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

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
	http    httpClient
	sendURL string
	options *Options
}

func New(options *Options) Client {
	c := &client{
		options: options,
	}

	c.http = &http.Client{
		Timeout: c.options.HTTPTimeout,
	}

	c.sendURL = c.makeSendURL()

	return c
}

func (c *client) makeSendURL() string {
	params := url.Values{}

	changed := false

	if len(c.options.Bucket) > 0 {
		params.Add("bucket", c.options.Bucket)
		changed = true
	}

	if len(c.options.Precision) > 0 {
		params.Add("precision", c.options.Precision)
		changed = true
	}

	url := c.options.ServerURL + "/api/v2/write"

	if changed {
		url += "?" + params.Encode()
	}

	return url
}

type SendError struct {
	StatusCode    int
	ResponseError string
	HTTPError     string
	RequestID     string
}

func (e *SendError) Error() string {
	if len(e.ResponseError) > 0 {
		return e.ResponseError
	}
	return e.HTTPError
}

type responseError struct {
	Error string `json:"error"`
}

func (c *client) Send(ctx context.Context, data io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.sendURL, data)
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
		sendErr := &SendError{
			StatusCode: resp.StatusCode,
			RequestID:  resp.Header.Get("X-Request-Id"),
		}

		respErr := &responseError{}

		if err := json.Unmarshal(body, respErr); err != nil {
			sendErr.HTTPError = strings.ReplaceAll(string(body), "\n", " ")
		} else {
			sendErr.ResponseError = respErr.Error
		}

		return sendErr
	}

	return nil
}
