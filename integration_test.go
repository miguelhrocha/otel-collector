package main_test

import (
	"context"
	"sync"
	"testing"
	"time"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"

	"github.com/miguelhrocha/otel-collector/config"
	"github.com/miguelhrocha/otel-collector/ingestor"
	"github.com/miguelhrocha/otel-collector/service"
	"github.com/stretchr/testify/assert"
)

func TestHighThroughput(t *testing.T) {
	cfg := config.Config{
		Addr:         ":4317",
		AttributeKey: "foo",
		Shards:       256,
		QueueSize:    10000,
		Workers:      4,
	}

	aggregator := ingestor.NewAggregator(cfg)
	deduplicator := ingestor.NewDeduplicator(cfg)
	ingestor := ingestor.NewIngestor(cfg, aggregator, deduplicator)

	svc := service.NewLogService(cfg, ingestor)

	var wg sync.WaitGroup

	numGoroutines := 100
	recordsPerRequest := 100

	start := time.Now()

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()

			for j := range recordsPerRequest {
				value := "bar" // Same value to test aggregation
				if j%3 == 0 {
					value = "qux"
				} else if j%5 == 0 {
					value = "baz"
				}

				req := createRequest(value)
				_, err := svc.Export(ctx, req)
				if err != nil {
					t.Errorf("Export failed: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	time.Sleep(2 * time.Second) // Wait for processing to complete

	snapshot := aggregator.Flush()

	gotQux := snapshot["qux"]
	gotBaz := snapshot["baz"]
	gotBar := snapshot["bar"]

	assert.Greater(t, gotQux, int64(0))
	assert.Greater(t, gotBaz, int64(0))
	assert.Greater(t, gotBar, int64(0))

	t.Logf("Processed %d records in %v", numGoroutines*recordsPerRequest, duration)
}

func createRequest(value string) *collogspb.ExportLogsServiceRequest {
	return &collogspb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				ScopeLogs: []*logspb.ScopeLogs{
					{
						LogRecords: []*logspb.LogRecord{
							{
								Attributes: []*commonpb.KeyValue{
									{
										Key: "foo",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_StringValue{StringValue: value},
										},
									},
								},
								TimeUnixNano:         uint64(time.Now().UnixNano()),
								ObservedTimeUnixNano: uint64(time.Now().UnixNano()),
								SeverityNumber:       logspb.SeverityNumber_SEVERITY_NUMBER_INFO,
								Body: &commonpb.AnyValue{
									Value: &commonpb.AnyValue_IntValue{
										IntValue: 42,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
