package main

import (
	"fmt"
	"log"
	"net/http"
	"service-monitor/internal/config"
	"service-monitor/internal/health"
	"time"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	startServer(cfg)
}

func startServer(cfg config.Config) {

	mux := http.NewServeMux()

	s := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("HTTP server is running at", cfg.ServerAddr)

	go health.ServiceMonitor(cfg)

	log.Fatal(s.ListenAndServe())
}
