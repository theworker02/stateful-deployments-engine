package main

// Example: journal mutations through the public SDK (MemorySink).
//
//	go run ./sdk/go/examples/basic
import (
	"fmt"

	"github.com/theworker02/stateful-deployments-engine/sdk/go/sdesdk"
)

func main() {
	sink := &sdesdk.MemorySink{}
	client := sdesdk.New(sink, "example-deploy")
	_ = client.RecordWrite("/var/lib/app/db.sqlite", 0, []byte("SQLite format 3"), "op-1")
	_ = client.RecordWrite("/var/lib/app/db.sqlite", 100, []byte("page"), "op-2")
	fmt.Printf("journaled %d mutations\n", len(sink.All))
	for _, m := range sink.All {
		fmt.Printf("  %s %s hash=%s\n", m.Kind, m.Path, m.ContentHash[:12])
	}
}
