package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"service-monitor/internal/incident/client"
	"service-monitor/internal/incident/config"
	"service-monitor/internal/incident/consumer"
	"service-monitor/internal/incident/repository"
	"service-monitor/internal/incident/service"
	"service-monitor/pkg/database"
	"service-monitor/pkg/kafka"
	"service-monitor/pkg/logging"
	"service-monitor/pkg/metrics"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	logger := logging.InitLogger(cfg.LogLevel, "text")
	logger.Info("starting Incident Service", "metrics_addr", cfg.MetricsAddr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Database connection & migrations
	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := database.RunMigrations(ctx, dbPool); err != nil {
		logger.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}

	// 2. gRPC Client to Notification Service
	notifClient, err := client.NewNotificationClient(cfg.NotificationURL)
	if err != nil {
		logger.Warn("could not connect to Notification gRPC service immediately; will retry during calls", "error", err)
	} else {
		defer notifClient.Close()
	}

	// 3. Repository and Service
	repo := repository.NewPostgresIncidentRepository(dbPool)
	incidentSvc := service.NewIncidentService(repo, notifClient)

	// 4. Kafka Consumer
	kafkaConsumer := kafka.NewConsumer(cfg.KafkaBroker, cfg.KafkaTopic, cfg.KafkaGroupID)
	defer kafkaConsumer.Close()

	eventConsumer := consumer.NewEventConsumer(kafkaConsumer, incidentSvc)

	// Start consumer loop in background
	go func() {
		if err := eventConsumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("Kafka consumer loop stopped with error", "error", err)
		}
	}()

	// 5. Lightweight HTTP Server for Metrics & Health
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"incident"}`))
	})

	server := &http.Server{
		Addr:         cfg.MetricsAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("Metrics server listening", "addr", cfg.MetricsAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("metrics server error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("received termination signal, shutting down Incident Service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)
	logger.Info("Incident Service cleanly stopped")
}
