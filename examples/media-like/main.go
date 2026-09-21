package main

// media-like demo: rare large object writes + frequent tiny metadata updates.
import (
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/sdk/go/sdesdk"
)

func main() {
	sink := &sdesdk.MemorySink{}
	c := sdesdk.New(sink, "media-like")
	blob := make([]byte, 256*1024)
	_ = c.RecordWrite("/data/media/video-001.bin", 0, blob, "upload-1")
	for i := 0; i < 20; i++ {
		_ = c.RecordWrite("/data/media/index.json", 0, []byte(fmt.Sprintf(`{"n":%d}`, i)), fmt.Sprintf("meta-%d", i))
	}
	fmt.Printf("media-like: %d mutations (1 large + 20 meta)\n", len(sink.All))
}
