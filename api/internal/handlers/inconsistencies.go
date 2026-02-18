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

	// Extract document ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/inconsistencies")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get inconsistencies
	incs, err := h.db.ListInconsistenciesByDocument(r.Context(), database.UUID(parsedUUID))
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

	// Extract inconsistency ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/inconsistencies/")
	idStr = strings.TrimSuffix(idStr, "/items")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid inconsistency ID", http.StatusBadRequest)
		return
	}

	// Get inconsistency items with event/entity details
	items, err := h.db.ListInconsistencyItems(r.Context(), database.UUID(parsedUUID))
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

	// Extract document ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/detect-inconsistencies")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Run inconsistency detection in background
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

	// Extract inconsistency ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/inconsistencies/")
	idStr = strings.TrimSuffix(idStr, "/resolve")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid inconsistency ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Status string `json:"status"` // resolved, noted, dismissed
		Note   string `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status
	if req.Status != "resolved" && req.Status != "noted" && req.Status != "dismissed" {
		http.Error(w, "Invalid status. Must be: resolved, noted, or dismissed", http.StatusBadRequest)
		return
	}

	// Update inconsistency
	updated, err := h.db.UpdateInconsistencyResolution(r.Context(), database.UpdateInconsistencyResolutionParams{
		ID:             database.UUID(parsedUUID),
		ResolutionStatus: req.Status,
		ResolutionNote:   req.Note,
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
func (h *InconsistencyHandler) runDetection(ctx context.Context, documentID string) {
	h.logger.Info("starting inconsistency detection", "document_id", documentID)

	// Detect all types of inconsistencies
	inconsistencies, err := h.detector.DetectAll(ctx, documentID)
	if err != nil {
		h.logger.Error("inconsistency detection failed", "document_id", documentID, "error", err)
		return
	}

	// Run LLM-based contradiction detection (optional, more expensive)
	// contradictions, err := h.detector.DetectContradictionsWithLLM(ctx, documentID)
	// if err != nil {
	// 	h.logger.Error("LLM contradiction detection failed", "document_id", documentID, "error", err)
	// }

	h.logger.Info("inconsistency detection complete",
		"document_id", documentID,
		"total", len(inconsistencies))
}
