package metrics

import (
	"go.opentelemetry.io/otel/metric"
)

var (
	LogsReceivedCounter metric.Int64Counter
	LogsEnqueuedCounter metric.Int64Counter

	IngestTotal   metric.Int64Counter
	IngestDropped metric.Int64Counter

	DeduplicationSeen       metric.Int64Counter
	DeduplicationDuplicates metric.Int64Counter

	WindowFlushes       metric.Int64Counter
	WindowFlushDuration metric.Int64Histogram
	CountKeys           metric.Int64Gauge
)

// InitMetrics initializes all metrics used in the application.
//
// It should be called once at application startup.
func InitMetrics(meter metric.Meter) error {
	var err error

	LogsReceivedCounter, err = meter.Int64Counter("logs.received",
		metric.WithDescription("The number of logs received by the log-processor-backend"),
		metric.WithUnit("{log}"))

	if err != nil {
		return err
	}

	LogsEnqueuedCounter, err = meter.Int64Counter("logs.enqueued",
		metric.WithDescription("The number of logs enqueued for processing"),
		metric.WithUnit("{log}"))

	if err != nil {
		return err
	}

	IngestTotal, err = meter.Int64Counter("ingest.total",
		metric.WithDescription("The total number of logs ingested"),
		metric.WithUnit("{log}"))

	if err != nil {
		return err
	}

	IngestDropped, err = meter.Int64Counter("ingest.dropped",
		metric.WithDescription("The total number of logs dropped during ingestion"),
		metric.WithUnit("{log}"))

	if err != nil {
		return err
	}

	DeduplicationDuplicates, err = meter.Int64Counter("deduplication.total",
		metric.WithDescription("The total number of logs deduplicated"),
		metric.WithUnit("{log}"))

	if err != nil {
		return err
	}

	DeduplicationSeen, err = meter.Int64Counter("deduplication.seen",
		metric.WithDescription("The total number of unique logs seen for deduplication"),
		metric.WithUnit("{log}"))

	if err != nil {
		return err
	}

	WindowFlushes, err = meter.Int64Counter("window.flushes",
		metric.WithDescription("The total number of window flushes"),
		metric.WithUnit("{flush}"))

	if err != nil {
		return err
	}

	WindowFlushDuration, err = meter.Int64Histogram("window.flush.duration",
		metric.WithDescription("The duration of window flushes"),
		metric.WithUnit("ms"))

	if err != nil {
		return err
	}

	CountKeys, err = meter.Int64Gauge("count.keys",
		metric.WithDescription("The number of unique keys being tracked"),
		metric.WithUnit("{key}"))

	if err != nil {
		return err
	}

	return nil
}
