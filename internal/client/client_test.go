package client

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_New(t *testing.T) {
	testClient := New(&Options{})
	assert.IsType(t, &client{}, testClient)
}

func Test_makeURL(t *testing.T) {
	tables := []struct {
		options *Options
		url     string
	}{
		{
			options: &Options{},
			url:     "/api/v2/write",
		},
		{
			options: &Options{
				ServerURL: "test",
			},
			url: "test/api/v2/write",
		},
		{
			options: &Options{
				Bucket: "test",
			},
			url: "/api/v2/write?bucket=test",
		},
		{
			options: &Options{
				Precision: "test",
			},
			url: "/api/v2/write?precision=test",
		},
	}

	for tt, table := range tables {
		testClient := &client{
			options: table.options,
		}

		url := testClient.makeURL()
		assert.Equalf(t, table.url, url, "%d", tt)
	}
}

func Test_makeRequest(t *testing.T) {
	testClient := &client{
		options: &Options{},
	}

	request, err := testClient.makeRequest(nil, nil) //nolint
	assert.Nil(t, request)
	assert.EqualError(t, err, "net/http: nil Context")

	request, err = testClient.makeRequest(context.Background(), nil)
	assert.IsType(t, &http.Request{}, request)
	assert.Nil(t, err)
}

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test")
}

func (e *errReader) Close() error {
	return errors.New("test")
}

func Test_makeResponse(t *testing.T) {
	testClient := &client{
		options: &Options{},
	}

	clientResponse, err := testClient.makeResponse(&http.Response{
		Body: &errReader{},
	})
	assert.Nil(t, clientResponse)
	assert.EqualError(t, err, "test")

	tables := []struct {
		statusCode     int
		responseBody   []byte
		clientResponse *ClientResponse
	}{
		{
			statusCode:   204,
			responseBody: []byte{},
			clientResponse: &ClientResponse{
				StatusCode: 204,
			},
		},
		{
			statusCode:   500,
			responseBody: []byte("test"),
			clientResponse: &ClientResponse{
				StatusCode: 500,
				Response:   "test",
			},
		},
		{
			statusCode:   500,
			responseBody: []byte(`{"error":"test"}`),
			clientResponse: &ClientResponse{
				StatusCode:    500,
				ResponseError: "test",
			},
		},
	}

	for tt, table := range tables {
		response := &http.Response{
			StatusCode: table.statusCode,
			Body:       ioutil.NopCloser(bytes.NewBuffer(table.responseBody)),
		}

		clientResponse, err := testClient.makeResponse(response)
		assert.Equalf(t, table.clientResponse, clientResponse, "%d", tt)
		assert.Nilf(t, err, "%d", tt)
	}
}

func Test_Send(t *testing.T) {
	testClient := &client{
		options: &Options{},
	}

	clientResponse, err := testClient.Send(nil, nil) //nolint
	assert.Nil(t, clientResponse)
	assert.EqualError(t, err, "net/http: nil Context")

	testHTTPClient := &mockHTTPClient{}
	testHTTPClient.On("Do", mock.Anything, mock.Anything).Return(nil, errors.New("test"))
	testClient.http = testHTTPClient
	clientResponse, err = testClient.Send(context.Background(), nil)
	assert.Nil(t, clientResponse)
	assert.EqualError(t, err, "test")

	testHTTPClient = &mockHTTPClient{}
	testHTTPClient.On("Do", mock.Anything, mock.Anything).Return(&http.Response{
		Body: &errReader{},
	}, nil)
	testClient.http = testHTTPClient
	clientResponse, err = testClient.Send(context.Background(), nil)
	assert.Nil(t, clientResponse)
	assert.EqualError(t, err, "test")
}
