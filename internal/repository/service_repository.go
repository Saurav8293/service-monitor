package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
		INSERT INTO services (name, endpoint)
		VALUES ($1,$2)
		RETURNING id
	`

	err := r.db.QueryRow(
		ctx,
		query,
		name,
		endpoint,
	).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}
