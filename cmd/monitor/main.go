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
	"service-monitor/internal/model"
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

	services, err := serviceManager.GetAllServices(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Registered services:", len(services))

	serviceHandler := handler.NewServiceHandler(serviceManager)

	startServer(cfg, serviceHandler, services)
}

func startServer(cfg config.Config, serviceHandler *handler.ServiceHandler, services []model.RegisteredService) {

	mux := http.NewServeMux()

	mux.HandleFunc("/services", serviceHandler.RegisterService)

	s := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("HTTP server is running at", cfg.ServerAddr)

	go health.ServiceMonitor(cfg, services)

	log.Fatal(s.ListenAndServe())
}
