package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"service-monitor/pkg/metrics"
	"service-monitor/pkg/model"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer handles publishing ServiceEvents to a Kafka topic.
type Producer struct {
	writer *kafka.Writer
	topic  string
}

// NewProducer creates a Kafka writer for the given broker and topic.
func NewProducer(broker string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	return &Producer{
		writer: writer,
		topic:  topic,
	}
}

// PublishEvent serializes and publishes a ServiceEvent to Kafka.
func (p *Producer) PublishEvent(ctx context.Context, event model.ServiceEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.ServiceName),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}

	metrics.KafkaEventsPublished.WithLabelValues(string(event.EventType)).Inc()
	slog.Info("published event to kafka",
		"topic", p.topic,
		"event_id", event.EventID,
		"event_type", event.EventType,
		"service_name", event.ServiceName,
	)

	return nil
}

// Close gracefully flushes and closes the Kafka writer.
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
