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

// InconsistencyHandler handles inconsistency-related HTTP requests.
type InconsistencyHandler struct {
	db       *database.Queries
	detector *extraction.InconsistencyDetector
	logger   *slog.Logger
}

// NewInconsistencyHandler creates a new inconsistency handler.
func NewInconsistencyHandler(db *database.Queries, cfg *config.Config, logger *slog.Logger) *InconsistencyHandler {
	claudeClient := claude.NewClient(cfg, logger)
	detector := extraction.NewInconsistencyDetector(db, claudeClient, logger, cfg.AnthropicModelExtraction)

	return &InconsistencyHandler{
		db:       db,
		detector: detector,
		logger:   logger,
	}
}

// GetInconsistencies handles GET /api/documents/:id/inconsistencies
func (h *InconsistencyHandler) GetInconsistencies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/inconsistencies")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	incs, err := h.db.ListInconsistenciesBySource(r.Context(), database.PgUUID(parsedUUID))
	if err != nil {
		h.logger.Error("failed to get inconsistencies", "error", err)
		http.Error(w, "Failed to get inconsistencies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incs)
}

// GetInconsistencyItems handles GET /api/inconsistencies/:id/items
func (h *InconsistencyHandler) GetInconsistencyItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/inconsistencies/")
	idStr = strings.TrimSuffix(idStr, "/items")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid inconsistency ID", http.StatusBadRequest)
		return
	}

	items, err := h.db.ListInconsistencyItems(r.Context(), database.PgUUID(parsedUUID))
	if err != nil {
		h.logger.Error("failed to get inconsistency items", "error", err)
		http.Error(w, "Failed to get inconsistency items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// TriggerInconsistencyDetection handles POST /api/documents/:id/detect-inconsistencies
func (h *InconsistencyHandler) TriggerInconsistencyDetection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/detect-inconsistencies")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	go h.runDetection(context.Background(), parsedUUID.String())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "Inconsistency detection started",
	})
}

// ResolveInconsistency handles PUT /api/inconsistencies/:id/resolve
func (h *InconsistencyHandler) ResolveInconsistency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/inconsistencies/")
	idStr = strings.TrimSuffix(idStr, "/resolve")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid inconsistency ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"` // resolved, noted, dismissed
		Note   string `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status != "resolved" && req.Status != "noted" && req.Status != "dismissed" {
		http.Error(w, "Invalid status. Must be: resolved, noted, or dismissed", http.StatusBadRequest)
		return
	}

	updated, err := h.db.UpdateInconsistencyResolution(r.Context(), database.UpdateInconsistencyResolutionParams{
		ID:               database.PgUUID(parsedUUID),
		ResolutionStatus: req.Status,
		ResolutionNote:   database.PgText(req.Note),
	})
	if err != nil {
		h.logger.Error("failed to update inconsistency", "error", err)
		http.Error(w, "Failed to update inconsistency", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// runDetection runs the full inconsistency detection pipeline
func (h *InconsistencyHandler) runDetection(ctx context.Context, sourceID string) {
	h.logger.Info("starting inconsistency detection", "source_id", sourceID)

	inconsistencies, err := h.detector.DetectAll(ctx, sourceID)
	if err != nil {
		h.logger.Error("inconsistency detection failed", "source_id", sourceID, "error", err)
		return
	}

	h.logger.Info("inconsistency detection complete",
		"source_id", sourceID,
		"total", len(inconsistencies))
}
