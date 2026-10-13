package service

import (
	"context"
	"fmt"
	"log/slog"
	"service-monitor/internal/incident/client"
	"service-monitor/internal/incident/repository"
	"service-monitor/pkg/metrics"
	"service-monitor/pkg/model"
)

type IncidentService struct {
	repo               repository.IncidentRepository
	notificationClient *client.NotificationClient
}

func NewIncidentService(repo repository.IncidentRepository, notifClient *client.NotificationClient) *IncidentService {
	return &IncidentService{
		repo:               repo,
		notificationClient: notifClient,
	}
}

func (s *IncidentService) HandleServiceEvent(ctx context.Context, event model.ServiceEvent) error {
	switch event.EventType {
	case model.EventServiceDown:
		return s.handleServiceDown(ctx, event)
	case model.EventServiceRecovered:
		return s.handleServiceRecovered(ctx, event)
	default:
		slog.Warn("unknown event type received", "type", event.EventType)
		return nil
	}
}

func (s *IncidentService) handleServiceDown(ctx context.Context, event model.ServiceEvent) error {
	inc := model.Incident{
		ServiceName:    event.ServiceName,
		EventType:      event.EventType,
		Status:         "open",
		IdempotencyKey: event.EventID,
	}

	id, created, err := s.repo.CreateIdempotent(ctx, inc)
	if err != nil {
		return fmt.Errorf("create incident in repo: %w", err)
	}

	if !created {
		slog.Info("duplicate event ignored due to idempotency key",
			"event_id", event.EventID,
			"service", event.ServiceName,
		)
		metrics.IncidentsTotal.WithLabelValues(event.ServiceName, "duplicate").Inc()
		return nil
	}

	metrics.IncidentsTotal.WithLabelValues(event.ServiceName, "created").Inc()
	slog.Info("created new incident", "incident_id", id, "service", event.ServiceName)

	// Call Notification Service via gRPC
	if s.notificationClient != nil {
		msg := fmt.Sprintf("INCIDENT: Service %s is DOWN! Endpoint: %s", event.ServiceName, event.Endpoint)
		if err := s.notificationClient.SendNotification(ctx, id, event.ServiceName, string(event.EventType), msg); err != nil {
			slog.Error("failed to dispatch notification via gRPC after retries", "incident_id", id, "error", err)
		} else {
			_ = s.repo.UpdateStatus(ctx, id, "notified")
		}
	}

	return nil
}

func (s *IncidentService) handleServiceRecovered(ctx context.Context, event model.ServiceEvent) error {
	resolvedInc, err := s.repo.ResolveLatest(ctx, event.ServiceName)
	if err != nil {
		return fmt.Errorf("resolve latest incident: %w", err)
	}

	if resolvedInc == nil {
		slog.Info("no active open incident found to resolve for service", "service", event.ServiceName)
		return nil
	}

	metrics.IncidentsTotal.WithLabelValues(event.ServiceName, "resolved").Inc()
	slog.Info("resolved incident", "incident_id", resolvedInc.ID, "service", event.ServiceName)

	// Notify recovery via gRPC
	if s.notificationClient != nil {
		msg := fmt.Sprintf("RESOLVED: Service %s has recovered and is healthy. Endpoint: %s", event.ServiceName, event.Endpoint)
		if err := s.notificationClient.SendNotification(ctx, resolvedInc.ID, event.ServiceName, string(event.EventType), msg); err != nil {
			slog.Error("failed to send recovery notification via gRPC", "incident_id", resolvedInc.ID, "error", err)
		}
	}

	return nil
}
