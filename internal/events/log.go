// Package events provides a structured event log for transactional deploy UX.
package events

import (
	"encoding/json"
	"sync"
	"time"
)

// Level classifies event severity for operator UIs.
type Level string

const (
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelOK    Level = "ok"
)

// Kind is a stable event type string.
type Kind string

const (
	KindPhaseEnter      Kind = "phase.enter"
	KindPhaseExit       Kind = "phase.exit"
	KindMutationSynced  Kind = "mutation.synced"
	KindVerifyResult    Kind = "verify.result"
	KindSafetyGate      Kind = "safety.gate"
	KindBarrierHold     Kind = "barrier.hold"
	KindBarrierRelease  Kind = "barrier.release"
	KindTrafficTransfer Kind = "traffic.transfer"
	KindReceiptSealed   Kind = "receipt.sealed"
	KindRollback        Kind = "rollback"
	KindChaosInject     Kind = "chaos.inject"
)

// Event is one structured log line for the deploy timeline.
type Event struct {
	At       time.Time              `json:"at"`
	DeployID string                 `json:"deploy_id"`
	Kind     Kind                   `json:"kind"`
	Level    Level                  `json:"level"`
	Message  string                 `json:"message"`
	Phase    string                 `json:"phase,omitempty"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
}

// Log is an in-memory append-only event buffer (demo / CLI UX).
type Log struct {
	mu     sync.Mutex
	events []Event
}

// Append records an event.
func (l *Log) Append(e Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	l.events = append(l.events, e)
}

// Info is a convenience helper.
func (l *Log) Info(deployID string, kind Kind, msg string, fields map[string]interface{}) {
	l.Append(Event{DeployID: deployID, Kind: kind, Level: LevelInfo, Message: msg, Fields: fields})
}

// OK records a success milestone (checkmarks in CLI).
func (l *Log) OK(deployID string, kind Kind, msg string, fields map[string]interface{}) {
	l.Append(Event{DeployID: deployID, Kind: kind, Level: LevelOK, Message: msg, Fields: fields})
}

// Snapshot returns a copy of events.
func (l *Log) Snapshot() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

// JSON renders the log for receipts / APIs.
func (l *Log) JSON() ([]byte, error) {
	return json.MarshalIndent(l.Snapshot(), "", "  ")
}

// RenderHuman produces CLI-friendly lines.
func RenderHuman(evts []Event) []string {
	var lines []string
	for _, e := range evts {
		prefix := "·"
		switch e.Level {
		case LevelOK:
			prefix = "✓"
		case LevelWarn:
			prefix = "!"
		case LevelError:
			prefix = "✗"
		}
		lines = append(lines, prefix+" "+e.Message)
	}
	return lines
}
