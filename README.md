# go-influxdb-writer

This library implements data writing in inflxudb 1.8+.

## Usage

Add import `github.com/a-kataev/go-influxdb-writer` to your source code and sync dependencies or directly edit the `go.mod` file.

## Options

The Writer uses set of options to configure behavior. These are available in the Options object creating a Writer instance using

```golang
w := writer.NewWriter("http://localhost:8086", "test-token", "test-bucket")
```

will use the default options.

To set different configuration values, e.g. to set send interval or entries limit, get default options and change what is needed:

```golang
w := writer.NewWriterWithOptions(writer.DefaultOptions().
    SetServerURL("http://localhost:8086").
    SetAuthToken("test-token").
    SetBucket("test-bucket").
    SetSendTimeout(30 * time.Second))
```

## Writer

Data are asynchronously written to the underlying buffer and they are automatically sent to a server when the size of the write buffer reaches the batch size, default 5000, or the flush interval, default 10s, times out.

Always use `Close()` method of the writer to stop all background processes.

## Example

```golang
package main

import (
	"fmt"
	"math/rand"
	"time"

	writer "github.com/a-kataev/go-influxdb-writer"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func main() {
	w := writer.NewWriter("http://localhost:8086", "test-token", "test-bucket")
	for i := 0; i < 100; i++ {
		p := write.NewPoint(
			"system",
			map[string]string{
				"id":       fmt.Sprintf("rack_%v", i%10),
				"vendor":   "AWS",
				"hostname": fmt.Sprintf("host_%v", i%100),
			},
			map[string]interface{}{
				"temperature": rand.Float64() * 80.0,
				"disk_free":   rand.Float64() * 1000.0,
				"disk_total":  (i/10 + 1) * 1000000,
				"mem_total":   (i/100 + 1) * 10000000,
				"mem_free":    rand.Uint64(),
			},
			time.Now())
		w.WriteLine(write.PointToLineProtocol(p, time.Nanosecond))
	}
	w.Close()
}
```
