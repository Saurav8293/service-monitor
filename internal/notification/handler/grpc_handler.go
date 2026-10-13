package handler

import (
	"context"
	"service-monitor/internal/notification/service"
	"service-monitor/pkg/metrics"
	notificationv1 "service-monitor/pkg/proto/notification/v1"
	"time"
)

type GRPCHandler struct {
	notificationv1.UnimplementedNotificationServiceServer
	svc *service.NotificationService
}

func NewGRPCHandler(svc *service.NotificationService) *GRPCHandler {
	return &GRPCHandler{
		svc: svc,
	}
}

func (h *GRPCHandler) SendNotification(ctx context.Context, req *notificationv1.SendNotificationRequest) (*notificationv1.SendNotificationResponse, error) {
	notifID, err := h.svc.ProcessNotification(ctx, req.IncidentId, req.ServiceName, req.EventType, req.Message)
	if err != nil {
		metrics.GRPCRequestsTotal.WithLabelValues("SendNotification", "error").Inc()
		return &notificationv1.SendNotificationResponse{
			Success:   false,
			Timestamp: time.Now().Format(time.RFC3339),
		}, err
	}

	metrics.GRPCRequestsTotal.WithLabelValues("SendNotification", "success").Inc()
	return &notificationv1.SendNotificationResponse{
		Success:        true,
		NotificationId: notifID,
		Timestamp:      time.Now().Format(time.RFC3339),
	}, nil
}
