package health

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"service-monitor/internal/monitor/config"
	"service-monitor/internal/monitor/repository"
	"service-monitor/pkg/kafka"
	"service-monitor/pkg/metrics"
	"service-monitor/pkg/model"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CheckResult struct {
	Service model.RegisteredService
	State   model.ServiceState
	Error   error
}

type HealthChecker struct {
	cfg            config.Config
	repo           repository.ServiceRepository
	kafkaProducer  *kafka.Producer
	httpClient     *http.Client
	previousStates map[int64]model.ServiceState
	mu             sync.RWMutex
}

func NewHealthChecker(cfg config.Config, repo repository.ServiceRepository, producer *kafka.Producer) *HealthChecker {
	return &HealthChecker{
		cfg:            cfg,
		repo:           repo,
		kafkaProducer:  producer,
		httpClient:     &http.Client{Timeout: cfg.RequestTimeout},
		previousStates: make(map[int64]model.ServiceState),
	}
}

// Start runs the periodic health check loop until ctx is canceled.
func (hc *HealthChecker) Start(ctx context.Context) {
	slog.Info("starting health checker worker pool",
		"interval", hc.cfg.HealthCheckInterval,
		"workers", hc.cfg.HealthWorkerCount,
		"timeout", hc.cfg.RequestTimeout,
	)

	ticker := time.NewTicker(hc.cfg.HealthCheckInterval)
	defer ticker.Stop()

	// Run initial check immediately
	hc.runCheckCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("health checker received shutdown signal, exiting loop")
			return
		case <-ticker.C:
			hc.runCheckCycle(ctx)
		}
	}
}

// runCheckCycle executes a bounded worker pool fan-out / fan-in check across all registered services.
func (hc *HealthChecker) runCheckCycle(ctx context.Context) {
	services, err := hc.repo.GetAll(ctx)
	if err != nil {
		slog.Error("failed to query monitored services from database", "error", err)
		return
	}

	if len(services) == 0 {
		return
	}

	metrics.ServicesMonitored.Set(float64(len(services)))

	numWorkers := hc.cfg.HealthWorkerCount
	if numWorkers > len(services) {
		numWorkers = len(services)
	}

	jobs := make(chan model.RegisteredService, len(services))
	results := make(chan CheckResult, len(services))

	var wg sync.WaitGroup

	// Spawn bounded worker pool
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for svc := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					res := hc.probeService(ctx, svc)
					results <- res
				}
			}
		}(w)
	}

	// Fan-out: send all services into the jobs channel
	for _, svc := range services {
		jobs <- svc
	}
	close(jobs)

	// Wait for workers to finish in a separate goroutine and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Fan-in: process check results
	for res := range results {
		hc.handleResult(ctx, res)
	}
}

func (hc *HealthChecker) probeService(ctx context.Context, svc model.RegisteredService) CheckResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svc.Endpoint, nil)
	if err != nil {
		metrics.HealthChecksTotal.WithLabelValues(svc.Name, "error").Inc()
		return CheckResult{Service: svc, State: model.StateDown, Error: err}
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		metrics.HealthChecksTotal.WithLabelValues(svc.Name, "down").Inc()
		return CheckResult{Service: svc, State: model.StateDown, Error: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		metrics.HealthChecksTotal.WithLabelValues(svc.Name, "healthy").Inc()
		return CheckResult{Service: svc, State: model.StateHealthy}
	}

	metrics.HealthChecksTotal.WithLabelValues(svc.Name, "down").Inc()
	return CheckResult{
		Service: svc,
		State:   model.StateDown,
		Error:   fmt.Errorf("unexpected status code: %d", resp.StatusCode),
	}
}

func (hc *HealthChecker) handleResult(ctx context.Context, res CheckResult) {
	svc := res.Service
	currentState := res.State

	hc.mu.Lock()
	prevState, exists := hc.previousStates[svc.ID]
	hc.previousStates[svc.ID] = currentState
	hc.mu.Unlock()

	var stateTransitionEvent *model.ServiceEvent

	if exists {
		if prevState == model.StateHealthy && currentState == model.StateDown {
			slog.Warn("service transitioned to DOWN", "service", svc.Name, "endpoint", svc.Endpoint, "err", res.Error)
			stateTransitionEvent = &model.ServiceEvent{
				EventID:     uuid.NewString(),
				EventType:   model.EventServiceDown,
				ServiceID:   svc.ID,
				ServiceName: svc.Name,
				Endpoint:    svc.Endpoint,
				Timestamp:   time.Now(),
				Message:     fmt.Sprintf("Service %s failed health probe: %v", svc.Name, res.Error),
			}
		} else if prevState == model.StateDown && currentState == model.StateHealthy {
			slog.Info("service RECOVERED to healthy", "service", svc.Name, "endpoint", svc.Endpoint)
			stateTransitionEvent = &model.ServiceEvent{
				EventID:     uuid.NewString(),
				EventType:   model.EventServiceRecovered,
				ServiceID:   svc.ID,
				ServiceName: svc.Name,
				Endpoint:    svc.Endpoint,
				Timestamp:   time.Now(),
				Message:     fmt.Sprintf("Service %s recovered and is responding with 200 OK", svc.Name),
			}
		}
	} else {
		// Initial probe
		if currentState == model.StateDown {
			slog.Warn("initial probe: service is DOWN", "service", svc.Name, "endpoint", svc.Endpoint, "err", res.Error)
			stateTransitionEvent = &model.ServiceEvent{
				EventID:     uuid.NewString(),
				EventType:   model.EventServiceDown,
				ServiceID:   svc.ID,
				ServiceName: svc.Name,
				Endpoint:    svc.Endpoint,
				Timestamp:   time.Now(),
				Message:     fmt.Sprintf("Initial probe failed for service %s: %v", svc.Name, res.Error),
			}
		}
	}

	// Update database status
	if err := hc.repo.UpdateStatus(ctx, svc.ID, currentState); err != nil {
		slog.Error("failed to update service status in db", "service_id", svc.ID, "error", err)
	}

	// Publish to Kafka if a transition occurred
	if stateTransitionEvent != nil && hc.kafkaProducer != nil {
		if err := hc.kafkaProducer.PublishEvent(ctx, *stateTransitionEvent); err != nil {
			slog.Error("failed to publish state change event to Kafka", "service", svc.Name, "error", err)
		}
	}
}
