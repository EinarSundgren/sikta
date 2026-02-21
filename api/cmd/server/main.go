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
	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction"
	graphhandlers "github.com/einarsundgren/sikta/internal/handlers/graph"
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

	// Document handlers (shared between models)
	docHandler := handlers.NewDocumentHandler(pool, logger)
	mux.HandleFunc("POST /api/documents", docHandler.UploadDocument)
	mux.HandleFunc("GET /api/documents", docHandler.ListDocuments)
	mux.HandleFunc("GET /api/documents/{id}", docHandler.GetDocument)
	mux.HandleFunc("GET /api/documents/{id}/status", docHandler.GetDocumentStatus)
	mux.HandleFunc("DELETE /api/documents/{id}", docHandler.DeleteDocument)

	// Start background document processor (chunks uploaded documents)
	stopCh := make(chan struct{})
	go docHandler.ProcessDocuments(stopCh)

	db := database.New(pool)

	// Progress tracker for real-time extraction updates
	progressTracker := extraction.NewProgressTracker()

	// Extraction endpoints (shared - extraction process is the same)
	extractionHandler := handlers.NewExtractionHandler(db, cfg, logger, progressTracker)
	mux.HandleFunc("POST /api/documents/{id}/extract", extractionHandler.TriggerExtraction)
	mux.HandleFunc("GET /api/documents/{id}/extract/progress", extractionHandler.StreamProgress)

	// Graph model handlers
	graphTimelineHandler := graphhandlers.NewTimelineHandler(db, logger)
	graphEntitiesHandler := graphhandlers.NewEntitiesHandler(db, logger)
	graphRelationshipsHandler := graphhandlers.NewRelationshipsHandler(db, logger)
	graphStatusHandler := graphhandlers.NewStatusHandler(db, logger)
	graphReviewHandler := graphhandlers.NewReviewHandler(db, logger)

	mux.HandleFunc("GET /api/documents/{id}/timeline", graphTimelineHandler.GetTimeline)
	mux.HandleFunc("GET /api/documents/{id}/entities", graphEntitiesHandler.GetEntities)
	mux.HandleFunc("GET /api/documents/{id}/relationships", graphRelationshipsHandler.GetRelationships)
	mux.HandleFunc("GET /api/documents/{id}/events", graphTimelineHandler.GetTimeline) // Alias for timeline
	mux.HandleFunc("GET /api/documents/{id}/extract/status", graphStatusHandler.GetExtractionStatus)

	mux.HandleFunc("PATCH /api/claims/{id}/review", graphReviewHandler.UpdateNodeReview)
	mux.HandleFunc("PATCH /api/claims/{id}", graphReviewHandler.UpdateNodeData)
	mux.HandleFunc("PATCH /api/entities/{id}/review", graphReviewHandler.UpdateNodeReview)
	mux.HandleFunc("PATCH /api/relationships/{id}/review", graphReviewHandler.UpdateEdgeReview)
	mux.HandleFunc("GET /api/documents/{id}/review-progress", graphReviewHandler.GetReviewProgress)

	// Inconsistency handlers (currently only legacy)
	incHandler := handlers.NewInconsistencyHandler(db, cfg, logger)
	mux.HandleFunc("GET /api/documents/{id}/inconsistencies", incHandler.GetInconsistencies)
	mux.HandleFunc("POST /api/documents/{id}/detect-inconsistencies", incHandler.TriggerInconsistencyDetection)
	mux.HandleFunc("GET /api/inconsistencies/{id}/items", incHandler.GetInconsistencyItems)
	mux.HandleFunc("PUT /api/inconsistencies/{id}/resolve", incHandler.ResolveInconsistency)

	handler := middleware.RequestLogger(logger)(middleware.CORS(cfg.AllowedOrigins)(mux))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 10 * time.Minute, // Long timeout for SSE streams
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
	close(stopCh)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}
	logger.Info("server stopped")
}
