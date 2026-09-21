package sdesdk

import "testing"

func TestClientMemorySink(t *testing.T) {
	sink := &MemorySink{}
	c := New(sink, "deploy-1")
	if err := c.RecordWrite("/data/x", 0, []byte("hi"), ""); err != nil {
		t.Fatal(err)
	}
	if err := c.RecordDelete("/data/y", "op-2"); err != nil {
		t.Fatal(err)
	}
	if len(sink.All) != 2 || sink.All[0].ContentHash == "" {
		t.Fatalf("%+v", sink.All)
	}
}
