package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/einarsundgren/sikta/internal/config"
	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Get document path from args
	if len(os.Args) < 2 {
		logger.Error("usage: extract <document-path>")
		os.Exit(1)
	}
	docPath := os.Args[1]

	ctx := context.Background()

	// Connect to database
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := database.New(pool)

	// Find document by path
	doc, err := findDocumentByPath(ctx, queries, docPath)
	if err != nil {
		logger.Error("failed to find document", "error", err)
		os.Exit(1)
	}

	logger.Info("starting extraction", "document", doc.Title, "chapters", countChunks(ctx, queries, doc.ID))

	// Create extraction service
	claudeClient := claude.NewClient(cfg, logger)
	extractService := extraction.NewService(queries, claudeClient, logger)

	// Run extraction with progress tracking
	startTime := time.Now()

	err = extractService.ExtractDocument(ctx, doc.ID.String(), func(progress extraction.ExtractionProgress) {
		if progress.TotalChunks > 0 {
			percentage := (progress.ProcessedChunks * 100) / progress.TotalChunks
			logger.Info("progress",
				"chunk", fmt.Sprintf("%d/%d", progress.ProcessedChunks, progress.TotalChunks),
				"events", progress.EventsExtracted,
				"entities", progress.EntitiesExtracted,
				"relationships", progress.RelationshipsExtracted,
				"status", progress.Status)
		} else if progress.Status == "complete" {
			elapsed := time.Since(startTime)
			logger.Info("extraction complete", "duration", elapsed)
		}
	})

	if err != nil {
		logger.Error("extraction failed", "error", err)
		os.Exit(1)
	}

	// Show summary
	showSummary(ctx, queries, doc.ID, logger)
}

func findDocumentByPath(ctx context.Context, db *database.Queries, path string) (database.Document, error) {
	docs, err := db.ListDocuments(ctx)
	if err != nil {
		return database.Document{}, err
	}

	for _, doc := range docs {
		if doc.FilePath == path {
			return doc, nil
		}
	}

	return database.Document{}, fmt.Errorf("document not found: %s", path)
}

func countChunks(ctx context.Context, db *database.Queries, docID database.UUID) int {
	count, _ := db.CountChunksByDocument(ctx, docID)
	return int(count)
}

func showSummary(ctx context.Context, db *database.Queries, docID database.UUID, logger *slog.Logger) {
	eventCount, _ := db.CountEventsByDocument(ctx, docID)
	entityCount, _ := db.CountEntitiesByDocument(ctx, docID)
	relationshipCount, _ := db.CountRelationshipsByDocument(ctx, docID)

	logger.Info("extraction summary",
		"events", eventCount,
		"entities", entityCount,
		"relationships", relationshipCount)

	// Validate against acceptance criteria
	if eventCount < 50 {
		logger.Warn("event count below expected range (50-80)", "count", eventCount)
	}
	if entityCount < 20 {
		logger.Warn("entity count below expected range (20-30)", "count", entityCount)
	}
	if relationshipCount < 15 {
		logger.Warn("relationship count below expected range (15-25)", "count", relationshipCount)
	}

	logger.Info("extraction successful")
}
