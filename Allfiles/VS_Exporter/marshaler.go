package vusmartmaps

import (
	"github.com/Shopify/sarama"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// TracesMarshaler marshals traces into Message array.
type TracesMarshaler interface {

	// Marshal serializes spans into sarama's ProducerMessages
	Marshal(traces ptrace.Traces, topic string) ([]*sarama.ProducerMessage, error)

	// Encoding returns encoding name
	Encoding() string
}

type NewTracesMarshaler interface {

	// Marshal serializes spans into sarama's ProducerMessages
	Marshal(traces ptrace.Traces, topic string) ([]*sarama.ProducerMessage, error)

	// Encoding returns encoding name
	Encoding() string
}

// MetricsMarshaler marshals metrics into Message array
type MetricsMarshaler interface {
	// Marshal serializes metrics into sarama's ProducerMessages
	Marshal(metrics pmetric.Metrics, topic string) ([]*sarama.ProducerMessage, error)

	// Encoding returns encoding name
	Encoding() string
}

// LogsMarshaler marshals logs into Message array
type LogsMarshaler interface {
	// Marshal serializes logs into sarama's ProducerMessages
	Marshal(logs plog.Logs, topic string) ([]*sarama.ProducerMessage, error)

	// Encoding returns encoding name
	Encoding() string
}

// tracesMarshalers returns map of supported encodings with TracesMarshaler.
func tracesMarshalers() map[string]TracesMarshaler {
	otlpPb := newPdataTracesMarshaler(ptrace.NewProtoMarshaler(), defaultEncoding)
	otlpJSON := newPdataTracesMarshaler(ptrace.NewJSONMarshaler(), "otlp_json")
	return map[string]TracesMarshaler{
		otlpPb.Encoding():   otlpPb,
		otlpJSON.Encoding(): otlpJSON,
		//jaegerProto.Encoding(): jaegerProto,
		//jaegerJSON.Encoding():  jaegerJSON,
	}
}

// metricsMarshalers returns map of supported encodings and MetricsMarshaler
func metricsMarshalers() map[string]MetricsMarshaler {
	otlpPb := newPdataMetricsMarshaler(pmetric.NewProtoMarshaler(), defaultEncoding)
	otlpJSON := newPdataMetricsMarshaler(pmetric.NewJSONMarshaler(), "otlp_json")
	return map[string]MetricsMarshaler{
		otlpPb.Encoding():   otlpPb,
		otlpJSON.Encoding(): otlpJSON,
	}
}

// logsMarshalers returns map of supported encodings and LogsMarshaler
func logsMarshalers() map[string]LogsMarshaler {
	otlpPb := newPdataLogsMarshaler(plog.NewProtoMarshaler(), defaultEncoding)
	otlpJSON := newPdataLogsMarshaler(plog.NewJSONMarshaler(), "otlp_json")
	raw := newRawMarshaler()
	return map[string]LogsMarshaler{
		otlpPb.Encoding():   otlpPb,
		otlpJSON.Encoding(): otlpJSON,
		raw.Encoding():      raw, // raw is taken from the raw_marshaler.go
	}
}
