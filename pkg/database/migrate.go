package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS services (
		id         BIGSERIAL PRIMARY KEY,
		name       TEXT NOT NULL UNIQUE,
		endpoint   TEXT NOT NULL,
		status     TEXT NOT NULL DEFAULT 'unknown',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_services_status ON services(status);`,

	`CREATE TABLE IF NOT EXISTS users (
		id            BIGSERIAL PRIMARY KEY,
		username      TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	INSERT INTO users (username, password_hash)
	VALUES ('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi')
	ON CONFLICT (username) DO NOTHING;`,

	`CREATE TABLE IF NOT EXISTS incidents (
		id              BIGSERIAL PRIMARY KEY,
		service_name    TEXT NOT NULL,
		event_type      TEXT NOT NULL,
		status          TEXT NOT NULL DEFAULT 'open',
		idempotency_key TEXT NOT NULL UNIQUE,
		created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		resolved_at     TIMESTAMPTZ
	);
	CREATE INDEX IF NOT EXISTS idx_incidents_service ON incidents(service_name);
	CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);`,

	`CREATE TABLE IF NOT EXISTS notifications (
		id          BIGSERIAL PRIMARY KEY,
		incident_id BIGINT NOT NULL,
		channel     TEXT NOT NULL DEFAULT 'log',
		message     TEXT NOT NULL,
		sent_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_notifications_incident ON notifications(incident_id);`,
}

// RunMigrations applies schema migrations idempotently.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	slog.Info("running database migrations...")
	for idx, query := range migrations {
		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("migration %d failed: %w", idx+1, err)
		}
	}
	slog.Info("database migrations applied successfully", "count", len(migrations))
	return nil
}
