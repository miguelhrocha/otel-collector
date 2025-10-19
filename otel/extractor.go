package otel

import (
	"fmt"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

const unknownValue = "unknown"

// AttributeExtractor extracts the configured
// AttributeKey from OTLP log records
type AttributeExtractor struct {
	attributeKey string
}

// NewAttributeExtractor creates a new AttributeExtractor
// with the given attribute key.
func NewAttributeExtractor(attributeKey string) *AttributeExtractor {
	return &AttributeExtractor{attributeKey}
}

// Extract retrieves the application-configured attributer value from the log record.
//
// The hierarchy for attribute extraction is as follows:
// 1. Log Record Attributes
// 2. Instrumentation Scope Attributes
// 3. Resource Attributes
// If the attribute is not found, an empty string is returned.
func (extractor *AttributeExtractor) Extract(
	logRecord *logspb.LogRecord,
	scope *commonpb.InstrumentationScope,
	resource *resourcepb.Resource,
) string {
	if value := extractor.findInAttributes(logRecord.GetAttributes()); value != "" {
		return value
	}

	if scope != nil {
		if value := extractor.findInAttributes(scope.GetAttributes()); value != "" {
			return value
		}
	}

	if resource != nil {
		if value := extractor.findInAttributes(resource.GetAttributes()); value != "" {
			return value
		}
	}

	return unknownValue
}

func (extractor *AttributeExtractor) findInAttributes(attributes []*commonpb.KeyValue) string {
	for _, attr := range attributes {
		if attr.GetKey() == extractor.attributeKey {
			return extractor.getValueAsString(attr.GetValue())
		}
	}
	return ""
}

func (extractor *AttributeExtractor) getValueAsString(value *commonpb.AnyValue) string {
	if value == nil {
		return unknownValue
	}

	switch v := value.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		if v.StringValue == "" {
			return unknownValue
		}
		return v.StringValue
	case *commonpb.AnyValue_IntValue:
		return fmt.Sprintf("%d", v.IntValue)
	case *commonpb.AnyValue_DoubleValue:
		return fmt.Sprintf("%g", v.DoubleValue)
	case *commonpb.AnyValue_BoolValue:
		return fmt.Sprintf("%t", v.BoolValue)
	case *commonpb.AnyValue_BytesValue:
		return string(v.BytesValue)
	default:
		// For complex types (arrays, maps), return unknown
		return unknownValue
	}
}
