package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/miguelhrocha/otel-collector/config"
	"github.com/miguelhrocha/otel-collector/ingestor"
	"github.com/miguelhrocha/otel-collector/metrics"
	internalotel "github.com/miguelhrocha/otel-collector/otel"
	"github.com/miguelhrocha/otel-collector/service"
)

const name = "miguelhrocha.com/otel-collector"

var (
	tracer = otel.Tracer(name)
	meter  = otel.Meter(name)
	logger = otelslog.NewLogger(name)
)

func init() {
	err := metrics.InitMetrics(meter)
	if err != nil {
		panic(err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	slog.SetDefault(logger)
	logger.Info("Starting application")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Set up OpenTelemetry.
	otelShutdown, err := internalotel.SetupSDK(ctx)
	if err != nil {
		return
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	cfg := config.NewConfig()

	aggregator := ingestor.NewAggregator(cfg)
	deduplicator := ingestor.NewDeduplicator(cfg)
	windowManager := ingestor.NewWindowManager(cfg, aggregator, deduplicator)
	ingestor := ingestor.NewIngestor(cfg, aggregator, deduplicator)

	windowManager.Start(ctx)

	slog.Debug("Starting listener", slog.String("listenAddr", cfg.Addr))
	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.MaxRecvMsgSize(cfg.MaxReceiveMessageSize),
		grpc.Creds(insecure.NewCredentials()),
	)

	healthSrver := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthSrver)
	collogspb.RegisterLogsServiceServer(grpcServer, service.NewLogService(cfg, ingestor))

	go func() {
		slog.Info("starting gRPC", "addr", cfg.Addr)
		if err := grpcServer.Serve(listener); err != nil {
			slog.Error("gRPC server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down gRPC server")
	grpcServer.GracefulStop()

	ingestor.Stop()
	windowManager.Stop()

	slog.Info("application stopped")

	return nil
}
