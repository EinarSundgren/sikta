package graph

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/graph"
	"github.com/google/uuid"
)

// RelationshipsHandler handles relationship requests using the graph model
type RelationshipsHandler struct {
	db     *database.Queries
	views  *graph.Views
	logger *slog.Logger
}

// NewRelationshipsHandler creates a new graph-based relationships handler
func NewRelationshipsHandler(db *database.Queries, logger *slog.Logger) *RelationshipsHandler {
	viewsService := graph.NewViews(db, logger)

	return &RelationshipsHandler{
		db:     db,
		views:  viewsService,
		logger: logger,
	}
}

// Relationship represents a relationship (matches legacy database.Relationship format)
type Relationship struct {
	ID               string         `json:"id"`
	SourceID         string         `json:"source_id"`
	EntityAID        string         `json:"entity_a_id"`
	EntityBID        string         `json:"entity_b_id"`
	RelationshipType string         `json:"relationship_type"`
	Description      *string        `json:"description"`
	StartClaimID     *string        `json:"start_claim_id"`
	EndClaimID       *string        `json:"end_claim_id"`
	Confidence       float64        `json:"confidence"`
	ReviewStatus     string         `json:"review_status"`
	Metadata         map[string]any `json:"metadata"`
}

// GetRelationships handles GET /api/documents/:id/relationships using graph model
func (h *RelationshipsHandler) GetRelationships(w http.ResponseWriter, r *http.Request) {
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

	// Get relationships from graph
	graphRelationships, err := h.views.GetRelationshipsForGraph(r.Context(), parsedUUID)
	if err != nil {
		h.logger.Error("failed to get relationships", "error", err)
		http.Error(w, "Failed to get relationships", http.StatusInternalServerError)
		return
	}

	// Convert to legacy format
	relationships := make([]Relationship, len(graphRelationships))
	for i, rel := range graphRelationships {
		relationships[i] = Relationship{
			ID:               rel.ID,
			SourceID:         idStr,
			EntityAID:        rel.EntityAID,
			EntityBID:        rel.EntityBID,
			RelationshipType: rel.RelationshipType,
			Description:      stringPtr(rel.Description),
			StartClaimID:     rel.StartClaimID,
			EndClaimID:       rel.EndClaimID,
			Confidence:       float64(rel.Confidence),
			ReviewStatus:     rel.ReviewStatus,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationships)
}
