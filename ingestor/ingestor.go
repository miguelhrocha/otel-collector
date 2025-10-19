package ingestor

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/miguelhrocha/otel-collector/config"
	"github.com/miguelhrocha/otel-collector/metrics"
)

// Record represents a log record to be ingested.
//
// It is a subset of the fields from the OTLP LogRecord.
// It is used internally by the Ingestor to process incoming logs.
//
// All fields from this record are used for deduplication purposes.
type Record struct {
	// AttrValue is the value of the attribute used for aggregation.
	AttrValue string

	// TimeUnix is the timestamp of the log record in Unix nanoseconds.
	// Set from LogRecord.TimeUnixNano.
	TimeUnix uint64

	// ObsUnix is the observed timestamp of the log record in Unix nanoseconds.
	// Set from LogRecord.ObservedTimeUnixNano.
	ObsUnix uint64

	// Severity is the severity number of the log record.
	// Set from LogRecord.SeverityNumber.
	Severity int32

	// Body is the string representation of the log record's body.
	// Set from LogRecord.Body.
	Body string

	// TraceID is the trace ID associated with the log record.
	// Set from LogRecord.TraceId.
	TraceID string

	// SpanID is the span ID associated with the log record.
	// Set from LogRecord.SpanId.
	SpanID string
}

// Ingestor handles ingestion of log records.
//
// The Ingestor struct is responsible for receiving log records,
// processing them (deduplication and aggregation), and managing
// worker goroutines to handle the workload efficiently.
//
// Use NewIngestor to create a new Ingestor instance.
//
// Stop the Ingestor by calling the Stop method.
type Ingestor struct {
	q  chan Record
	wg sync.WaitGroup

	aggregator   *Aggregator
	deduplicator *Deduplicator

	stopped atomic.Bool
}

// NewIngestor creates a new Ingestor instance.
//
// It starts worker goroutines to process incoming log records.
//
// The number of workers is determined by the Workers field in the config.
// If Workers is less than or equal to 0, a default of 4 workers is used.
func NewIngestor(cfg config.Config, a *Aggregator, d *Deduplicator) *Ingestor {
	in := &Ingestor{
		q:            make(chan Record, cfg.QueueSize),
		aggregator:   a,
		deduplicator: d,
	}

	workers := cfg.Workers
	if workers <= 0 {
		workers = 4
	}

	for i := 0; i < workers; i++ {
		in.wg.Add(1)
		go func() {
			defer in.wg.Done()
			for r := range in.q {
				in.process(context.Background(), r)
			}
		}()
	}

	return in
}

// TryEnqueue attempts to enqueue a Record for processing.
//
// It returns true if the record was successfully enqueued,
// or false if the Ingestor is stopped or the queue is full.
//
// This method is non-blocking.
func (i *Ingestor) TryEnqueue(ctx context.Context, r Record) bool {
	if i.stopped.Load() {
		metrics.IngestDropped.Add(ctx, 1)
		return false
	}

	select {
	case i.q <- r:
		metrics.IngestTotal.Add(ctx, 1)
		return true
	default:
		metrics.IngestDropped.Add(ctx, 1)
		return false
	}
}

// Stop stops the Ingestor.
//
// It closes the internal queue and waits for all worker goroutines to finish processing.
func (i *Ingestor) Stop() {
	i.stopped.Store(true)
	close(i.q)
	i.wg.Wait()
}

func (i *Ingestor) process(ctx context.Context, r Record) {
	metrics.DeduplicationSeen.Add(ctx, 1)

	if !i.deduplicator.IsNew(r) {
		metrics.DeduplicationDuplicates.Add(ctx, 1)
		return
	}

	i.aggregator.Inc(r.AttrValue)
}
