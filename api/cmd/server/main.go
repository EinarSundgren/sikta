package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/einarsundgren/sikta/internal/config"
	"github.com/einarsundgren/sikta/internal/handlers"
	"github.com/einarsundgren/sikta/internal/middleware"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health)

	// Document handlers
	docHandler := handlers.NewDocumentHandler(pool, logger)
	mux.HandleFunc("POST /api/documents", docHandler.UploadDocument)
	mux.HandleFunc("GET /api/documents/", docHandler.GetDocument)
	mux.HandleFunc("GET /api/documents", docHandler.ListDocuments)
	mux.HandleFunc("DELETE /api/documents/", docHandler.DeleteDocument)
	mux.HandleFunc("GET /api/documents/", docHandler.GetDocumentStatus)

	// Start background document processor
	stopCh := make(chan struct{})
	go docHandler.ProcessDocuments(stopCh)

	handler := middleware.CORS(cfg.AllowedOrigins)(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}
	logger.Info("server stopped")
}
