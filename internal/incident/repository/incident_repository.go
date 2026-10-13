package repository

import (
	"context"
	"fmt"
	"service-monitor/pkg/model"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IncidentRepository interface {
	CreateIdempotent(ctx context.Context, inc model.Incident) (int64, bool, error)
	ResolveLatest(ctx context.Context, serviceName string) (*model.Incident, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	GetByID(ctx context.Context, id int64) (*model.Incident, error)
}

type PostgresIncidentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresIncidentRepository(db *pgxpool.Pool) *PostgresIncidentRepository {
	return &PostgresIncidentRepository{
		db: db,
	}
}

// CreateIdempotent inserts an incident. Returns (id, created, error).
// If idempotency_key already exists, returns created = false without error.
func (r *PostgresIncidentRepository) CreateIdempotent(ctx context.Context, inc model.Incident) (int64, bool, error) {
	query := `
		INSERT INTO incidents (service_name, event_type, status, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, inc.ServiceName, string(inc.EventType), inc.Status, inc.IdempotencyKey).Scan(&id)
	if err != nil {
		// If no row was inserted due to ON CONFLICT DO NOTHING
		if err.Error() == "no rows in result set" {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("insert incident: %w", err)
	}

	return id, true, nil
}

// ResolveLatest marks any open or notified incident for this service as resolved.
func (r *PostgresIncidentRepository) ResolveLatest(ctx context.Context, serviceName string) (*model.Incident, error) {
	query := `
		UPDATE incidents
		SET status = 'resolved', resolved_at = NOW()
		WHERE id = (
			SELECT id FROM incidents
			WHERE service_name = $1 AND status != 'resolved'
			ORDER BY created_at DESC
			LIMIT 1
		)
		RETURNING id, service_name, event_type, status, idempotency_key, created_at, resolved_at
	`
	var inc model.Incident
	var eventTypeStr string
	err := r.db.QueryRow(ctx, query, serviceName).Scan(
		&inc.ID,
		&inc.ServiceName,
		&eventTypeStr,
		&inc.Status,
		&inc.IdempotencyKey,
		&inc.CreatedAt,
		&inc.ResolvedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // No active open incident to resolve
		}
		return nil, fmt.Errorf("resolve incident: %w", err)
	}
	inc.EventType = model.EventType(eventTypeStr)
	return &inc, nil
}

func (r *PostgresIncidentRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE incidents SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *PostgresIncidentRepository) GetByID(ctx context.Context, id int64) (*model.Incident, error) {
	query := `
		SELECT id, service_name, event_type, status, idempotency_key, created_at, resolved_at
		FROM incidents
		WHERE id = $1
	`
	var inc model.Incident
	var eventTypeStr string
	var resolvedAt *time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(
		&inc.ID,
		&inc.ServiceName,
		&eventTypeStr,
		&inc.Status,
		&inc.IdempotencyKey,
		&inc.CreatedAt,
		&resolvedAt,
	)
	if err != nil {
		return nil, err
	}
	inc.EventType = model.EventType(eventTypeStr)
	inc.ResolvedAt = resolvedAt
	return &inc, nil
}
