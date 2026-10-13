package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"service-monitor/internal/monitor/auth"
	"service-monitor/internal/monitor/config"
	"service-monitor/internal/monitor/handler"
	"service-monitor/internal/monitor/health"
	"service-monitor/internal/monitor/repository"
	"service-monitor/internal/monitor/service"
	"service-monitor/pkg/database"
	"service-monitor/pkg/kafka"
	"service-monitor/pkg/logging"
	"service-monitor/pkg/metrics"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := logging.InitLogger(cfg.LogLevel, "text")
	logger.Info("starting Monitor Service", "addr", cfg.ServerAddr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	kafkaProducer := kafka.NewProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	defer kafkaProducer.Close()

	repo := repository.NewPostgresServiceRepository(dbPool)
	tokenMgr := auth.NewTokenManager(cfg.JWTSecret)
	svc := service.NewService(repo, tokenMgr)
	h := handler.NewHandler(svc)

	healthChecker := health.NewHealthChecker(cfg, repo, kafkaProducer)
	go healthChecker.Start(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("GET /api/v1/auth/validate", h.ValidateToken)
	mux.HandleFunc("POST /api/v1/services", h.RegisterService)
	mux.HandleFunc("GET /api/v1/services", h.GetServices)
	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("GET /metrics", metrics.Handler())

	// Backward compatibility routes for legacy endpoints
	mux.HandleFunc("POST /services", h.RegisterService)
	mux.HandleFunc("GET /services/list", h.GetServices)

	instrumentedMux := metrics.Middleware("monitor-service", mux)

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      instrumentedMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start HTTP server
	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", "addr", cfg.ServerAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// Wait for shutdown signal or fatal server error
	select {
	case <-ctx.Done():
		logger.Info("received termination signal, initiating graceful shutdown...")
	case err := <-serverErrChan:
		logger.Error("server encountered fatal error", "error", err)
	}

	// Graceful shutdown with 5-second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to gracefully shutdown HTTP server", "error", err)
	} else {
		logger.Info("HTTP server cleanly stopped")
	}

	logger.Info("Monitor Service shutdown complete")
}
