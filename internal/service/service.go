package service

import (
	"context"
	"fmt"
	"strings"
)

type ServiceRepository interface {
	Create(ctx context.Context, name string, endpoint string) (int64, error)
}

type Service struct {
	repo ServiceRepository
}

func NewService(repo ServiceRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) RegisterService(ctx context.Context, name string, endpoint string) (int64, error) {

	name = strings.TrimSpace(name)
	endpoint = strings.TrimSpace(endpoint)

	if name == "" {
		return 0, fmt.Errorf("service name is required")
	}

	if endpoint == "" {
		return 0, fmt.Errorf("service endpoint is required")
	}
	return s.repo.Create(ctx, name, endpoint)

}
