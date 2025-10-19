package config

import (
	"flag"
	"time"
)

type Config struct {
	Addr              string
	AttributeKey      string
	AggregationWindow time.Duration

	MaxReceiveMessageSize int

	// Shards is the number of shards to use for the log aggregator.
	//
	// The value should be a power of two for optimal performance.
	Shards int

	// Workers is the number of worker goroutines to process logs.
	//
	// Each worker will read from the log processing queue and process logs concurrently.
	//
	// The value should be a fraction of the number of shards to avoid contention.
	Workers   int
	QueueSize int
}

func NewConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.Addr, "addr", ":4317", "The address the OTLP log processor backend will listen on (Default is 4317).")
	flag.StringVar(&cfg.AttributeKey, "attributeKey", "foo", "The log attribute key to aggregate on.")
	flag.IntVar(&cfg.MaxReceiveMessageSize, "maxReceiveMessageSize", 16777216, "The maximum receive message size in bytes.")
	flag.DurationVar(&cfg.AggregationWindow, "aggregationWindow", 10*time.Second, "The time window for aggregation.")
	flag.IntVar(&cfg.Workers, "workers", 4, "The number of worker goroutines to process logs.")
	flag.IntVar(&cfg.QueueSize, "queueSize", 1000, "The size of the log processing queue.")
	flag.IntVar(&cfg.Shards, "shards", 32, "The number of shards for the aggregator.")

	flag.Parse()

	return cfg
}
