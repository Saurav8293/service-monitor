package repository

import (
	"context"
	"errors"
	"fmt"
	"service-monitor/pkg/model"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServiceRepository interface {
	Create(ctx context.Context, name string, endpoint string) (int64, error)
	GetAll(ctx context.Context) ([]model.RegisteredService, error)
	GetByID(ctx context.Context, id int64) (*model.RegisteredService, error)
	UpdateStatus(ctx context.Context, id int64, status model.ServiceState) error
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	CreateUser(ctx context.Context, username, passwordHash string) (int64, error)
}

type PostgresServiceRepository struct {
	db *pgxpool.Pool
}

func NewPostgresServiceRepository(db *pgxpool.Pool) *PostgresServiceRepository {
	return &PostgresServiceRepository{
		db: db,
	}
}

func (r *PostgresServiceRepository) Create(ctx context.Context, name string, endpoint string) (int64, error) {
	var id int64
	query := `
		INSERT INTO services (name, endpoint, status, created_at, updated_at)
		VALUES ($1, $2, 'unknown', NOW(), NOW())
		RETURNING id
	`
	err := r.db.QueryRow(ctx, query, name, endpoint).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert service: %w", err)
	}
	return id, nil
}

func (r *PostgresServiceRepository) GetAll(ctx context.Context) ([]model.RegisteredService, error) {
	query := `
		SELECT id, name, endpoint, status, created_at, updated_at
		FROM services
		ORDER BY id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query services: %w", err)
	}
	defer rows.Close()

	var services []model.RegisteredService
	for rows.Next() {
		var s model.RegisteredService
		if err := rows.Scan(&s.ID, &s.Name, &s.Endpoint, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		services = append(services, s)
	}

	return services, rows.Err()
}

func (r *PostgresServiceRepository) GetByID(ctx context.Context, id int64) (*model.RegisteredService, error) {
	query := `
		SELECT id, name, endpoint, status, created_at, updated_at
		FROM services
		WHERE id = $1
	`
	var s model.RegisteredService
	err := r.db.QueryRow(ctx, query, id).Scan(&s.ID, &s.Name, &s.Endpoint, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query service by id: %w", err)
	}
	return &s, nil
}

func (r *PostgresServiceRepository) UpdateStatus(ctx context.Context, id int64, status model.ServiceState) error {
	query := `
		UPDATE services
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, string(status), time.Now(), id)
	if err != nil {
		return fmt.Errorf("update service status: %w", err)
	}
	return nil
}

func (r *PostgresServiceRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = $1
	`
	var u model.User
	err := r.db.QueryRow(ctx, query, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by username: %w", err)
	}
	return &u, nil
}

func (r *PostgresServiceRepository) CreateUser(ctx context.Context, username, passwordHash string) (int64, error) {
	var id int64
	query := `
		INSERT INTO users (username, password_hash, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id
	`
	err := r.db.QueryRow(ctx, query, username, passwordHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}
