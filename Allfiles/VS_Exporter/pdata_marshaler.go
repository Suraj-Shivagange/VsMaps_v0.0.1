package vusmartmaps

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type pdataLogsMarshaler struct {
	marshaler plog.Marshaler
	encoding  string
}

func (p pdataLogsMarshaler) Marshal(ld plog.Logs, topic string) ([]*sarama.ProducerMessage, error) {
	bts, err := p.marshaler.MarshalLogs(ld)
	if err != nil {
		return nil, err
	}
	return []*sarama.ProducerMessage{
		{
			Topic: topic,
			Value: sarama.ByteEncoder(bts),
		},
	}, nil
}

func (p pdataLogsMarshaler) Encoding() string {
	return p.encoding
}

func newPdataLogsMarshaler(marshaler plog.Marshaler, encoding string) LogsMarshaler {
	return pdataLogsMarshaler{
		marshaler: marshaler,
		encoding:  encoding,
	}
}

type pdataMetricsMarshaler struct {
	marshaler pmetric.Marshaler
	encoding  string
}

func (p pdataMetricsMarshaler) Marshal(ld pmetric.Metrics, topic string) ([]*sarama.ProducerMessage, error) {
	bts, err := p.marshaler.MarshalMetrics(ld)
	if err != nil {
		return nil, err
	}
	return []*sarama.ProducerMessage{
		{
			Topic: topic,
			Value: sarama.ByteEncoder(bts),
		},
	}, nil
}

func (p pdataMetricsMarshaler) Encoding() string {
	return p.encoding
}

func newPdataMetricsMarshaler(marshaler pmetric.Marshaler, encoding string) MetricsMarshaler {
	return pdataMetricsMarshaler{
		marshaler: marshaler,
		encoding:  encoding,
	}
}

type pdataTracesMarshaler struct {
	marshaler ptrace.Marshaler
	encoding  string
}

type TracerawMarshaler struct {
}

func TracenewRawMarshaler() TracerawMarshaler {
	return TracerawMarshaler{}
}

func (p pdataTracesMarshaler) Marshal(td ptrace.Traces, topic string) ([]*sarama.ProducerMessage, error) {
	var messages []*sarama.ProducerMessage

	resourcedetails_2 := map[string]interface{}{}
	//scopeSpandetails_2 := map[string]interface{}{}

	spandetails2 := map[string]interface{}{}

	resourceSpans := td.ResourceSpans()
	for i := 0; i < resourceSpans.Len(); i++ {
		il := resourceSpans.At(i)

		scopeSpans := il.ScopeSpans()
		for j := 0; j < scopeSpans.Len(); j++ {

			spans := scopeSpans.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)

				if k == 0 {
					data2 := map[string]interface{}{
						"attributes": il.Resource().Attributes().AsRaw()}
					resourcedetails_2 = data2

				}

				data1 := map[string]interface{}{

					"scope": map[string]interface{}{
						"name":    il.ScopeSpans().AppendEmpty().Scope().Name(),
						"version": scopeSpans.AppendEmpty().Scope().Version(),
					},

					"span": map[string]interface{}{
						"ParentSpanID":      span.ParentSpanID().HexString(),
						"Name":              span.Name(),
						"startTimeUnixNano": span.StartTimestamp().AsTime().UnixNano(),
						"endTimeUnixNano":   span.EndTimestamp().AsTime().UnixNano(),
						"attributes":        span.Attributes().AsRaw(),
						"event":             span.Events(),
						"status":            span.Status().Code().String(),
					}}

				spandetails2 = data1
				resource := map[string]interface{}{
					"resourceSpans": map[string]interface{}{
						"resource":   resourcedetails_2,
						"scopespans": spandetails2,
					},
				}

				outputjson, err := json.Marshal(resource)
				if err != nil {
					return nil, err
				}
				if len(outputjson) == 0 {
					continue
				}

				messages = append(messages, &sarama.ProducerMessage{
					Topic: topic,
					Value: sarama.ByteEncoder(outputjson),
				})
			}

		}

	}
	return messages, nil
}

func (p pdataTracesMarshaler) Encoding() string {
	return p.encoding
}

func newPdataTracesMarshaler(marshaler ptrace.Marshaler, encoding string) TracesMarshaler {
	return pdataTracesMarshaler{
		marshaler: marshaler,
		encoding:  encoding,
	}
}

func (p pdataTracesMarshaler) TraceBodyAsBytes(value pcommon.Value) ([]byte, error) {
	switch value.Type() {
	case pcommon.ValueTypeStr:
		return p.interfaceAsBytes(value.Str())
	case pcommon.ValueTypeBytes:
		return value.Bytes().AsRaw(), nil
	case pcommon.ValueTypeBool:
		return p.interfaceAsBytes(value.Bool())
	case pcommon.ValueTypeDouble:
		return p.interfaceAsBytes(value.Double())
	case pcommon.ValueTypeInt:
		return p.interfaceAsBytes(value.Int())
	case pcommon.ValueTypeEmpty:
		return []byte{}, nil
	case pcommon.ValueTypeSlice:
		return p.interfaceAsBytes(value.Slice().AsRaw())
	case pcommon.ValueTypeMap:
		return p.interfaceAsBytes(value.Map().AsRaw())
	default:
		return nil, errUnsupported1
	}
}

func (p pdataTracesMarshaler) interfaceAsBytes(value interface{}) ([]byte, error) {
	if value == nil {
		return []byte{}, nil
	}
	res, err := json.Marshal(value)
	return res, err
}

func createKeyValuePairs(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}
