package consumer

import (
	"context"
	"log/slog"
	"service-monitor/internal/incident/service"
	"service-monitor/pkg/kafka"
	"service-monitor/pkg/model"
)

type EventConsumer struct {
	kafkaConsumer *kafka.Consumer
	svc           *service.IncidentService
}

func NewEventConsumer(kafkaConsumer *kafka.Consumer, svc *service.IncidentService) *EventConsumer {
	return &EventConsumer{
		kafkaConsumer: kafkaConsumer,
		svc:           svc,
	}
}

func (ec *EventConsumer) Start(ctx context.Context) error {
	handler := func(ctx context.Context, event model.ServiceEvent) error {
		return ec.svc.HandleServiceEvent(ctx, event)
	}

	slog.Info("starting Kafka event consumer in Incident Service...")
	return ec.kafkaConsumer.StartConsume(ctx, handler)
}

func (ec *EventConsumer) Close() error {
	return ec.kafkaConsumer.Close()
}
