package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"service-monitor/internal/monitor/auth"
	"service-monitor/internal/monitor/repository"
	"service-monitor/pkg/metrics"
	"service-monitor/pkg/model"
	"strings"
)

type Service struct {
	repo         repository.ServiceRepository
	tokenManager *auth.TokenManager
}

func NewService(repo repository.ServiceRepository, tokenManager *auth.TokenManager) *Service {
	return &Service{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

func (s *Service) RegisterService(ctx context.Context, name, endpoint string) (int64, error) {
	name = strings.TrimSpace(name)
	endpoint = strings.TrimSpace(endpoint)

	if name == "" {
		return 0, errors.New("service name is required")
	}

	if endpoint == "" {
		return 0, errors.New("service endpoint is required")
	}

	u, err := url.ParseRequestURI(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return 0, errors.New("invalid endpoint URL: must include valid scheme (e.g. http:// or https://) and host")
	}

	id, err := s.repo.Create(ctx, name, endpoint)
	if err != nil {
		return 0, err
	}

	// Update Prometheus monitored gauge
	services, _ := s.repo.GetAll(ctx)
	metrics.ServicesMonitored.Set(float64(len(services)))

	return id, nil
}

func (s *Service) GetAllServices(ctx context.Context) ([]model.RegisteredService, error) {
	services, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	metrics.ServicesMonitored.Set(float64(len(services)))
	return services, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", errors.New("username and password are required")
	}

	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	token, err := s.tokenManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (s *Service) ValidateToken(tokenString string) (*auth.Claims, error) {
	return s.tokenManager.ValidateToken(tokenString)
}
