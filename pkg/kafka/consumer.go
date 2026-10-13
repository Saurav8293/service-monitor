package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"service-monitor/pkg/metrics"
	"service-monitor/pkg/model"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer handles reading ServiceEvents from a Kafka topic within a consumer group.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer initializes a Kafka Reader for a consumer group.
func NewConsumer(broker string, topic string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	return &Consumer{
		reader: reader,
	}
}

// HandlerFunc defines the callback signature for consumed events.
type HandlerFunc func(ctx context.Context, event model.ServiceEvent) error

// StartConsume runs the consume loop until the context is canceled.
func (c *Consumer) StartConsume(ctx context.Context, handler HandlerFunc) error {
	slog.Info("starting Kafka consumer loop...")

	for {
		select {
		case <-ctx.Done():
			slog.Info("consumer loop context canceled, shutting down")
			return ctx.Err()
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("failed to fetch message from Kafka", "error", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var event model.ServiceEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			slog.Error("failed to deserialize service event", "error", err, "raw", string(msg.Value))
			// Commit bad message to avoid blocking partition
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		metrics.KafkaEventsConsumed.WithLabelValues(string(event.EventType)).Inc()

		if err := handler(ctx, event); err != nil {
			slog.Error("handler error processing event", "event_id", event.EventID, "error", err)
			// Depending on strategy, we can retry or commit; here we log and proceed
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("failed to commit message offset", "error", err)
		}
	}
}

// Close gracefully closes the Kafka reader.
func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
