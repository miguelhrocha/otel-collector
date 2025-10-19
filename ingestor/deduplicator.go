package ingestor

import (
	"encoding/binary"
	"sync"

	"github.com/miguelhrocha/otel-collector/config"
	"github.com/segmentio/fasthash/fnv1a"
)

var (
	capacityPerShard = 32 * 1024 // 32K entries per shard
)

// Deduplicator handles deduplication of incoming logs.
//
// Logs in OTEL format can sometimes be duplicated due to various reasons such as network issues or retries.
// The Deduplicator struct is responsible for identifying and removing these duplicate logs before further processing.
type Deduplicator struct {
	shards []shard
}

type shard struct {
	mu       sync.Mutex
	seen     map[uint64]struct{}
	capacity int
}

// NewDeduplicator creates a new Dedupe instance with the specified number of shards and capacity per shard.
func NewDeduplicator(cfg config.Config) *Deduplicator {
	s := make([]shard, cfg.Shards)
	for i := range s {
		s[i] = shard{
			seen:     make(map[uint64]struct{}, capacityPerShard),
			capacity: capacityPerShard,
		}
	}
	return &Deduplicator{
		shards: s,
	}
}

// IsNew checks if a given Record is new (not a duplicate).
//
// It computes the hash of the Record and checks if it has been seen before.
// If the Record is new, it adds its hash to the seen set and returns true.
// If it is a duplicate, it returns false.
func (d *Deduplicator) IsNew(r Record) bool {
	hash := hashRecord(r)
	key := int(hash % uint64(len(d.shards)))

	sh := &d.shards[key]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	if _, ok := sh.seen[hash]; ok {
		return false
	}

	sh.seen[hash] = struct{}{}
	return true
}

// Reset clears all seen records in the Dedupe instance.
func (d *Deduplicator) Reset() {
	for i := range d.shards {
		sh := &d.shards[i]
		sh.mu.Lock()
		sh.seen = make(map[uint64]struct{}, sh.capacity)
		sh.mu.Unlock()
	}
}

// hashRecord computes a hash for a given Record based on its significant fields.
//
// This function concatenates the relevant fields of the Record, separated by null bytes,
// and then computes a FNV-1a hash of the resulting byte slice.
//
// This hash is used to identify duplicate records efficiently.
func hashRecord(r Record) uint64 {
	parts := make([]byte, 0, 128)

	writeSep := func() {
		parts = append(parts, 0)
	}
	writeString := func(s string) {
		writeSep()
		parts = append(parts, s...)
	}
	writeUint64 := func(v uint64) {
		writeSep()
		var b [8]byte

		// Use LittleEndian to avoid string conversion issues across different architectures
		binary.LittleEndian.PutUint64(b[:], v)
		parts = append(parts, b[:]...)
	}
	writeInt32 := func(v int32) {
		writeSep()
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], uint32(v))
		parts = append(parts, b[:]...)
	}

	writeString(r.AttrValue)
	writeUint64(r.TimeUnix)
	writeUint64(r.ObsUnix)
	writeInt32(r.Severity)
	writeString(r.Body)
	writeString(r.TraceID)
	writeString(r.SpanID)

	return fnv1a.HashBytes64(parts)
}
