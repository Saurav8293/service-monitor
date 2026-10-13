package client

import (
	"context"
	"fmt"
	"log/slog"
	notificationv1 "service-monitor/pkg/proto/notification/v1"
	"service-monitor/pkg/retry"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationClient struct {
	client notificationv1.NotificationServiceClient
	conn   *grpc.ClientConn
}

func NewNotificationClient(targetAddr string) (*NotificationClient, error) {
	conn, err := grpc.NewClient(
		targetAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to notification grpc service at %s: %w", targetAddr, err)
	}

	return &NotificationClient{
		client: notificationv1.NewNotificationServiceClient(conn),
		conn:   conn,
	}, nil
}

// SendNotification calls the Notification gRPC service with exponential backoff retry.
func (nc *NotificationClient) SendNotification(ctx context.Context, incidentID int64, serviceName, eventType, message string) error {
	req := &notificationv1.SendNotificationRequest{
		IncidentId:  incidentID,
		ServiceName: serviceName,
		EventType:   eventType,
		Message:     message,
	}

	op := func(opCtx context.Context) error {
		callCtx, cancel := context.WithTimeout(opCtx, 3*time.Second)
		defer cancel()

		res, err := nc.client.SendNotification(callCtx, req)
		if err != nil {
			slog.Warn("gRPC SendNotification failed, retrying...", "error", err, "incident_id", incidentID)
			return err
		}

		slog.Info("gRPC SendNotification confirmed by server",
			"incident_id", incidentID,
			"notification_id", res.NotificationId,
			"success", res.Success,
		)
		return nil
	}

	return retry.Do(ctx, retry.DefaultConfig, op)
}

func (nc *NotificationClient) Close() error {
	if nc.conn != nil {
		return nc.conn.Close()
	}
	return nil
}
