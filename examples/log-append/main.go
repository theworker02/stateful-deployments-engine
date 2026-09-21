package main

// log-append demo: sequential appends to a hot log file.
import (
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/sdk/go/sdesdk"
)

func main() {
	sink := &sdesdk.MemorySink{}
	c := sdesdk.New(sink, "log-append")
	var off int64
	for i := 0; i < 100; i++ {
		line := []byte(fmt.Sprintf("event=%d\n", i))
		_ = c.RecordWrite("/data/app.log", off, line, fmt.Sprintf("log-%d", i))
		off += int64(len(line))
	}
	fmt.Printf("log-append: %d lines, final offset=%d\n", len(sink.All), off)
}
