//go:generate mockery -name Client -filename client.go
//go:generate mockery -name httpClient -structname mockHTTPClient -inpkg -filename client_mock_test.go

package client

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type httpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type ClientResponse struct {
	RequestID     string
	StatusCode    int
	Response      string
	ResponseError string
}

type Client interface {
	Send(ctx context.Context, reader io.Reader) (*ClientResponse, error)
}

type Options struct {
	ServerURL   string
	AuthToken   string
	Bucket      string
	Precision   string
	HTTPTimeout time.Duration
}

type client struct {
	http  httpClient
	url   string
	token string
}

func New(options *Options) Client {
	c := &client{
		http: &http.Client{
			Timeout: options.HTTPTimeout,
		},
	}

	c.url = makeURL(options)
	c.token = options.AuthToken

	return c
}

func makeURL(options *Options) string {
	params := url.Values{}

	changed := false

	if len(options.Bucket) > 0 {
		params.Add("bucket", options.Bucket)
		changed = true
	}

	if len(options.Precision) > 0 {
		params.Add("precision", options.Precision)
		changed = true
	}

	url := options.ServerURL + "/api/v2/write"

	if changed {
		url += "?" + params.Encode()
	}

	return url
}

func (c *client) makeRequest(ctx context.Context, reader io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.url, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Token "+c.token)
	req.Header.Add("User-Agent", "go-influxdb-writer")

	return req, nil
}

func (c *client) makeResponse(resp *http.Response) (*ClientResponse, error) {
	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 512))
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	clientResp := &ClientResponse{
		RequestID:  resp.Header.Get("X-Request-Id"),
		StatusCode: resp.StatusCode,
	}

	if resp.StatusCode != 204 {
		respErr := map[string]string{"error": ""}

		if err := json.Unmarshal(body, &respErr); err != nil {
			clientResp.Response = strings.ReplaceAll(string(body), "\n", " ")
		} else {
			clientResp.ResponseError = strings.ReplaceAll(respErr["error"], "\n", " ")
		}
	}

	return clientResp, nil
}

func (c *client) Send(ctx context.Context, reader io.Reader) (*ClientResponse, error) {
	req, err := c.makeRequest(ctx, reader)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	return c.makeResponse(resp)
}
