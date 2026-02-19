package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
)

// ReviewHandler handles review status and edit endpoints.
type ReviewHandler struct {
	db     *database.Queries
	logger *slog.Logger
}

// NewReviewHandler creates a new review handler.
func NewReviewHandler(db *database.Queries, logger *slog.Logger) *ReviewHandler {
	return &ReviewHandler{db: db, logger: logger}
}

// reviewStatusRequest is the body for review status updates.
type reviewStatusRequest struct {
	ReviewStatus string `json:"review_status"`
}

// UpdateClaimReview handles PATCH /api/claims/{id}/review
func (h *ReviewHandler) UpdateClaimReview(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(r.URL.Path, "/api/claims/", "/review")
	if !ok {
		http.Error(w, "Invalid claim ID", http.StatusBadRequest)
		return
	}

	var req reviewStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !validReviewStatus(req.ReviewStatus) {
		http.Error(w, "Invalid review_status", http.StatusBadRequest)
		return
	}

	claim, err := h.db.UpdateClaimReviewStatus(r.Context(), database.UpdateClaimReviewStatusParams{
		ID:           database.PgUUID(id),
		ReviewStatus: req.ReviewStatus,
	})
	if err != nil {
		h.logger.Error("failed to update claim review status", "id", id, "error", err)
		http.Error(w, "Failed to update review status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claim)
}

// updateClaimDataRequest is the body for editing a claim.
type updateClaimDataRequest struct {
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	DateText    *string  `json:"date_text"`
	EventType   *string  `json:"event_type"`
	Confidence  *float64 `json:"confidence"`
}

// UpdateClaim handles PATCH /api/claims/{id}
func (h *ReviewHandler) UpdateClaim(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(r.URL.Path, "/api/claims/", "")
	if !ok {
		http.Error(w, "Invalid claim ID", http.StatusBadRequest)
		return
	}

	var req updateClaimDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	confidence := float32(0.8)
	if req.Confidence != nil {
		confidence = float32(*req.Confidence)
	}

	claim, err := h.db.UpdateClaimData(r.Context(), database.UpdateClaimDataParams{
		ID:          database.PgUUID(id),
		Title:       req.Title,
		Description: database.PgTextPtr(req.Description),
		DateText:    database.PgTextPtr(req.DateText),
		EventType:   database.PgTextPtr(req.EventType),
		Confidence:  confidence,
	})
	if err != nil {
		h.logger.Error("failed to update claim", "id", id, "error", err)
		http.Error(w, "Failed to update claim", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claim)
}

// UpdateEntityReview handles PATCH /api/entities/{id}/review
func (h *ReviewHandler) UpdateEntityReview(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(r.URL.Path, "/api/entities/", "/review")
	if !ok {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	var req reviewStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !validReviewStatus(req.ReviewStatus) {
		http.Error(w, "Invalid review_status", http.StatusBadRequest)
		return
	}

	entity, err := h.db.UpdateEntityReviewStatus(r.Context(), database.UpdateEntityReviewStatusParams{
		ID:           database.PgUUID(id),
		ReviewStatus: req.ReviewStatus,
	})
	if err != nil {
		h.logger.Error("failed to update entity review status", "id", id, "error", err)
		http.Error(w, "Failed to update review status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// UpdateRelationshipReview handles PATCH /api/relationships/{id}/review
func (h *ReviewHandler) UpdateRelationshipReview(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(r.URL.Path, "/api/relationships/", "/review")
	if !ok {
		http.Error(w, "Invalid relationship ID", http.StatusBadRequest)
		return
	}

	var req reviewStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !validReviewStatus(req.ReviewStatus) {
		http.Error(w, "Invalid review_status", http.StatusBadRequest)
		return
	}

	rel, err := h.db.UpdateRelationshipReviewStatus(r.Context(), database.UpdateRelationshipReviewStatusParams{
		ID:           database.PgUUID(id),
		ReviewStatus: req.ReviewStatus,
	})
	if err != nil {
		h.logger.Error("failed to update relationship review status", "id", id, "error", err)
		http.Error(w, "Failed to update review status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rel)
}

// ReviewProgress holds review counts per item type.
type ReviewProgress struct {
	Claims        ReviewCounts `json:"claims"`
	Entities      ReviewCounts `json:"entities"`
	TotalReviewed int          `json:"total_reviewed"`
	TotalItems    int          `json:"total_items"`
}

// ReviewCounts holds counts per review status.
type ReviewCounts struct {
	Total    int `json:"total"`
	Pending  int `json:"pending"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
	Edited   int `json:"edited"`
}

// GetReviewProgress handles GET /api/documents/{id}/review-progress
func (h *ReviewHandler) GetReviewProgress(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/review-progress")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	sourceID := database.PgUUID(parsedUUID)

	claimRows, err := h.db.CountClaimsByStatusForSource(r.Context(), sourceID)
	if err != nil {
		h.logger.Error("failed to count claims by status", "error", err)
		http.Error(w, "Failed to get review progress", http.StatusInternalServerError)
		return
	}

	entityRows, err := h.db.CountEntitiesByStatusForSource(r.Context(), sourceID)
	if err != nil {
		h.logger.Error("failed to count entities by status", "error", err)
		http.Error(w, "Failed to get review progress", http.StatusInternalServerError)
		return
	}

	progress := ReviewProgress{}
	for _, row := range claimRows {
		n := int(row.Count)
		progress.Claims.Total += n
		switch row.ReviewStatus {
		case "pending":
			progress.Claims.Pending = n
		case "approved":
			progress.Claims.Approved = n
		case "rejected":
			progress.Claims.Rejected = n
		case "edited":
			progress.Claims.Edited = n
		}
	}
	for _, row := range entityRows {
		n := int(row.Count)
		progress.Entities.Total += n
		switch row.ReviewStatus {
		case "pending":
			progress.Entities.Pending = n
		case "approved":
			progress.Entities.Approved = n
		case "rejected":
			progress.Entities.Rejected = n
		case "edited":
			progress.Entities.Edited = n
		}
	}

	progress.TotalItems = progress.Claims.Total + progress.Entities.Total
	reviewed := progress.Claims.Approved + progress.Claims.Rejected + progress.Claims.Edited +
		progress.Entities.Approved + progress.Entities.Rejected + progress.Entities.Edited
	progress.TotalReviewed = reviewed

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

// parseIDFromPath extracts a UUID from a URL path, trimming prefix and suffix.
// If suffix is empty string, trims nothing after the prefix.
func parseIDFromPath(path, prefix, suffix string) (uuid.UUID, bool) {
	s := strings.TrimPrefix(path, prefix)
	if suffix != "" {
		s = strings.TrimSuffix(s, suffix)
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, false
	}
	return id, true
}

// validReviewStatus checks if a review status value is allowed.
func validReviewStatus(s string) bool {
	switch s {
	case "pending", "approved", "rejected", "edited":
		return true
	}
	return false
}

