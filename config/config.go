package config

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	// Addr is the address for the service to listen on.

	// Default is ":4317".
	Addr string `env:"ADDR, default=:4317"`

	// AttributeKey is the log attribute key to aggregate on.
	//
	// This key is required.
	AttributeKey string `env:"ATTRIBUTE_KEY, required"`

	// AggregationWindow is the time window for aggregation.
	//
	// Value should be a valid golang duration string (e.g., "10s", "1m", "1h").
	//
	// Default is 10s.
	AggregationWindow time.Duration `env:"AGGREGATION_WINDOW, default=10s"`

	// MaxReceiveMessageSize is the maximum gRPC receive message size in bytes.
	//
	// Default is 4MB.
	MaxReceiveMessageSize int `env:"MAX_RECEIVE_MESSAGE_SIZE, default=4194304"`

	// Shards is the number of shards to use for the log aggregator.
	//
	// More shards can improve concurrency and performance when processing
	// logs, especially under high load.
	//
	// However, too many shards can lead to increased memory usage
	// and overhead in managing them.
	//
	// The value should be a power of two for optimal performance.
	Shards int `env:"SHARDS, default=32"`

	// Workers is the number of worker goroutines to process logs.
	//
	// Each worker will read from the log processing queue and process logs concurrently.
	//
	// The value should be a fraction of the number of shards to avoid contention.
	Workers int `env:"WORKERS, default=4"`

	// QueueSize is the size of the log processing queue.
	//
	// A larger queue can help absorb bursts of incoming logs,
	// but will also consume more memory.
	//
	// Default is 1000.
	QueueSize int `env:"QUEUE_SIZE, default=1000"`

	// OtelEnabled indicates whether OpenTelemetry instrumentation is enabled.
	//
	// Default is true.
	OtelEnabled bool `env:"OTEL_ENABLED, default=true"`
}

// NewConfig creates a new Config instance
//
// It populates the config from environment variables.
func NewConfig(ctx context.Context) (Config, error) {
	var cfg Config

	if err := envconfig.Process(ctx, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
