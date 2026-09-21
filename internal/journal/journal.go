// Package journal implements a durable append-only mutation log with
// checkpoints, duplicate detection, CRC trailers, truncation recovery,
// corruption detection, and resumable replay.
package journal

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/types"
)

const crcTrailerPrefix = "\t#crc32="

// Journal is a durable, ordered mutation log keyed by sequence number.
type Journal struct {
	mu         sync.Mutex
	path       string
	cpPath     string
	file       *os.File
	epoch      uint64
	fenceToken uint64
	nextSeq    uint64
	closed     bool
	seenOpIDs  map[string]uint64
	checkpoint types.JournalCheckpoint
	deployID   string
}

// Open creates or resumes a journal at dir/mutations.jsonl.
// Incomplete trailing lines (crash mid-write) are truncated before resume.
func Open(dir string, epoch uint64) (*Journal, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "mutations.jsonl")
	if err := recoverTruncation(path); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	j := &Journal{
		path:      path,
		cpPath:    filepath.Join(dir, "checkpoint.json"),
		file:      f,
		epoch:     epoch,
		nextSeq:   1,
		seenOpIDs: make(map[string]uint64),
	}
	if err := j.replayMeta(); err != nil {
		_ = f.Close()
		return nil, err
	}
	_ = j.loadCheckpoint()
	return j, nil
}

// recoverTruncation removes a partial last line that lacks a terminating newline.
// Mid-file corruption is left intact so VerifyIntegrity / Open still fail loudly.
func recoverTruncation(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if fi.Size() == 0 {
		return nil
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	const maxScan = 8 * 1024 * 1024
	size := fi.Size()
	start := size - maxScan
	if start < 0 {
		start = 0
	}
	buf := make([]byte, size-start)
	if _, err := f.ReadAt(buf, start); err != nil && err != io.EOF {
		return err
	}
	if len(buf) == 0 {
		return nil
	}
	if buf[len(buf)-1] == '\n' {
		return nil
	}
	// Find last complete record boundary.
	lastNL := -1
	for i := len(buf) - 1; i >= 0; i-- {
		if buf[i] == '\n' {
			lastNL = i
			break
		}
	}
	newSize := start
	if lastNL >= 0 {
		newSize = start + int64(lastNL) + 1
	} else {
		newSize = 0 // entire file is one incomplete line
	}
	if err := f.Truncate(newSize); err != nil {
		return fmt.Errorf("journal truncation recovery: %w", err)
	}
	return f.Sync()
}

func splitLineCRC(line []byte) (payload []byte, wantCRC uint32, hasCRC bool, err error) {
	s := string(line)
	idx := strings.LastIndex(s, crcTrailerPrefix)
	if idx < 0 {
		return line, 0, false, nil
	}
	payload = []byte(s[:idx])
	hexPart := s[idx+len(crcTrailerPrefix):]
	if len(hexPart) != 8 {
		return nil, 0, false, fmt.Errorf("malformed crc trailer")
	}
	var b [4]byte
	n, err := hex.Decode(b[:], []byte(hexPart))
	if err != nil || n != 4 {
		return nil, 0, false, fmt.Errorf("malformed crc hex: %w", err)
	}
	return payload, binary.BigEndian.Uint32(b[:]), true, nil
}

func appendCRCTrailer(jsonPayload []byte) []byte {
	sum := crc32.ChecksumIEEE(jsonPayload)
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], sum)
	trailer := crcTrailerPrefix + hex.EncodeToString(b[:])
	out := make([]byte, 0, len(jsonPayload)+len(trailer)+1)
	out = append(out, jsonPayload...)
	out = append(out, trailer...)
	out = append(out, '\n')
	return out
}

func parseMutationLine(line []byte, lineNo int) (types.Mutation, error) {
	payload, wantCRC, hasCRC, err := splitLineCRC(line)
	if err != nil {
		return types.Mutation{}, fmt.Errorf("corrupt journal crc at line %d: %w", lineNo, err)
	}
	if hasCRC {
		got := crc32.ChecksumIEEE(payload)
		if got != wantCRC {
			return types.Mutation{}, fmt.Errorf("corrupt journal crc mismatch at line %d", lineNo)
		}
	}
	var m types.Mutation
	if err := json.Unmarshal(payload, &m); err != nil {
		return types.Mutation{}, fmt.Errorf("corrupt journal at line %d: %w", lineNo, err)
	}
	if m.EntryChecksum != "" {
		expect := checksumEntry(m)
		if m.EntryChecksum != expect {
			return types.Mutation{}, fmt.Errorf("corrupt journal checksum at seq %d (line %d)", m.Seq, lineNo)
		}
	}
	return m, nil
}

func (j *Journal) replayMeta() error {
	f, err := os.Open(j.path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	var max uint64
	var prevSeq uint64
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		m, err := parseMutationLine(line, lineNo)
		if err != nil {
			return err
		}
		if prevSeq > 0 && m.Seq != prevSeq+1 && m.Seq != prevSeq {
			if m.Seq < prevSeq {
				return fmt.Errorf("journal sequence regression at line %d: %d < %d", lineNo, m.Seq, prevSeq)
			}
		}
		if m.Seq > max {
			max = m.Seq
		}
		if m.Epoch > j.epoch {
			j.epoch = m.Epoch
		}
		if m.FenceToken > j.fenceToken {
			j.fenceToken = m.FenceToken
		}
		if m.ClientOpID != "" {
			j.seenOpIDs[m.ClientOpID] = m.Seq
		}
		prevSeq = m.Seq
	}
	if err := sc.Err(); err != nil {
		return err
	}
	j.nextSeq = max + 1
	if j.nextSeq == 0 {
		j.nextSeq = 1
	}
	return nil
}

