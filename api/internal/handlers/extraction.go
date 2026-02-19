package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/einarsundgren/sikta/internal/config"
	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
	"github.com/google/uuid"
)

// ExtractionHandler handles extraction-related HTTP requests.
type ExtractionHandler struct {
	db       *database.Queries
	extract  *extraction.Service
	dedupe   *extraction.Deduplicator
	chrono   *extraction.ChronologicalEstimator
	logger   *slog.Logger
	progress chan extraction.ExtractionProgress
}

// NewExtractionHandler creates a new extraction handler.
func NewExtractionHandler(db *database.Queries, cfg *config.Config, logger *slog.Logger) *ExtractionHandler {
	claudeClient := claude.NewClient(cfg, logger)
	extractService := extraction.NewService(db, claudeClient, logger, cfg.AnthropicModelExtraction)
	dedupeService := extraction.NewDeduplicator(db, claudeClient, logger, cfg.AnthropicModelClassification)
	chronoService := extraction.NewChronologicalEstimator(db, claudeClient, logger, cfg.AnthropicModelChronology)

	return &ExtractionHandler{
		db:       db,
		extract:  extractService,
		dedupe:   dedupeService,
		chrono:   chronoService,
		logger:   logger,
		progress: make(chan extraction.ExtractionProgress, 100),
	}
}

// GetEvents handles GET /api/documents/:id/events
func (h *ExtractionHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/events")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	claims, err := h.db.ListClaimsBySource(r.Context(), database.PgUUID(parsedUUID))
	if err != nil {
		h.logger.Error("failed to get claims", "error", err)
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claims)
}

// GetEntities handles GET /api/documents/:id/entities
func (h *ExtractionHandler) GetEntities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/entities")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	entities, err := h.db.ListEntitiesBySource(r.Context(), database.PgUUID(parsedUUID))
	if err != nil {
		h.logger.Error("failed to get entities", "error", err)
		http.Error(w, "Failed to get entities", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}

// GetRelationships handles GET /api/documents/:id/relationships
func (h *ExtractionHandler) GetRelationships(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/relationships")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	relationships, err := h.db.ListRelationshipsBySource(r.Context(), database.PgUUID(parsedUUID))
	if err != nil {
		h.logger.Error("failed to get relationships", "error", err)
		http.Error(w, "Failed to get relationships", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationships)
}

// TriggerExtraction handles POST /api/documents/:id/extract
func (h *ExtractionHandler) TriggerExtraction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/extract")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	go h.runExtraction(context.Background(), parsedUUID.String())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "Extraction started",
	})
}

// GetExtractionStatus handles GET /api/documents/:id/extract/status
func (h *ExtractionHandler) GetExtractionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/extract/status")
	idStr = strings.TrimSuffix(idStr, "/status")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	pgID := database.PgUUID(parsedUUID)
	eventCount, _ := h.db.CountClaimsBySource(r.Context(), pgID)
	entityCount, _ := h.db.CountEntitiesBySource(r.Context(), pgID)
	relationshipCount, _ := h.db.CountRelationshipsBySource(r.Context(), pgID)
	chunkCount, _ := h.db.CountChunksBySource(r.Context(), pgID)

	status := map[string]interface{}{
		"status":        "processing",
		"total_chunks":  chunkCount,
		"events":        eventCount,
		"entities":      entityCount,
		"relationships": relationshipCount,
	}

	if eventCount > 0 && chunkCount > 0 {
		status["status"] = "complete"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// runExtraction runs the full extraction pipeline.
func (h *ExtractionHandler) runExtraction(ctx context.Context, sourceID string) {
	h.logger.Info("starting extraction pipeline", "source_id", sourceID)

	err := h.extract.ExtractDocument(ctx, sourceID, func(progress extraction.ExtractionProgress) {
		h.logger.Info("extraction progress",
			"source_id", sourceID,
			"chunk", progress.CurrentChunk,
			"total", progress.TotalChunks,
			"events", progress.EventsExtracted,
			"entities", progress.EntitiesExtracted,
			"relationships", progress.RelationshipsExtracted)
	})
	if err != nil {
		h.logger.Error("extraction failed", "source_id", sourceID, "error", err)
		return
	}

	_, err = h.dedupe.DeduplicateEntities(ctx, sourceID)
	if err != nil {
		h.logger.Error("deduplication failed", "source_id", sourceID, "error", err)
	}

	_, err = h.chrono.EstimateChronology(ctx, sourceID)
	if err != nil {
		h.logger.Error("chronology estimation failed", "source_id", sourceID, "error", err)
	}

	h.logger.Info("extraction pipeline complete", "source_id", sourceID)
}
