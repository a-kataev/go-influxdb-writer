package writer

import (
	"errors"
	"testing"
	"time"

	"github.com/a-kataev/go-influxdb-writer/internal/batch"
	mocksBatch "github.com/a-kataev/go-influxdb-writer/internal/batch/mocks"
	"github.com/a-kataev/go-influxdb-writer/internal/client"
	mocksClient "github.com/a-kataev/go-influxdb-writer/internal/client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_New_Close(t *testing.T) {
	testWriter1 := NewWriterWithOptions(nil)
	testWriter1.Close()

	testWriter2 := NewWriter("url", "token", "bucket")
	testWriter2.Close()

	logger := &mockLogger{
		InfoLines:  make([]string, 0),
		ErrorLines: make([]string, 0),
	}
	defaultOptions := DefaultOptions()
	options := defaultOptions.
		SetLogger(logger).
		SetSendInterval(defaultOptions.Writer.SendInterval).
		SetSendTimeout(defaultOptions.Writer.SendTimeout).
		SetServerURL(defaultOptions.Client.ServerURL).
		SetAuthToken(defaultOptions.Client.AuthToken).
		SetBucket(defaultOptions.Client.Bucket).
		SetPrecision(defaultOptions.Client.Precision).
		SetHTTPTimeout(defaultOptions.Client.HTTPTimeout).
		SetBatchSize(defaultOptions.Batch.BufferSize).
		SetEntriesLimit(defaultOptions.Batch.EntriesLimit)
	testWriter3 := NewWriterWithOptions(options)
	time.Sleep(10 * time.Millisecond)
	testWriter3.Close()

	assert.IsType(t, &writer{}, testWriter3)
	assert.Equal(t, []string{"started", "stopped"}, logger.InfoLines)
	assert.Equal(t, []string{}, logger.ErrorLines)
}

func Test_run(t *testing.T) {
	tables := []struct {
		batch  func() batch.Batch
		logger []string
	}{
		{
			batch: func() batch.Batch {
				testBatch := &mocksBatch.Batch{}
				testBatch.On("Write", mock.Anything).Return(nil)
				testBatch.On("Reader").Return(&batch.BatchReader{})
				testBatch.On("Reset").Return()
				return testBatch
			},
			logger: []string{},
		},
		{
			batch: func() batch.Batch {
				testBatch := &mocksBatch.Batch{}
				testBatch.On("Write", mock.Anything).Return(errors.New("test"))
				testBatch.On("Reader").Return(&batch.BatchReader{})
				testBatch.On("Reset").Return()
				return testBatch
			},
			logger: []string{"batch.write: test"},
		},
		{
			batch: func() batch.Batch {
				count := 0
				testBatch := &mocksBatch.Batch{}
				testBatch.On("Write", mock.Anything).Return(func(_ []byte) error {
					if count > 0 {
						return errors.New("test")
					}
					count++
					return nil
				})
				testBatch.On("Reader").Return(&batch.BatchReader{})
				testBatch.On("Reset").Return()
				return testBatch
			},
			logger: []string{"batch.write: test"},
		},
	}

	for tt, table := range tables {
		logger := &mockLogger{
			ErrorLines: make([]string, 0),
		}

		testWriter := &writer{
			batch:        table.batch(),
			write:        make(chan []byte, 1),
			logger:       logger,
			sendInterval: 1 * time.Millisecond,
		}

		go func() {
			testWriter.write <- []byte("test")
			time.Sleep(10 * time.Millisecond)
			close(testWriter.write)
		}()
		testWriter.run()

		assert.Equalf(t, table.logger, logger.ErrorLines, "%d", tt)
	}
}

func Test_send(t *testing.T) {
	tables := []struct {
		batch       func() batch.Batch
		client      func() client.Client
		loggerInfo  []string
		loggerError []string
	}{
		{
			batch: func() batch.Batch {
				testBatch := &mocksBatch.Batch{}
				testBatch.On("Reader").Return(&batch.BatchReader{})
				testBatch.On("Reset").Return()
				return testBatch
			},
			client: func() client.Client {
				testClient := &mocksClient.Client{}
				return testClient
			},
			loggerInfo:  []string{},
			loggerError: []string{},
		},
		{
			batch: func() batch.Batch {
				testBatch := &mocksBatch.Batch{}
				testBatch.On("Reader").Return(&batch.BatchReader{
					Entries: 1,
					Size:    1,
				})
				testBatch.On("Reset").Return()
				return testBatch
			},
			client: func() client.Client {
				testClient := &mocksClient.Client{}
				testClient.On("Send", mock.Anything, mock.Anything).Return(nil, errors.New("test"))
				return testClient
			},
			loggerInfo:  []string{},
			loggerError: []string{"client.send: test"},
		},
		{
			batch: func() batch.Batch {
				testBatch := &mocksBatch.Batch{}
				testBatch.On("Reader").Return(&batch.BatchReader{
					Entries: 1,
					Size:    1,
				})
				testBatch.On("Reset").Return()
				return testBatch
			},
			client: func() client.Client {
				testClient := &mocksClient.Client{}
				testClient.On("Send", mock.Anything, mock.Anything).Return(&client.ClientResponse{
					StatusCode: 204,
				}, nil)
				return testClient
			},
			loggerInfo:  []string{"send batch: size: 1, entries: 1"},
			loggerError: []string{},
		},
		{
			batch: func() batch.Batch {
				testBatch := &mocksBatch.Batch{}
				testBatch.On("Reader").Return(&batch.BatchReader{
					Entries: 1,
					Size:    1,
				})
				testBatch.On("Reset").Return()
				return testBatch
			},
			client: func() client.Client {
				testClient := &mocksClient.Client{}
				testClient.On("Send", mock.Anything, mock.Anything).Return(&client.ClientResponse{
					StatusCode:    500,
					ResponseError: "test",
				}, nil)
				return testClient
			},
			loggerInfo:  []string{},
			loggerError: []string{"client.send: request_id: , status_code: 500, error: 'test'"},
		},
		{
			batch: func() batch.Batch {
				testBatch := &mocksBatch.Batch{}
				testBatch.On("Reader").Return(&batch.BatchReader{
					Entries: 1,
					Size:    1,
				})
				testBatch.On("Reset").Return()
				return testBatch
			},
			client: func() client.Client {
				testClient := &mocksClient.Client{}
				testClient.On("Send", mock.Anything, mock.Anything).Return(&client.ClientResponse{
					StatusCode: 500,
					Response:   "test",
				}, nil)
				return testClient
			},
			loggerInfo:  []string{},
			loggerError: []string{"client.send: request_id: , status_code: 500, response: 'test'"},
		},
	}

	for tt, table := range tables {
		logger := &mockLogger{
			InfoLines:  make([]string, 0),
			ErrorLines: make([]string, 0),
		}

		testWriter := &writer{
			batch:  table.batch(),
			client: table.client(),
			logger: logger,
		}

		testWriter.send()
		assert.Equalf(t, table.loggerInfo, logger.InfoLines, "%d", tt)
		assert.Equalf(t, table.loggerError, logger.ErrorLines, "%d", tt)
	}
}

func Test_Write(t *testing.T) {
	testClient := &writer{
		write: make(chan []byte, 3),
	}
	defer close(testClient.write)

	buffer := [3][]byte{
		[]byte("1"), []byte("2"), []byte("3"),
	}

	for i := 0; i < len(buffer); i++ {
		testClient.WriteLine(string(buffer[i]))
	}

	for i := 0; i < len(buffer); i++ {
		e := <-testClient.write
		assert.Equalf(t, buffer[i], e, "%d", i)
	}
}
