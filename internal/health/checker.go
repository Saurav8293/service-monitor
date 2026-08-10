package health

import (
	"context"
	"fmt"
	"net/http"
	"service-monitor/internal/config"
	"service-monitor/internal/model"
	"time"
)

type ServiceState string

const (
	StateHealthy ServiceState = "healthy"
	StateDown    ServiceState = "down"
)

type ServiceRepository interface {
	GetAll(ctx context.Context) ([]model.RegisteredService, error)
}

func ServiceMonitor(cfg config.Config, repo ServiceRepository) {

	client := &http.Client{
		Timeout: cfg.RequestTimeout,
	}
	ctx := context.Background()

	previousStates := make(map[int64]ServiceState)

	ticker := time.NewTicker(cfg.HealthCheckInterval)
	defer ticker.Stop()

	for {
		services, err := repo.GetAll(ctx)
		if err != nil {
			fmt.Println("failed to load services:", err)
			<-ticker.C
			continue
		}

		for _, service := range services {
			currentState := checkHealth(client, service)
			previousState, exists := previousStates[service.ID]

			if exists {
				if previousState == StateHealthy && currentState == StateDown {
					fmt.Printf("[ALERT] %s is DOWN\n", service.Name)
				} else if previousState == StateDown && currentState == StateHealthy {
					fmt.Printf("[INFO] %s RECOVERED\n", service.Name)
				}
			} else {
				if currentState == StateDown {
					fmt.Printf("[ALERT] %s is DOWN (initial check)\n", service.Name)
				}
			}
			previousStates[service.ID] = currentState
		}
		<-ticker.C
	}
}

func checkHealth(client *http.Client, service model.RegisteredService) ServiceState {

	resp, err := client.Get(service.Endpoint)

	if err != nil {
		fmt.Printf("%s health check failed: %v\n", service.Name, err)
		return StateDown
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return StateHealthy
	} else {
		return StateDown
	}
}
