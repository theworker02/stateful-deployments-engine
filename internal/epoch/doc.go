// Package epoch provides durable fencing epochs and lease-aware epoch stores.
//
// An epoch is a monotonically increasing state-generation counter. After cutover
// or rollback, a new epoch fences stale writers so they cannot mutate newer
// state. This package complements internal/fence with store abstractions,
// comparison helpers, and persistence layouts suitable for multi-process use.
package epoch
