package model

import "time"

// ServiceState represents the operational status of a monitored service.
type ServiceState string

const (
	StateHealthy ServiceState = "healthy"
	StateDown    ServiceState = "down"
	StateUnknown ServiceState = "unknown"
)

// EventType represents the category of health transition event.
type EventType string

const (
	EventServiceDown      EventType = "SERVICE_DOWN"
	EventServiceRecovered EventType = "SERVICE_RECOVERED"
)

// ServiceEvent defines the payload published to Kafka upon health state transitions.
type ServiceEvent struct {
	EventID     string    `json:"event_id"`
	EventType   EventType `json:"event_type"`
	ServiceID   int64     `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Endpoint    string    `json:"endpoint"`
	Timestamp   time.Time `json:"timestamp"`
	Message     string    `json:"message,omitempty"`
}

// RegisteredService represents a service registered with the monitoring system.
type RegisteredService struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Endpoint  string       `json:"endpoint"`
	Status    ServiceState `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// Incident represents a trackable incident created by the Incident Service.
type Incident struct {
	ID             int64      `json:"id"`
	ServiceName    string     `json:"service_name"`
	EventType      EventType  `json:"event_type"`
	Status         string     `json:"status"` // open, notified, resolved
	IdempotencyKey string     `json:"idempotency_key"`
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

// Notification represents an alert dispatched by the Notification Service.
type Notification struct {
	ID         int64     `json:"id"`
	IncidentID int64     `json:"incident_id"`
	Channel    string    `json:"channel"`
	Message    string    `json:"message"`
	SentAt     time.Time `json:"sent_at"`
}

// User represents an operator with access to the Monitor API.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
