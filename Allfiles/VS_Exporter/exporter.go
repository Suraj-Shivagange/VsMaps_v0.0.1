package vusmartmaps

import (
	"context"
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

var errUnrecognizedEncoding = fmt.Errorf("unrecognized encoding")

// VSTracesProducer uses sarama to produce trace messages to VS(kafka).
type VSTracesProducer struct {
	producer  sarama.SyncProducer
	topic     string
	marshaler TracesMarshaler
	logger    *zap.Logger
}

type VSErrors struct {
	count int
	err   string
}

func (ke VSErrors) Error() string {
	return fmt.Sprintf("Failed to deliver %d messages due to %s", ke.count, ke.err)
}

func (e *VSTracesProducer) tracesPusher(_ context.Context, td ptrace.Traces) error {
	var messages []*sarama.ProducerMessage
	messages, err := e.marshaler.Marshal(td, e.topic)
	if err != nil {
		return consumererror.NewPermanent(err)
	}
	err = e.producer.SendMessages(messages)
	if err != nil {
		var prodErr sarama.ProducerErrors
		if errors.As(err, &prodErr) {
			if len(prodErr) > 0 {
				return VSErrors{len(prodErr), prodErr[0].Err.Error()}
			}
		}
		return err
	}
	return nil
}

func (e *VSTracesProducer) Close(context.Context) error {
	return e.producer.Close()
}

// VSMetricsProducer uses sarama to produce metrics messages to VS (kafka)
type VSMetricsProducer struct {
	producer  sarama.SyncProducer
	topic     string
	marshaler MetricsMarshaler
	logger    *zap.Logger
}

func (e *VSMetricsProducer) metricsDataPusher(_ context.Context, md pmetric.Metrics) error {
	messages, err := e.marshaler.Marshal(md, e.topic)
	if err != nil {
		return consumererror.NewPermanent(err)
	}
	err = e.producer.SendMessages(messages)
	if err != nil {
		var prodErr sarama.ProducerErrors
		if errors.As(err, &prodErr) {
			if len(prodErr) > 0 {
				return VSErrors{len(prodErr), prodErr[0].Err.Error()}
			}
		}
		return err
	}
	return nil
}

func (e *VSMetricsProducer) Close(context.Context) error {
	return e.producer.Close()
}

// VSLogsProducer uses sarama to produce logs messages to VS (kafka)
type VSLogsProducer struct {
	producer  sarama.SyncProducer
	topic     string
	marshaler LogsMarshaler
	logger    *zap.Logger
}

func (e *VSLogsProducer) logsDataPusher(_ context.Context, ld plog.Logs) error {
	messages, err := e.marshaler.Marshal(ld, e.topic)
	if err != nil {
		return consumererror.NewPermanent(err)
	}
	err = e.producer.SendMessages(messages)
	if err != nil {
		var prodErr sarama.ProducerErrors
		if errors.As(err, &prodErr) {
			if len(prodErr) > 0 {
				return VSErrors{len(prodErr), prodErr[0].Err.Error()}
			}
		}
		return err
	}
	return nil
}

func (e *VSLogsProducer) Close(context.Context) error {
	return e.producer.Close()
}

func newSaramaProducer(config Config) (sarama.SyncProducer, error) {
	c := sarama.NewConfig()
	// These setting are required by the sarama.SyncProducer implementation.
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true
	c.Producer.RequiredAcks = config.Producer.RequiredAcks
	// Because sarama does not accept a Context for every message, set the Timeout here.
	c.Producer.Timeout = config.Timeout
	c.Metadata.Full = config.Metadata.Full
	c.Metadata.Retry.Max = config.Metadata.Retry.Max
	c.Metadata.Retry.Backoff = config.Metadata.Retry.Backoff
	c.Producer.MaxMessageBytes = config.Producer.MaxMessageBytes
	c.Producer.Flush.MaxMessages = config.Producer.FlushMaxMessages

	if config.ProtocolVersion != "" {
		version, err := sarama.ParseKafkaVersion(config.ProtocolVersion)
		if err != nil {
			return nil, err
		}
		c.Version = version
	}

	//if err := ConfigureAuthentication(config.Authentication, c); err != nil {
	//	return nil, err
	//}

	compression, err := saramaProducerCompressionCodec(config.Producer.Compression)
	if err != nil {
		return nil, err
	}
	c.Producer.Compression = compression

	producer, err := sarama.NewSyncProducer(config.Brokers, c)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

func newMetricsExporter(config Config, set component.ExporterCreateSettings, marshalers map[string]MetricsMarshaler) (*VSMetricsProducer, error) {
	marshaler := marshalers[config.Encoding]
	if marshaler == nil {
		return nil, errUnrecognizedEncoding
	}
	producer, err := newSaramaProducer(config)
	if err != nil {
		return nil, err
	}

	return &VSMetricsProducer{
		producer:  producer,
		topic:     config.Topic,
		marshaler: marshaler,
		logger:    set.Logger,
	}, nil

}

// newTracesExporter creates VS exporter.
func newTracesExporter(config Config, set component.ExporterCreateSettings, marshalers map[string]TracesMarshaler) (*VSTracesProducer, error) {
	marshaler := marshalers[config.Encoding]
	if marshaler == nil {
		return nil, errUnrecognizedEncoding
	}
	producer, err := newSaramaProducer(config)
	if err != nil {
		return nil, err
	}
	return &VSTracesProducer{
		producer:  producer,
		topic:     config.Topic,
		marshaler: marshaler,
		logger:    set.Logger,
	}, nil
}

func newLogsExporter(config Config, set component.ExporterCreateSettings, marshalers map[string]LogsMarshaler) (*VSLogsProducer, error) {
	marshaler := marshalers[config.Encoding]
	if marshaler == nil {
		return nil, errUnrecognizedEncoding
	}
	producer, err := newSaramaProducer(config)
	if err != nil {
		return nil, err
	}

	return &VSLogsProducer{
		producer:  producer,
		topic:     config.Topic,
		marshaler: marshaler,
		logger:    set.Logger,
	}, nil

}
