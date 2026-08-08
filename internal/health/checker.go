package health

import (
	"fmt"
	"net/http"
	"service-monitor/internal/config"
	"time"
)

func ServiceMonitor(cfg config.Config) {

	client := &http.Client{
		Timeout: cfg.RequestTimeout,
	}

	ticker := time.NewTicker(cfg.HealthCheckInterval)
	defer ticker.Stop()

	for {
		checkHealth(client, cfg.HealthURL)
		<-ticker.C
	}
}

func checkHealth(client *http.Client, healthURL string) {

	resp, err := client.Get(healthURL)

	if err != nil {
		fmt.Println("Payment service has problem:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Payment service is healthy")
	} else {
		fmt.Println("Payment service has problem:", resp.Status)
	}
}
