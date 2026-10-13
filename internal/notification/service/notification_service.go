package service

import (
	"context"
	"fmt"
	"log/slog"
	"service-monitor/internal/notification/repository"
	"service-monitor/pkg/metrics"
	"service-monitor/pkg/model"
	"time"

	"github.com/google/uuid"
)

type NotificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{
		repo: repo,
	}
}

func (s *NotificationService) ProcessNotification(ctx context.Context, incidentID int64, serviceName, eventType, message string) (string, error) {
	notificationID := uuid.NewString()

	slog.Warn("DISPATCHING NOTIFICATION ALERT",
		"notification_id", notificationID,
		"incident_id", incidentID,
		"service_name", serviceName,
		"event_type", eventType,
		"message", message,
	)

	notif := model.Notification{
		IncidentID: incidentID,
		Channel:    "log",
		Message:    fmt.Sprintf("[%s] %s: %s", eventType, serviceName, message),
		SentAt:     time.Now(),
	}

	if _, err := s.repo.Create(ctx, notif); err != nil {
		metrics.NotificationsTotal.WithLabelValues("log", "failed").Inc()
		slog.Error("failed to record notification in database", "error", err)
		return "", err
	}

	metrics.NotificationsTotal.WithLabelValues("log", "sent").Inc()
	return notificationID, nil
}
