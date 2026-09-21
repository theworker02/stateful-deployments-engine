// Package fsm defines the deployment state-machine transition table, persistence
// helpers, and validation for the Stateful Deployments Engine.
//
// The authoritative phase constants live in internal/types. This package owns
// the legal transition graph, serialization of FSM snapshots, and recovery
// annotations used by coordinators without coupling to platform adapters.
package fsm
