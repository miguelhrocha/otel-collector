package ingestor

import (
	"sync"

	"github.com/segmentio/fasthash/fnv1"

	"github.com/miguelhrocha/otel-collector/config"
)

type Aggregator struct {
	shards []aggregatorShard
}

type aggregatorShard struct {
	// Chose a regular mutext because the aggregator expects a write-heavy workload
	// and sync.RWMutex would add unnecessary overhead.
	mu sync.Mutex

	// Chose a map + mutex intead of sync.Map because
	// the access pattern is mostly writes.
	data map[string]int64
}

// NewAggregator creates a new Aggregator instance with the amount
// of shards specified in the config's Shards field.
func NewAggregator(cfg config.Config) *Aggregator {
	s := make([]aggregatorShard, cfg.Shards)

	for i := range s {
		s[i] = aggregatorShard{
			data: make(map[string]int64),
		}
	}

	return &Aggregator{
		shards: s,
	}
}

// Inc increments the counter for the given key.
func (a *Aggregator) Inc(key string) {
	// Use FNV-1a hash to determine the shard for the given key.
	//
	// FNV-1a is a fast, non-cryptographic hash function that
	// provides a good distribution of hash values, minimizing
	// the chances of collisions and ensuring even load across shards.
	//
	// This design gives us a lock granularity of 1/shards,
	// which improves concurrency and throughput in write-heavy workloads.
	hash := fnv1.HashBytes64([]byte(key))
	shardKey := hash % uint64(len(a.shards))

	shard := &a.shards[shardKey]
	shard.mu.Lock()
	shard.data[key]++
	shard.mu.Unlock()
}

// Flush returns a snapshot of the current aggregated data
// and resets the internal state of the aggregator.
//
// This is used to periodically flush the aggregated data.
func (a *Aggregator) Flush() map[string]int64 {
	result := make(map[string]int64)

	for i := range a.shards {
		sh := &a.shards[i]
		sh.mu.Lock()
		for k, v := range sh.data {
			result[k] += v
		}
		sh.data = make(map[string]int64)
		sh.mu.Unlock()
	}

	return result
}
