package metrics

import (
	"go.opentelemetry.io/otel/metric"
)

var (
	LogsReceivedCounter metric.Int64Counter
)

func InitMetrics(meter metric.Meter) error {
	var err error

	LogsReceivedCounter, err = meter.Int64Counter("logs.received",
		metric.WithDescription("The number of logs received by the log-processor-backend"),
		metric.WithUnit("{log}"))

	return err
}
