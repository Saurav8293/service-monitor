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
			previousState := previousStates[service.ID]

			if previousState != "" && previousState != currentState {
				if previousState == StateHealthy && currentState == StateDown {
					fmt.Printf("%s is DOWN\n", service.Name)
				}

				if previousState == StateDown && currentState == StateHealthy {
					fmt.Printf("%s RECOVERED\n", service.Name)
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
