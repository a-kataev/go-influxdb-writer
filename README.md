# go-influxdb-writer

This library implements data writing in inflxudb 1.8+.

## Usage

Add import `github.com/a-kataev/go-influxdb-writer` to your source code and sync dependencies or edit directly the `go.mod` file.

## Options

The Writer uses a set of options to configure behavior. These are available in the object Options to create a Writer instance with default options using:

```golang
w := writer.NewWriter("http://localhost:8086", "test-token", "test-bucket")
```

To set different configuration values, e.g. to set the send interval or entries limit, get default options and change what is needed:

```golang
w := writer.NewWriterWithOptions(writer.DefaultOptions().
    SetServerURL("http://localhost:8086").
    SetAuthToken("test-token").
    SetBucket("test-bucket").
    SetSendTimeout(30 * time.Second))
```

## Writer

Data are asynchronously written to the underlying buffer and they are automatically sent to a server when the size of the write buffer reaches the batch size (default 3Mb), or the flush interval expires(default 10s).

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
## Limitations

Sending a batch is a blocking operation, the buffer is not available for writing while it is being sent.

It is necessary to take into account the sending interval and http-stimeout.

If the batch is not sent within the specified timeout, the data will not be saved and the buffer will be overwritten.
