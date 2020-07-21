package batcher

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	o := &Options{
		BatchSize: 200,
	}

	b := New(o)
	assert.IsType(t, &batch{}, b)

	bb, _ := b.(*batch)
	assert.Equal(t, o.BatchSize, uint32(bb.buffer.Cap()))
}

func Test_Write(t *testing.T) {
	b := New(&Options{
		LinesLimit: 3,
		BatchSize:  10,
	})

	tables := []struct {
		lines     []string
		returnErr error
	}{
		{
			lines:     []string{"1", "2", "3"},
			returnErr: ErrLimitExceeded,
		},
		{
			lines:     []string{"4444", "666666"},
			returnErr: ErrLimitExceeded,
		},
		{
			lines:     []string{"11", "22"},
			returnErr: nil,
		},
	}

	for tt, table := range tables {
		if len(table.lines) < 2 {
			t.Error()
		}

		b.Reset()

		n := len(table.lines) - 1

		for i := 0; i < n; i++ {
			err := b.Write(table.lines[i])
			assert.Nilf(t, err, "%d %d - %s", tt, i, table.lines[i])
		}

		err := b.Write(table.lines[n])
		assert.Equalf(t, table.returnErr, err, "%d %d - %s", tt, n, table.lines[n])
	}
}

func Test_Reader(t *testing.T) {
	b := New(&Options{
		LinesLimit: 3,
		BatchSize:  15,
	})

	reader, size, count := b.Reader()
	assert.Nil(t, reader)
	assert.Zero(t, size)
	assert.Zero(t, count)

	lines := []string{"test1", "test2"}
	var joinLine string
	for _, line := range lines {
		err := b.Write(line)
		assert.Nil(t, err)
		joinLine += line + "\n"
	}

	reader, size, count = b.Reader()
	line, err := ioutil.ReadAll(reader)
	assert.Equal(t, joinLine, string(line))
	assert.Nil(t, err)
	assert.Equal(t, uint32(len(joinLine)), size)
	assert.Equal(t, uint32(2), count)
}
