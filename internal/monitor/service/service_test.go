package service_test

import (
	"context"
	"errors"
	"service-monitor/internal/monitor/auth"
	"service-monitor/internal/monitor/service"
	"service-monitor/pkg/model"
	"testing"
)

type mockServiceRepo struct {
	services []model.RegisteredService
	users    map[string]*model.User
}

func (m *mockServiceRepo) Create(ctx context.Context, name string, endpoint string) (int64, error) {
	id := int64(len(m.services) + 1)
	m.services = append(m.services, model.RegisteredService{
		ID:       id,
		Name:     name,
		Endpoint: endpoint,
		Status:   model.StateUnknown,
	})
	return id, nil
}

func (m *mockServiceRepo) GetAll(ctx context.Context) ([]model.RegisteredService, error) {
	return m.services, nil
}

func (m *mockServiceRepo) GetByID(ctx context.Context, id int64) (*model.RegisteredService, error) {
	for _, s := range m.services {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, nil
}

func (m *mockServiceRepo) UpdateStatus(ctx context.Context, id int64, status model.ServiceState) error {
	for i, s := range m.services {
		if s.ID == id {
			m.services[i].Status = status
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockServiceRepo) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockServiceRepo) CreateUser(ctx context.Context, username, passwordHash string) (int64, error) {
	id := int64(len(m.users) + 1)
	m.users[username] = &model.User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
	}
	return id, nil
}

func TestRegisterServiceValidation(t *testing.T) {
	repo := &mockServiceRepo{users: make(map[string]*model.User)}
	tm := auth.NewTokenManager("test-secret")
	svc := service.NewService(repo, tm)

	ctx := context.Background()

	// Empty name
	_, err := svc.RegisterService(ctx, "", "http://localhost:8081/health")
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}

	// Empty endpoint
	_, err = svc.RegisterService(ctx, "payment", "")
	if err == nil {
		t.Fatal("expected error for empty endpoint, got nil")
	}

	// Invalid URL
	_, err = svc.RegisterService(ctx, "payment", "invalid-no-scheme")
	if err == nil {
		t.Fatal("expected error for invalid URL without scheme, got nil")
	}

	// Valid registration
	id, err := svc.RegisterService(ctx, "payment-service", "http://localhost:8081/health")
	if err != nil {
		t.Fatalf("unexpected error on valid registration: %v", err)
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}

	services, err := svc.GetAllServices(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting services: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
}

func TestLoginAuthentication(t *testing.T) {
	repo := &mockServiceRepo{users: make(map[string]*model.User)}
	tm := auth.NewTokenManager("test-secret")
	svc := service.NewService(repo, tm)

	ctx := context.Background()

	hash, _ := auth.HashPassword("secret123")
	_, _ = repo.CreateUser(ctx, "admin", hash)

	// Valid login
	token, err := svc.Login(ctx, "admin", "secret123")
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Invalid password
	_, err = svc.Login(ctx, "admin", "wrongpass")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}

	// Non-existent user
	_, err = svc.Login(ctx, "nonexistent", "secret123")
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}
