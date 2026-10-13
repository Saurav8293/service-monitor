package service_test

import (
	"context"
	"service-monitor/internal/incident/service"
	"service-monitor/pkg/model"
	"testing"
	"time"
)

type mockIncidentRepo struct {
	incidents map[string]model.Incident // keyed by idempotency_key
	latestID  int64
}

func newMockIncidentRepo() *mockIncidentRepo {
	return &mockIncidentRepo{
		incidents: make(map[string]model.Incident),
	}
}

func (m *mockIncidentRepo) CreateIdempotent(ctx context.Context, inc model.Incident) (int64, bool, error) {
	if _, exists := m.incidents[inc.IdempotencyKey]; exists {
		return 0, false, nil // duplicate ignored
	}
	m.latestID++
	inc.ID = m.latestID
	inc.CreatedAt = time.Now()
	m.incidents[inc.IdempotencyKey] = inc
	return inc.ID, true, nil
}

func (m *mockIncidentRepo) ResolveLatest(ctx context.Context, serviceName string) (*model.Incident, error) {
	for k, inc := range m.incidents {
		if inc.ServiceName == serviceName && inc.Status != "resolved" {
			now := time.Now()
			inc.Status = "resolved"
			inc.ResolvedAt = &now
			m.incidents[k] = inc
			return &inc, nil
		}
	}
	return nil, nil
}

func (m *mockIncidentRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	for k, inc := range m.incidents {
		if inc.ID == id {
			inc.Status = status
			m.incidents[k] = inc
			return nil
		}
	}
	return nil
}

func (m *mockIncidentRepo) GetByID(ctx context.Context, id int64) (*model.Incident, error) {
	for _, inc := range m.incidents {
		if inc.ID == id {
			return &inc, nil
		}
	}
	return nil, nil
}

func TestIncidentIdempotency(t *testing.T) {
	repo := newMockIncidentRepo()
	svc := service.NewIncidentService(repo, nil) // no gRPC client in unit test
	ctx := context.Background()

	event := model.ServiceEvent{
		EventID:     "unique-uuid-101",
		EventType:   model.EventServiceDown,
		ServiceID:   1,
		ServiceName: "payment-service",
		Endpoint:    "http://localhost:8081/health",
		Timestamp:   time.Now(),
	}

	// First time receiving event
	err := svc.HandleServiceEvent(ctx, event)
	if err != nil {
		t.Fatalf("unexpected error on first event: %v", err)
	}

	if len(repo.incidents) != 1 {
		t.Fatalf("expected 1 incident stored, got %d", len(repo.incidents))
	}

	// Replaying duplicate event (same EventID / idempotency key)
	err = svc.HandleServiceEvent(ctx, event)
	if err != nil {
		t.Fatalf("unexpected error on duplicate event: %v", err)
	}

	// Must NOT create duplicate row
	if len(repo.incidents) != 1 {
		t.Fatalf("expected still exactly 1 incident stored after duplicate, got %d", len(repo.incidents))
	}

	// Recovery event
	recoveryEvent := model.ServiceEvent{
		EventID:     "unique-uuid-102",
		EventType:   model.EventServiceRecovered,
		ServiceID:   1,
		ServiceName: "payment-service",
		Endpoint:    "http://localhost:8081/health",
		Timestamp:   time.Now(),
	}

	err = svc.HandleServiceEvent(ctx, recoveryEvent)
	if err != nil {
		t.Fatalf("unexpected error on recovery event: %v", err)
	}

	inc := repo.incidents["unique-uuid-101"]
	if inc.Status != "resolved" {
		t.Fatalf("expected incident status 'resolved', got %s", inc.Status)
	}
	if inc.ResolvedAt == nil {
		t.Fatal("expected ResolvedAt timestamp to be set")
	}
}
