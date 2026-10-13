package repository

import (
	"context"
	"fmt"
	"service-monitor/pkg/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository interface {
	Create(ctx context.Context, notif model.Notification) (int64, error)
}

type PostgresNotificationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresNotificationRepository(db *pgxpool.Pool) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{
		db: db,
	}
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, notif model.Notification) (int64, error) {
	query := `
		INSERT INTO notifications (incident_id, channel, message, sent_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, notif.IncidentID, notif.Channel, notif.Message).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert notification: %w", err)
	}
	return id, nil
}
