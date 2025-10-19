package otel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/miguelhrocha/otel-collector/otel"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

func TestExtractor(t *testing.T) {
	extractor := otel.NewAttributeExtractor("foo")

	t.Run("extracts attribute from log record", func(t *testing.T) {
		logRecord := &logspb.LogRecord{
			Attributes: []*commonpb.KeyValue{
				{
					Key: "foo",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "bar"},
					},
				},
			},
		}

		result := extractor.Extract(logRecord, nil, nil)
		assert.Equal(t, "bar", result)
	})

	t.Run("extracts attribute from scope", func(t *testing.T) {
		logRecord := &logspb.LogRecord{}
		scope := &commonpb.InstrumentationScope{
			Attributes: []*commonpb.KeyValue{
				{
					Key: "foo",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "scope-value"},
					},
				},
			},
		}

		result := extractor.Extract(logRecord, scope, nil)
		assert.Equal(t, "scope-value", result)
	})

	t.Run("extracts attribut from resource", func(t *testing.T) {
		logRecord := &logspb.LogRecord{}
		resource := &resourcepb.Resource{
			Attributes: []*commonpb.KeyValue{
				{
					Key: "foo",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "resource-value"},
					},
				},
			},
		}

		result := extractor.Extract(logRecord, nil, resource)
		assert.Equal(t, "resource-value", result)
	})

	t.Run("extracts respect log record > scope > resource priority", func(t *testing.T) {
		logRecord := &logspb.LogRecord{
			Attributes: []*commonpb.KeyValue{
				{
					Key: "foo",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "log-value"},
					},
				},
			},
		}

		scope := &commonpb.InstrumentationScope{
			Attributes: []*commonpb.KeyValue{
				{
					Key: "foo",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "scope-value"},
					},
				},
			},
		}

		resource := &resourcepb.Resource{
			Attributes: []*commonpb.KeyValue{
				{
					Key: "foo",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "resource-value"},
					},
				},
			},
		}

		result := extractor.Extract(logRecord, scope, resource)

		assert.Equal(t, "log-value", result)
	})

	t.Run("not found key returns unknown string", func(t *testing.T) {
		logRecord := &logspb.LogRecord{
			Attributes: []*commonpb.KeyValue{
				{
					Key: "other",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "value"},
					},
				},
			},
		}

		result := extractor.Extract(logRecord, nil, nil)
		assert.Equal(t, "unknown", result)
	})

}
