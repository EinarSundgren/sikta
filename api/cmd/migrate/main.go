package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/einarsundgren/sikta/internal/config"
	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/graph"
	"github.com/google/uuid"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Connect to database
	db, err := database.Connect(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	queries := database.New(db)
	graphService := graph.NewService(queries, logger)
	migrator := graph.NewMigrator(queries, graphService, logger)

	// Get source ID from argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate <source_id>")
		fmt.Println("\nAvailable sources:")
		sources, _ := queries.ListSources(ctx)
		for _, s := range sources {
			fmt.Printf("  %s: %s\n", database.UUIDStr(s.ID), s.Title)
		}
		os.Exit(1)
	}

	sourceIDStr := os.Args[1]
	sourceID, err := uuid.Parse(sourceIDStr)
	if err != nil {
		logger.Error("invalid source ID", "error", err)
		os.Exit(1)
	}

	// Run migration
	logger.Info("starting migration", "source_id", sourceID)
	err = migrator.MigrateDocument(ctx, sourceID)
	if err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migration complete", "source_id", sourceID)
}
