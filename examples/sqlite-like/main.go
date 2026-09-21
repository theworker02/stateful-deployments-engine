package main

// sqlite-like demo: many small writes to a single hot file (DB page simulation).
import (
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/sdk/go/sdesdk"
)

func main() {
	sink := &sdesdk.MemorySink{}
	c := sdesdk.New(sink, "sqlite-like")
	page := make([]byte, 4096)
	for i := 0; i < 64; i++ {
		page[0] = byte(i)
		_ = c.RecordWrite("/data/app.db", int64(i*4096), page, fmt.Sprintf("page-%d", i))
	}
	fmt.Printf("sqlite-like: journaled %d page writes\n", len(sink.All))
}
