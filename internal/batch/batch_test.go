package batch

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	options := &Options{
		BufferSize: 123,
	}

	testBatch := New(options)
	assert.IsType(t, &batch{}, testBatch)

	batchStruct, _ := testBatch.(*batch)
	assert.Equal(t, options.BufferSize, uint64(batchStruct.buffer.Cap()))
}

func Test_Write(t *testing.T) {
	testBatch := New(&Options{
		EntriesLimit: 3,
		BufferSize:   10,
	})

	tables := []struct {
		lines []string
		err   error
	}{
		{
			lines: []string{"1", "2", "3"},
			err:   ErrLimitExceeded,
		},
		{
			lines: []string{"4444", "666666"},
			err:   ErrSizeExceeded,
		},
		{
			lines: []string{"11", "22"},
			err:   nil,
		},
	}

	for tt, table := range tables {
		if len(table.lines) < 2 {
			t.Errorf("%d minimum size of lines must be at least 2", tt)
		}

		testBatch.Reset()

		size := len(table.lines) - 1

		for i := 0; i < size; i++ {
			err := testBatch.Write([]byte(table.lines[i]))
			assert.Nilf(t, err, "%d %d - %s", tt, i, table.lines[i])
		}

		err := testBatch.Write([]byte(table.lines[size]))
		assert.Equalf(t, table.err, err, "%d %d - %s", tt, size, table.lines[size])
	}
}

func Test_Reader(t *testing.T) {
	testBatch := New(&Options{
		EntriesLimit: 3,
		BufferSize:   15,
	})

	reader := testBatch.Reader()
	readerBuffer, err := ioutil.ReadAll(reader.Reader)
	assert.Equal(t, "", string(readerBuffer))
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), reader.Size)
	assert.Equal(t, uint64(0), reader.Entries)

	lines := []string{"test1", "test2"}
	var buffer string
	for _, line := range lines {
		err := testBatch.Write([]byte(line))
		assert.Nil(t, err)
		buffer += line + "\n"
	}

	reader = testBatch.Reader()
	readerBuffer, err = ioutil.ReadAll(reader.Reader)
	assert.Equal(t, buffer, string(readerBuffer))
	assert.Nil(t, err)
	assert.Equal(t, uint64(len(buffer)), reader.Size)
	assert.Equal(t, uint64(len(lines)), reader.Entries)
}
