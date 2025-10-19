package service

import (
	"context"
	"log/slog"

	"github.com/miguelhrocha/otel-collector/config"
	"github.com/miguelhrocha/otel-collector/metrics"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

type LogsServiceServer struct {
	addr string

	collogspb.UnimplementedLogsServiceServer
}

func NewLogService(cfg config.Config) collogspb.LogsServiceServer {
	s := &LogsServiceServer{addr: cfg.Addr}
	return s
}

func (l *LogsServiceServer) Export(ctx context.Context, request *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	slog.DebugContext(ctx, "Received ExportLogsServiceRequest")
	metrics.LogsReceivedCounter.Add(ctx, 1)

	// Do something with the logs

	return &collogspb.ExportLogsServiceResponse{}, nil
}
