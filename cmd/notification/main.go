package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"service-monitor/internal/notification/config"
	"service-monitor/internal/notification/handler"
	"service-monitor/internal/notification/repository"
	"service-monitor/internal/notification/service"
	"service-monitor/pkg/database"
	"service-monitor/pkg/logging"
	"service-monitor/pkg/metrics"
	notificationv1 "service-monitor/pkg/proto/notification/v1"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	logger := logging.InitLogger(cfg.LogLevel, "text")
	logger.Info("starting Notification Service", "grpc_port", cfg.GRPCPort, "metrics_addr", cfg.MetricsAddr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Database Connection & Migrations
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

	// 2. Repository, Service, and gRPC Handler
	repo := repository.NewPostgresNotificationRepository(dbPool)
	notifSvc := service.NewNotificationService(repo)
	grpcHandler := handler.NewGRPCHandler(notifSvc)

	// 3. gRPC Server
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen on gRPC port", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	notificationv1.RegisterNotificationServiceServer(grpcServer, grpcHandler)

	go func() {
		logger.Info("gRPC server listening", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server stopped", "error", err)
		}
	}()

	// 4. HTTP Server for Metrics & Health
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"notification"}`))
	})

	httpServer := &http.Server{
		Addr:         cfg.MetricsAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("Metrics server listening", "addr", cfg.MetricsAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("metrics server error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("received termination signal, shutting down Notification Service...")

	// Graceful shutdown of gRPC and HTTP
	grpcServer.GracefulStop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)

	logger.Info("Notification Service cleanly stopped")
}
