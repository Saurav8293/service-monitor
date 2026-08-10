package repository

import (
	"context"
	"service-monitor/internal/model"

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

func (r *PostgresServiceRepository) GetAll(ctx context.Context) ([]model.RegisteredService, error) {
	query := `
		SELECT id, name, endpoint FROM services
	`

	rows, err := r.db.Query(
		ctx,
		query,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var services []model.RegisteredService

	for rows.Next() {
		var service model.RegisteredService

		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.Endpoint,
		)

		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}
