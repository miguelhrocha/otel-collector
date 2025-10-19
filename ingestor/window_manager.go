package ingestor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/miguelhrocha/otel-collector/config"
	"github.com/miguelhrocha/otel-collector/metrics"
)

// WindowManager manages aggregation windows.
//
// It periodically flushes the current aggregation window and resets the deduplicator.
type WindowManager struct {
	aggregator     *Aggregator
	deduplicator   *Deduplicator
	windowDuration time.Duration
	attributeKey   string
	ticker         *time.Ticker
	stopCh         chan struct{}
	doneCh         chan struct{}
}

// NewWindowManager creates a new WindowManager instance.
func NewWindowManager(cfg config.Config, a *Aggregator, d *Deduplicator) *WindowManager {
	return &WindowManager{
		aggregator:     a,
		deduplicator:   d,
		windowDuration: cfg.AggregationWindow,
		attributeKey:   cfg.AttributeKey,
		ticker:         time.NewTicker(cfg.AggregationWindow),
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}
}

// Start begins the window management process.
//
// It starts a goroutine that flushes the aggregation window at regular intervals.
//
// Stop the window manager by calling the Stop method.
func (wm *WindowManager) Start(ctx context.Context) {
	wm.ticker = time.NewTicker(wm.windowDuration)

	slog.InfoContext(ctx, "Window manager started",
		slog.Duration("window_duration", wm.windowDuration),
		slog.String("attribute_key", wm.attributeKey))

	go wm.run(ctx)
}

func (wm *WindowManager) run(ctx context.Context) {
	defer close(wm.doneCh)
	defer wm.ticker.Stop()

	for {
		select {
		case <-wm.ticker.C:
			wm.flushWindow(ctx)
		case <-wm.stopCh:
			slog.InfoContext(ctx, "Window manager stopping, performing final flush")
			wm.flushWindow(ctx)
			return
		case <-ctx.Done():
			slog.InfoContext(ctx, "Window manager context done, performing final flush")
			wm.flushWindow(ctx)
			return
		}
	}
}

func (wm *WindowManager) flushWindow(ctx context.Context) {
	start := time.Now()

	snapshot := wm.aggregator.Flush()
	metrics.WindowFlushDuration.Record(ctx, time.Since(start).Milliseconds())
	metrics.WindowFlushes.Add(ctx, 1)
	metrics.CountKeys.Record(ctx, int64(len(snapshot)))

	if len(snapshot) == 0 {
		fmt.Println("aggregation window is empty")
		return
	}

	// startTime := time.Now()
	// slog.InfoContext(ctx, "Flushed aggregation window",
	// 	slog.Int("unique_keys", len(snapshot)),
	// 	slog.Duration("flush_duration", time.Since(startTime)))

	// print snapshot
	fmt.Println("aggregation window")
	for k, v := range snapshot {
		fmt.Printf("%s - %d\n", k, v)
	}
	fmt.Println("-----")

	wm.deduplicator.Reset()
}

// Stop stops the WindowManager.
//
// It performs a final flush of the aggregation window before stopping.
func (wm *WindowManager) Stop() {
	close(wm.stopCh)
	<-wm.doneCh
}
