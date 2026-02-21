package graph

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
)

// ReviewHandler handles review status and edit endpoints using the graph model.
// Review status lives on provenance records, not on nodes directly.
type ReviewHandler struct {
	db     *database.Queries
	logger *slog.Logger
}

// NewReviewHandler creates a new graph-based review handler.
func NewReviewHandler(db *database.Queries, logger *slog.Logger) *ReviewHandler {
	return &ReviewHandler{db: db, logger: logger}
}

// reviewStatusRequest is the body for review status updates.
type reviewStatusRequest struct {
	ReviewStatus string `json:"review_status"`
}

// reviewResponse is returned after a review status update.
type reviewResponse struct {
	ID           string `json:"id"`
	ReviewStatus string `json:"review_status"`
}

// UpdateNodeReview handles PATCH /api/claims/{id}/review and PATCH /api/entities/{id}/review.
// Updates all asserted provenance records for the node to the new status.
func (h *ReviewHandler) UpdateNodeReview(w http.ResponseWriter, r *http.Request) {
	nodeID, ok := parseUUIDFromPath(r)
	if !ok {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
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

	err := h.db.UpdateProvenanceStatusByTarget(r.Context(), database.UpdateProvenanceStatusByTargetParams{
		TargetType: "node",
		TargetID:   database.PgUUID(nodeID),
		Status:     req.ReviewStatus,
	})
	if err != nil {
		h.logger.Error("failed to update node review status", "id", nodeID, "error", err)
		http.Error(w, "Failed to update review status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviewResponse{
		ID:           nodeID.String(),
		ReviewStatus: req.ReviewStatus,
	})
}

// UpdateEdgeReview handles PATCH /api/relationships/{id}/review.
// Updates all asserted provenance records for the edge to the new status.
func (h *ReviewHandler) UpdateEdgeReview(w http.ResponseWriter, r *http.Request) {
	edgeID, ok := parseUUIDFromPath(r)
	if !ok {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
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

	err := h.db.UpdateProvenanceStatusByTarget(r.Context(), database.UpdateProvenanceStatusByTargetParams{
		TargetType: "edge",
		TargetID:   database.PgUUID(edgeID),
		Status:     req.ReviewStatus,
	})
	if err != nil {
		h.logger.Error("failed to update edge review status", "id", edgeID, "error", err)
		http.Error(w, "Failed to update review status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviewResponse{
		ID:           edgeID.String(),
		ReviewStatus: req.ReviewStatus,
	})
}

// updateNodeDataRequest is the body for editing a claim node.
type updateNodeDataRequest struct {
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	DateText    *string  `json:"date_text"`
	EventType   *string  `json:"event_type"`
	Confidence  *float64 `json:"confidence"`
}

// updateNodeDataResponse is the response after editing a claim node.
type updateNodeDataResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	EventType    string  `json:"event_type"`
	Confidence   float64 `json:"confidence"`
	ReviewStatus string  `json:"review_status"`
}

// UpdateNodeData handles PATCH /api/claims/{id}.
// Updates node label and properties, then marks provenance as edited.
func (h *ReviewHandler) UpdateNodeData(w http.ResponseWriter, r *http.Request) {
	nodeID, ok := parseUUIDFromPath(r)
	if !ok {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req updateNodeDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	// Get current node to preserve existing properties
	node, err := h.db.GetNode(r.Context(), database.PgUUID(nodeID))
	if err != nil {
		h.logger.Error("failed to get node", "id", nodeID, "error", err)
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	// Merge updated fields into current properties
	description := node.GetProperty("description", "").(string)
	if req.Description != nil {
		description = *req.Description
	}
	eventType := node.GetProperty("event_type", "").(string)
	if req.EventType != nil {
		eventType = *req.EventType
	}

	// Build updated properties, preserving any fields we don't touch
	var currentProps map[string]interface{}
	if node.Properties != nil {
		json.Unmarshal(node.Properties, &currentProps)
	}
	if currentProps == nil {
		currentProps = make(map[string]interface{})
	}
	currentProps["description"] = description
	currentProps["event_type"] = eventType

	propsBytes, err := json.Marshal(currentProps)
	if err != nil {
		http.Error(w, "Failed to encode properties", http.StatusInternalServerError)
		return
	}

	// Update the node
	updatedNode, err := h.db.UpdateNode(r.Context(), database.UpdateNodeParams{
		ID:         node.ID,
		Label:      req.Title,
		Properties: propsBytes,
	})
	if err != nil {
		h.logger.Error("failed to update node", "id", nodeID, "error", err)
		http.Error(w, "Failed to update node", http.StatusInternalServerError)
		return
	}

	// Mark asserted provenance as edited
	_ = h.db.UpdateProvenanceStatusByTarget(r.Context(), database.UpdateProvenanceStatusByTargetParams{
		TargetType: "node",
		TargetID:   updatedNode.ID,
		Status:     "edited",
	})

	confidence := 0.8
	if req.Confidence != nil {
		confidence = *req.Confidence
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updateNodeDataResponse{
		ID:           nodeID.String(),
		Title:        updatedNode.Label,
		Description:  description,
		EventType:    eventType,
		Confidence:   confidence,
		ReviewStatus: "edited",
	})
}

// ReviewProgress holds review counts per item type (matches legacy format for frontend compat).
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

// GetReviewProgress handles GET /api/documents/{id}/review-progress.
// Counts provenance records by status for claim and entity nodes.
func (h *ReviewHandler) GetReviewProgress(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	sourceID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Find the document node for this legacy source ID
	docNode, err := h.db.GetDocumentNodeByLegacySourceID(r.Context(), sourceID.String())
	if err != nil {
		h.logger.Error("document node not found", "source_id", sourceID, "error", err)
		http.Error(w, "Document not found in graph", http.StatusNotFound)
		return
	}

	claimRows, err := h.db.CountClaimProvenanceByStatusForSource(r.Context(), docNode.ID)
	if err != nil {
		h.logger.Error("failed to count claim provenance", "error", err)
		http.Error(w, "Failed to get review progress", http.StatusInternalServerError)
		return
	}

	entityRows, err := h.db.CountEntityProvenanceByStatusForSource(r.Context(), docNode.ID)
	if err != nil {
		h.logger.Error("failed to count entity provenance", "error", err)
		http.Error(w, "Failed to get review progress", http.StatusInternalServerError)
		return
	}

	progress := ReviewProgress{}
	for _, row := range claimRows {
		n := int(row.Count)
		progress.Claims.Total += n
		switch row.Status {
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
		switch row.Status {
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

// parseUUIDFromPath extracts the {id} path variable from the request.
func parseUUIDFromPath(r *http.Request) (uuid.UUID, bool) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
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
