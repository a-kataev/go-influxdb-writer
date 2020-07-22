package httpclient

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	c := New(&Options{})
	require.IsType(t, &client{}, c)
}

func Test_makeSendURL(t *testing.T) {
	tables := []struct {
		options *Options
		sendURL string
	}{
		{
			options: &Options{},
			sendURL: "/api/v2/write",
		},
		{
			options: &Options{
				ServerURL: "test",
			},
			sendURL: "test/api/v2/write",
		},
		{
			options: &Options{
				Bucket: "test",
			},
			sendURL: "/api/v2/write?bucket=test",
		},
		{
			options: &Options{
				Precision: "test",
			},
			sendURL: "/api/v2/write?precision=test",
		},
	}

	for _, table := range tables {
		c := &client{
			options: table.options,
		}

		url := c.makeSendURL()

		require.Equal(t, table.sendURL, url)
	}
}

func Test_SendError(t *testing.T) {
	err := &SendError{
		HTTPError: "test",
	}

	require.EqualError(t, err, "test")

	err.HTTPError = ""
	err.ResponseError = "test"

	require.EqualError(t, err, "test")
}

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test")
}

func (e *errReader) Close() error {
	return errors.New("test")
}

func Test_Send(t *testing.T) {
	c := &client{
		http:    &http.Client{},
		options: &Options{},
	}

	var err error

	err = c.Send(nil, nil) //nolint
	require.EqualError(t, err, "net/http: nil Context")

	httpDoErr := &mockHTTPClient{}
	httpDoErr.On("Do", mock.Anything, mock.Anything).Return(nil, errors.New("test"))
	c.http = httpDoErr

	err = c.Send(context.Background(), nil)
	require.EqualError(t, err, "test")

	httpBodyErr := &mockHTTPClient{}
	httpBodyErr.On("Do", mock.Anything, mock.Anything).Return(
		&http.Response{
			Body: &errReader{},
		},
		nil)
	c.http = httpBodyErr

	err = c.Send(context.Background(), nil)
	require.EqualError(t, err, "test")
}

func Test_Send_Response(t *testing.T) {
	tables := []struct {
		statusCode   int
		responseBody []byte
		returnErr    error
	}{
		{
			statusCode:   204,
			responseBody: []byte{},
			returnErr:    nil,
		},
		{
			statusCode:   500,
			responseBody: []byte("test"),
			returnErr: &SendError{
				StatusCode: 500,
				HTTPError:  "test",
			},
		},
		{
			statusCode:   500,
			responseBody: []byte(`{"error":"test"}`),
			returnErr: &SendError{
				StatusCode:    500,
				ResponseError: "test",
			},
		},
	}

	c := &client{
		http:    &http.Client{},
		options: &Options{},
	}

	for _, table := range tables {
		func() {
			ts := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(table.statusCode)
					_, _ = w.Write(table.responseBody)
				}),
			)
			defer ts.Close()

			c.options.ServerURL = ts.URL
			c.sendURL = c.makeSendURL()

			r := bytes.NewReader([]byte("test"))

			err := c.Send(context.Background(), r)
			if table.returnErr == nil {
				require.Nil(t, err)
				return
			}
			require.IsType(t, &SendError{}, err)
			sendErr, ok := err.(*SendError)
			require.True(t, ok)
			require.Equal(t, sendErr, table.returnErr)
		}()
	}
}
