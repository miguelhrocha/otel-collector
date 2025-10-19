package ingestor_test

import (
	"testing"

	"github.com/miguelhrocha/otel-collector/config"
	"github.com/miguelhrocha/otel-collector/ingestor"
	"github.com/stretchr/testify/assert"
)

func TestDeduplicator(t *testing.T) {
	deduplicator := ingestor.NewDeduplicator(config.Config{
		Shards: 10,
	})

	record := ingestor.Record{
		AttrValue: "test",
		TimeUnix:  1625079600,
		ObsUnix:   1625079601,
		Severity:  1,
		Body:      "This is a test log",
		TraceID:   "trace-id-123",
		SpanID:    "span-id-456",
	}

	isNew := deduplicator.IsNew(record)
	assert.True(t, isNew, "Expected the record to be new")

	isNew = deduplicator.IsNew(record)
	assert.False(t, isNew, "Expected the record to be a duplicate")

	deduplicator.Reset()

	isNew = deduplicator.IsNew(record)
	assert.True(t, isNew, "Expected the record to be new after reset")
}
