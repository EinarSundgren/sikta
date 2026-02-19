package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/einarsundgren/sikta/internal/config"
	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		logger.Error("usage: extract <document-path>")
		os.Exit(1)
	}
	docPath := os.Args[1]

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := database.New(pool)

	src, err := findSourceByPath(ctx, queries, docPath)
	if err != nil {
		logger.Error("failed to find source", "error", err)
		os.Exit(1)
	}

	logger.Info("starting extraction", "source", src.Title, "chunks", countChunks(ctx, queries, src.ID))

	claudeClient := claude.NewClient(cfg, logger)
	extractService := extraction.NewService(queries, claudeClient, logger, cfg.AnthropicModelExtraction)

	startTime := time.Now()

	err = extractService.ExtractDocument(ctx, database.UUIDStr(src.ID), func(progress extraction.ExtractionProgress) {
		if progress.TotalChunks > 0 {
			_ = (progress.ProcessedChunks * 100) / progress.TotalChunks
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

	showSummary(ctx, queries, src.ID, logger)
}

func findSourceByPath(ctx context.Context, db *database.Queries, path string) (*database.Source, error) {
	sources, err := db.ListSources(ctx)
	if err != nil {
		return nil, err
	}

	for _, src := range sources {
		if src.FilePath == path {
			return src, nil
		}
	}

	return nil, fmt.Errorf("source not found: %s", path)
}

func countChunks(ctx context.Context, db *database.Queries, srcID pgtype.UUID) int {
	count, _ := db.CountChunksBySource(ctx, srcID)
	return int(count)
}

func showSummary(ctx context.Context, db *database.Queries, srcID pgtype.UUID, logger *slog.Logger) {
	pgID := srcID
	eventCount, _ := db.CountClaimsBySource(ctx, pgID)
	entityCount, _ := db.CountEntitiesBySource(ctx, pgID)
	relationshipCount, _ := db.CountRelationshipsBySource(ctx, pgID)

	logger.Info("extraction summary",
		"events", eventCount,
		"entities", entityCount,
		"relationships", relationshipCount)

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
