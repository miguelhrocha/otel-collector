package service

import (
	"context"
	"log/slog"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"

	"github.com/miguelhrocha/otel-collector/config"
	"github.com/miguelhrocha/otel-collector/ingestor"
	"github.com/miguelhrocha/otel-collector/metrics"
	"github.com/miguelhrocha/otel-collector/otel"
)

type LogsServiceServer struct {
	addr               string
	attributeExtractor otel.AttributeExtractor
	ingestor           *ingestor.Ingestor

	collogspb.UnimplementedLogsServiceServer
}

func NewLogService(cfg config.Config, ingestor *ingestor.Ingestor) collogspb.LogsServiceServer {
	s := &LogsServiceServer{
		addr:               cfg.Addr,
		attributeExtractor: *otel.NewAttributeExtractor(cfg.AttributeKey),
		ingestor:           ingestor,
	}
	return s
}

func (l *LogsServiceServer) Export(ctx context.Context, request *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	slog.DebugContext(ctx, "Received ExportLogsServiceRequest")
	metrics.LogsReceivedCounter.Add(ctx, 1)

	if request == nil {
		slog.WarnContext(ctx, "Received nil request")
		return &collogspb.ExportLogsServiceResponse{}, nil
	}

	for _, resourceLog := range request.GetResourceLogs() {
		resource := resourceLog.GetResource()
		for _, scopeLog := range resourceLog.GetScopeLogs() {
			scope := scopeLog.GetScope()
			for _, logRecord := range scopeLog.GetLogRecords() {
				attributeValue := l.attributeExtractor.Extract(logRecord, scope, resource)
				if attributeValue == "" {
					attributeValue = "unknown"
				}

				r := ingestor.Record{
					AttrValue: attributeValue,
					TimeUnix:  logRecord.GetTimeUnixNano(),
					ObsUnix:   logRecord.GetObservedTimeUnixNano(),
					Severity:  int32(logRecord.GetSeverityNumber()),
					Body:      bodyToString(logRecord.GetBody()),
					TraceID:   string(logRecord.GetTraceId()),
					SpanID:    string(logRecord.GetSpanId()),
				}

				if ok := l.ingestor.TryEnqueue(r); ok {
					// metrics.LogsEnqueuedCounter.Add(ctx, 1)
				}
			}
		}
	}

	return &collogspb.ExportLogsServiceResponse{}, nil
}

func bodyToString(v *common.AnyValue) string {
	if v == nil {
		return ""
	}
	return otel.AnyValueAsString(v)
}
