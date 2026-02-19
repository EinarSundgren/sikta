package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	db             *database.Queries
	extract        *extraction.Service
	dedupe         *extraction.Deduplicator
	chrono         *extraction.ChronologicalEstimator
	logger         *slog.Logger
	progress       chan extraction.ExtractionProgress
	progressTracker *extraction.ProgressTracker
}

// NewExtractionHandler creates a new extraction handler.
func NewExtractionHandler(db *database.Queries, cfg *config.Config, logger *slog.Logger, tracker *extraction.ProgressTracker) *ExtractionHandler {
	claudeClient := claude.NewClient(cfg, logger)
	extractService := extraction.NewService(db, claudeClient, logger, cfg.AnthropicModelExtraction)
	dedupeService := extraction.NewDeduplicator(db, claudeClient, logger, cfg.AnthropicModelClassification)
	chronoService := extraction.NewChronologicalEstimator(db, claudeClient, logger, cfg.AnthropicModelChronology)

	return &ExtractionHandler{
		db:              db,
		extract:         extractService,
		dedupe:          dedupeService,
		chrono:          chronoService,
		logger:          logger,
		progress:        make(chan extraction.ExtractionProgress, 100),
		progressTracker: tracker,
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

// StreamProgress handles GET /api/documents/:id/extract/progress (SSE)
func (h *ExtractionHandler) StreamProgress(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/extract/progress")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}
	sourceID := parsedUUID.String()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Subscribe to progress updates
	ch := h.progressTracker.Subscribe(sourceID)
	defer h.progressTracker.Unsubscribe(sourceID, ch)

	// Send initial state
	initialState := h.progressTracker.Get(sourceID)
	fmt.Fprintf(w, "data: %s\n\n", initialState.ToJSON())
	flusher.Flush()

	// Stream updates
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", event.Payload.ToJSON())
			flusher.Flush()

			// Close connection on complete or error
			if event.Type == "complete" || event.Type == "error" {
				return
			}

		case <-r.Context().Done():
			return
		}
	}
}

// runExtraction runs the full extraction pipeline.
func (h *ExtractionHandler) runExtraction(ctx context.Context, sourceID string) {
	h.logger.Info("starting extraction pipeline", "source_id", sourceID)

	// Get chunk count first
	chunks, err := h.extract.GetChunkCount(ctx, sourceID)
	if err != nil {
		h.logger.Error("failed to get chunks", "error", err)
		h.progressTracker.Error(sourceID, err.Error())
		return
	}
	h.progressTracker.Start(sourceID, chunks)

	var totalEvents, totalEntities, totalRelationships int
	err = h.extract.ExtractDocument(ctx, sourceID, func(progress extraction.ExtractionProgress) {
		if progress.Status == "complete" {
			return
		}
		totalEvents += progress.EventsExtracted
		totalEntities += progress.EntitiesExtracted
		totalRelationships += progress.RelationshipsExtracted

		h.progressTracker.Update(sourceID, progress.ProcessedChunks, totalEvents, totalEntities, totalRelationships)

		h.logger.Info("extraction progress",
			"chunk", progress.ProcessedChunks,
			"of", progress.TotalChunks,
			"events_total", totalEvents,
			"entities_total", totalEntities,
			"relationships_total", totalRelationships,
		)
	})
	if err != nil {
		h.logger.Error("extraction failed", "source_id", sourceID, "error", err)
		h.progressTracker.Error(sourceID, err.Error())
		return
	}

	h.logger.Info("extraction done, starting deduplication", "source_id", sourceID)
	_, err = h.dedupe.DeduplicateEntities(ctx, sourceID)
	if err != nil {
		h.logger.Error("deduplication failed", "source_id", sourceID, "error", err)
	}

	h.logger.Info("deduplication done, estimating chronology", "source_id", sourceID)
	_, err = h.chrono.EstimateChronology(ctx, sourceID)
	if err != nil {
		h.logger.Error("chronology estimation failed", "source_id", sourceID, "error", err)
	}

	h.progressTracker.Complete(sourceID, totalEvents, totalEntities, totalRelationships)

	h.logger.Info("extraction pipeline complete",
		"source_id", sourceID,
		"events_total", totalEvents,
		"entities_total", totalEntities,
		"relationships_total", totalRelationships,
	)
}