func (j *Journal) loadCheckpoint() error {
	b, err := os.ReadFile(j.cpPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(b, &j.checkpoint)
}

// SetDeploymentID stamps subsequent appends.
func (j *Journal) SetDeploymentID(id string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.deployID = id
}

// Epoch returns the current journal epoch.
func (j *Journal) Epoch() uint64 {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.epoch
}

// FenceToken returns the last observed fencing token.
func (j *Journal) FenceToken() uint64 {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.fenceToken
}

// NextSeq returns the next sequence number that will be assigned.
func (j *Journal) NextSeq() uint64 {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.nextSeq
}

// Checkpoint returns the last durable replay checkpoint.
func (j *Journal) Checkpoint() types.JournalCheckpoint {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.checkpoint
}

// SaveCheckpoint persists a resumable replay cursor.
func (j *Journal) SaveCheckpoint(seq uint64, note string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	cp := types.JournalCheckpoint{
		Seq:       seq,
		Epoch:     j.epoch,
		CreatedAt: time.Now().UTC(),
		Note:      note,
	}
	b, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(j.cpPath, b, 0o644); err != nil {
		return err
	}
	j.checkpoint = cp
	return nil
}

// Append records a mutation and returns the assigned sequence number.
// Duplicate ClientOpID returns the prior seq without appending (idempotent).
// Unrecognized op kinds are still persisted; callers must not silently discard them.
// Concurrent Append is serialized by j.mu; each record is CRC-trailed and Sync'd.
func (j *Journal) Append(m types.Mutation) (uint64, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		return 0, fmt.Errorf("journal closed")
	}
	if m.ClientOpID != "" {
		if seq, ok := j.seenOpIDs[m.ClientOpID]; ok {
			m.ReplayStatus = types.ReplayDuplicate
			return seq, nil
		}
	}
	if m.FenceToken != 0 && j.fenceToken != 0 && m.FenceToken < j.fenceToken {
		return 0, fmt.Errorf("stale fence token %d < %d", m.FenceToken, j.fenceToken)
	}
	m.Seq = j.nextSeq
	m.Epoch = j.epoch
	if m.FenceToken == 0 {
		m.FenceToken = j.fenceToken
	}
	if m.DeploymentID == "" {
		m.DeploymentID = j.deployID
	}
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now().UTC()
	}
	if m.ReplayStatus == "" {
		m.ReplayStatus = types.ReplayPending
	}
	m.EntryChecksum = checksumEntry(m)
	b, err := json.Marshal(m)
	if err != nil {
		return 0, err
	}
	if _, err := j.file.Write(appendCRCTrailer(b)); err != nil {
		return 0, err
	}
	if err := j.file.Sync(); err != nil {
		return 0, err
	}
	if m.ClientOpID != "" {
		j.seenOpIDs[m.ClientOpID] = m.Seq
	}
	j.nextSeq++
	return m.Seq, nil
}

// BumpEpoch advances the epoch (used after a successful cutover).
func (j *Journal) BumpEpoch() uint64 {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.epoch++
	return j.epoch
}

// SetEpoch forces the journal epoch (rollback).
func (j *Journal) SetEpoch(epoch uint64) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.epoch = epoch
}

// SetFenceToken installs a new fencing token (monotonic).
func (j *Journal) SetFenceToken(token uint64) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if token < j.fenceToken {
		return fmt.Errorf("fence token regression: %d < %d", token, j.fenceToken)
	}
	j.fenceToken = token
	return nil
}

// ReadFrom streams mutations with Seq >= fromSeq (inclusive).
func (j *Journal) ReadFrom(fromSeq uint64) ([]types.Mutation, error) {
	j.mu.Lock()
	path := j.path
	j.mu.Unlock()

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []types.Mutation
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		m, err := parseMutationLine(line, lineNo)
		if err != nil {
			return nil, err
		}
		if m.Seq >= fromSeq {
			out = append(out, m)
		}
	}
	return out, sc.Err()
}

// ReadRange returns mutations in [fromSeq, toSeq] inclusive.
func (j *Journal) ReadRange(fromSeq, toSeq uint64) ([]types.Mutation, error) {
	all, err := j.ReadFrom(fromSeq)
	if err != nil {
		return nil, err
	}
	var out []types.Mutation
	for _, m := range all {
		if m.Seq > toSeq {
			break
		}
		out = append(out, m)
	}
	return out, nil
}

// CountFrom returns how many mutations exist with Seq >= fromSeq.
func (j *Journal) CountFrom(fromSeq uint64) (uint64, error) {
	all, err := j.ReadFrom(fromSeq)
	if err != nil {
		return 0, err
	}
	return uint64(len(all)), nil
}

// VerifyIntegrity scans the full journal for corruption.
func (j *Journal) VerifyIntegrity() error {
	_, err := j.ReadFrom(1)
	return err
}

// Close flushes and closes the journal file.
func (j *Journal) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		return nil
	}
	j.closed = true
	return j.file.Close()
}

// Path returns the on-disk journal path.
func (j *Journal) Path() string { return j.path }

func checksumEntry(m types.Mutation) string {
	cp := m
	cp.EntryChecksum = ""
	b, _ := json.Marshal(cp)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
