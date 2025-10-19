//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := collogspb.NewLogsServiceClient(conn)

	log.Println("Starting to send logs...")

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	values := []string{"bar", "qux", "baz", "corge"}

	for range ticker.C {
		batchSize := rand.Intn(50) + 10 // 10-60 records per batch
		logRecords := make([]*logspb.LogRecord, batchSize)

		for i := range batchSize {
			value := values[rand.Intn(len(values))]

			var attributes []*commonpb.KeyValue
			if rand.Float32() > 0.1 { // 90% chance of having the attribute
				attributes = []*commonpb.KeyValue{
					{
						Key: "foo",
						Value: &commonpb.AnyValue{
							Value: &commonpb.AnyValue_StringValue{StringValue: value},
						},
					},
					{
						Key: "timestamp",
						Value: &commonpb.AnyValue{
							Value: &commonpb.AnyValue_IntValue{IntValue: time.Now().Unix()},
						},
					},
				}
			} else {
				attributes = []*commonpb.KeyValue{
					{
						Key: "other",
						Value: &commonpb.AnyValue{
							Value: &commonpb.AnyValue_StringValue{StringValue: "value"},
						},
					},
				}
			}

			logRecords[i] = &logspb.LogRecord{
				TimeUnixNano: uint64(time.Now().UnixNano()),
				Body: &commonpb.AnyValue{
					Value: &commonpb.AnyValue_StringValue{
						StringValue: fmt.Sprintf("Log message %d", i),
					},
				},
				Attributes:     attributes,
				SeverityNumber: logspb.SeverityNumber(rand.Intn(24) + 1),
			}
		}

		request := &collogspb.ExportLogsServiceRequest{
			ResourceLogs: []*logspb.ResourceLogs{
				{
					Resource: &resourcepb.Resource{
						Attributes: []*commonpb.KeyValue{
							{
								Key: "service.name",
								Value: &commonpb.AnyValue{
									Value: &commonpb.AnyValue_StringValue{StringValue: "example-service"},
								},
							},
							{
								Key: "service.version",
								Value: &commonpb.AnyValue{
									Value: &commonpb.AnyValue_StringValue{StringValue: "1.0.0"},
								},
							},
						},
					},
					ScopeLogs: []*logspb.ScopeLogs{
						{
							Scope: &commonpb.InstrumentationScope{
								Name:    "example-scope",
								Version: "0.1.0",
							},
							LogRecords: logRecords,
						},
					},
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := client.Export(ctx, request)
		cancel()

		if err != nil {
			log.Printf("Failed to export logs: %v", err)
		} else {
			log.Printf("Sent batch of %d log records", batchSize)
		}
	}
}
