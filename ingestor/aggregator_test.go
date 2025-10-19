package ingestor_test

import (
	"testing"

	"github.com/miguelhrocha/otel-collector/config"
	"github.com/miguelhrocha/otel-collector/ingestor"
	"github.com/stretchr/testify/assert"
)

func TestAggregator(t *testing.T) {

	aggregator := ingestor.NewAggregator(config.Config{
		Shards: 10,
	})

	aggregator.Inc("foo")
	aggregator.Inc("bar")
	aggregator.Inc("foo")

	snapshot := aggregator.Flush()

	assert.Equal(t, int64(2), snapshot["foo"], "Aggregated count for 'foo' does not match")
	assert.Equal(t, int64(1), snapshot["bar"], "Aggregated count for 'bar' does not match")
}
