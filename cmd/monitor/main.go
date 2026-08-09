package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"service-monitor/internal/config"
	"service-monitor/internal/database"
	"service-monitor/internal/handler"
	"service-monitor/internal/health"
	"service-monitor/internal/repository"
	"service-monitor/internal/service"
	"time"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("Connected to PostgreSQL!")

	serviceRepo := repository.NewPostgresServiceRepository(db)

	serviceManager := service.NewService(serviceRepo)

	serviceHandler := handler.NewServiceHandler(serviceManager)

	startServer(cfg, serviceHandler)
}

func startServer(cfg config.Config, serviceHandler *handler.ServiceHandler) {

	mux := http.NewServeMux()

	mux.HandleFunc("/services", serviceHandler.RegisterService)

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
