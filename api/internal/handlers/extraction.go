package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/einarsundgren/sikta/internal/config"
	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
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
	extractService := extraction.NewService(db, claudeClient, logger)
	dedupeService := extraction.NewDeduplicator(db, claudeClient, logger)
	chronoService := extraction.NewChronologicalEstimator(db, claudeClient, logger)

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

	// Extract document ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/events")

	docID, err := parseUUID(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get events
	events, err := h.db.ListEventsByDocument(r.Context(), database.UUID(docID))
	if err != nil {
		h.logger.Error("failed to get events", "error", err)
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// GetEntities handles GET /api/documents/:id/entities
func (h *ExtractionHandler) GetEntities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract document ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/entities")

	docID, err := parseUUID(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get entities
	entities, err := h.db.ListEntitiesByDocument(r.Context(), database.UUID(docID))
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

	// Extract document ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/relationships")

	docID, err := parseUUID(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get relationships
	relationships, err := h.db.ListRelationshipsByDocument(r.Context(), database.UUID(docID))
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

	// Extract document ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/extract")

	docID, err := parseUUID(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Start extraction in background
	go h.runExtraction(r.Context(), docID.String())

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

	// Extract document ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/extract/status")
	idStr = strings.TrimSuffix(idStr, "/status")

	docID, err := parseUUID(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get extraction counts
	eventCount, _ := h.db.CountEventsByDocument(r.Context(), database.UUID(docID))
	entityCount, _ := h.db.CountEntitiesByDocument(r.Context(), database.UUID(docID))
	relationshipCount, _ := h.db.CountRelationshipsByDocument(r.Context(), database.UUID(docID))
	chunkCount, _ := h.db.CountChunksByDocument(r.Context(), database.UUID(docID))

	status := map[string]interface{}{
		"status":       "processing",
		"total_chunks": chunkCount,
		"events":       eventCount,
		"entities":     entityCount,
		"relationships": relationshipCount,
	}

	// Check if extraction is complete
	if eventCount > 0 && chunkCount > 0 {
		// Rough heuristic: if we have events, assume complete
		status["status"] = "complete"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// runExtraction runs the full extraction pipeline.
func (h *ExtractionHandler) runExtraction(ctx context.Context, documentID string) {
	h.logger.Info("starting extraction pipeline", "document_id", documentID)

	// Step 1: Extract events, entities, relationships
	err := h.extract.ExtractDocument(ctx, documentID, func(progress extraction.ExtractionProgress) {
		h.logger.Info("extraction progress",
			"document_id", documentID,
			"chunk", progress.CurrentChunk,
			"total", progress.TotalChunks,
			"events", progress.EventsExtracted,
			"entities", progress.EntitiesExtracted,
			"relationships", progress.RelationshipsExtracted)
	})
	if err != nil {
		h.logger.Error("extraction failed", "document_id", documentID, "error", err)
		return
	}

	// Step 2: Deduplicate entities
	_, err = h.dedupe.DeduplicateEntities(ctx, documentID)
	if err != nil {
		h.logger.Error("deduplication failed", "document_id", documentID, "error", err)
		// Continue anyway
	}

	// Step 3: Estimate chronology
	_, err = h.chronology.EstimateChronology(ctx, documentID)
	if err != nil {
		h.logger.Error("chronology estimation failed", "document_id", documentID, "error", err)
		// Continue anyway
	}

	h.logger.Info("extraction pipeline complete", "document_id", documentID)
}

// parseUUID parses a UUID from string.
func parseUUID(s string) (string, error) {
	// Basic UUID validation
	if len(s) != 36 {
		return "", http.ErrNotSupported
	}
	return s, nil
}
