// Package sdesdk is the public Go SDK for applications to journal mutations
// into the Stateful Deployments Engine without importing internal packages.
//
// Evaluation / proprietary license applies — see repository LICENSE.
package sdesdk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// OpKind classifies a filesystem mutation.
type OpKind string

const (
	OpCreate OpKind = "create"
	OpWrite  OpKind = "write"
	OpRename OpKind = "rename"
	OpDelete OpKind = "delete"
	OpTrunc  OpKind = "truncate"
)

// Mutation is the public journal entry shape apps emit.
type Mutation struct {
	ClientOpID  string    `json:"client_op_id"`
	Kind        OpKind    `json:"kind"`
	Path        string    `json:"path"`
	DestPath    string    `json:"dest_path,omitempty"`
	Offset      int64     `json:"offset,omitempty"`
	Size        int64     `json:"size,omitempty"`
	ContentHash string    `json:"content_hash,omitempty"`
	Payload     []byte    `json:"payload,omitempty"`
	Timestamp   time.Time `json:"ts"`
}

// Sink receives mutations (implemented by engine agent or test doubles).
type Sink interface {
	Append(m Mutation) error
}

// Client is the app-facing journaler.
type Client struct {
	sink     Sink
	deployID string
	seq      uint64
	mu       sync.Mutex
}

// New creates a Client bound to a Sink.
func New(sink Sink, deployID string) *Client {
	return &Client{sink: sink, deployID: deployID}
}

// RecordWrite journals a write with optional payload hashing.
func (c *Client) RecordWrite(path string, offset int64, payload []byte, clientOpID string) error {
	m := Mutation{
		ClientOpID: clientOpID,
		Kind:       OpWrite,
		Path:       path,
		Offset:     offset,
		Size:       int64(len(payload)),
		Payload:    payload,
		Timestamp:  time.Now().UTC(),
	}
	if len(payload) > 0 {
		sum := sha256.Sum256(payload)
		m.ContentHash = hex.EncodeToString(sum[:])
	}
	if m.ClientOpID == "" {
		n := atomic.AddUint64(&c.seq, 1)
		m.ClientOpID = fmt.Sprintf("%s-%d", c.deployID, n)
	}
	return c.sink.Append(m)
}

// RecordDelete journals a delete.
func (c *Client) RecordDelete(path, clientOpID string) error {
	m := Mutation{
		ClientOpID: clientOpID,
		Kind:       OpDelete,
		Path:       path,
		Timestamp:  time.Now().UTC(),
	}
	if m.ClientOpID == "" {
		n := atomic.AddUint64(&c.seq, 1)
		m.ClientOpID = fmt.Sprintf("%s-%d", c.deployID, n)
	}
	return c.sink.Append(m)
}

// MemorySink stores mutations in memory (examples / tests).
type MemorySink struct {
	mu  sync.Mutex
	All []Mutation
}

func (s *MemorySink) Append(m Mutation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.All = append(s.All, m)
	return nil
}
