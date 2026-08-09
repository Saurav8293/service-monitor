package health

import (
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

func ServiceMonitor(cfg config.Config, services []model.RegisteredService) {

	client := &http.Client{
		Timeout: cfg.RequestTimeout,
	}

	previousStates := make(map[int64]ServiceState)

	ticker := time.NewTicker(cfg.HealthCheckInterval)
	defer ticker.Stop()

	for {
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
		fmt.Println("Payment service has problem:", err)
		return StateDown
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return StateHealthy
	} else {
		return StateDown
	}
}
