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

// EntitiesHandler handles entity requests using the graph model
type EntitiesHandler struct {
	db     *database.Queries
	views  *graph.Views
	logger *slog.Logger
}

// NewEntitiesHandler creates a new graph-based entities handler
func NewEntitiesHandler(db *database.Queries, logger *slog.Logger) *EntitiesHandler {
	viewsService := graph.NewViews(db, logger)

	return &EntitiesHandler{
		db:     db,
		views:  viewsService,
		logger: logger,
	}
}

// Entity represents an entity (matches legacy database.Entity format)
type Entity struct {
	ID                string                 `json:"id"`
	SourceID          string                 `json:"source_id"`
	Name              string                 `json:"name"`
	EntityType        string                 `json:"entity_type"`
	Aliases           []string               `json:"aliases"`
	Description       *string                `json:"description"`
	FirstAppearanceChunk int32                `json:"first_appearance_chunk"`
	LastAppearanceChunk  int32                `json:"last_appearance_chunk"`
	Confidence        float64                `json:"confidence"`
	ReviewStatus      string                 `json:"review_status"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreatedAt         string                 `json:"created_at"`
	UpdatedAt         string                 `json:"updated_at"`
}

// GetEntities handles GET /api/documents/:id/entities using graph model
func (h *EntitiesHandler) GetEntities(w http.ResponseWriter, r *http.Request) {
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

	// Get entities from graph
	graphEntities, err := h.views.GetEntitiesForGraph(r.Context(), parsedUUID)
	if err != nil {
		h.logger.Error("failed to get entities", "error", err)
		http.Error(w, "Failed to get entities", http.StatusInternalServerError)
		return
	}

	// Convert to legacy format
	entities := make([]Entity, len(graphEntities))
	for i, e := range graphEntities {
		entities[i] = Entity{
			ID:                  e.ID,
			SourceID:            idStr,
			Name:                e.Name,
			EntityType:          e.Type,
			Aliases:             e.Aliases,
			Description:         stringPtr(e.Description),
			FirstAppearanceChunk: e.FirstAppearance,
			LastAppearanceChunk:  e.LastAppearance,
			Confidence:          float64(e.Confidence),
			ReviewStatus:        e.ReviewStatus,
			Metadata:            map[string]interface{}{
				"relationship_count": e.RelationshipCount,
				"same_as_edges":      e.SameAsEdges,
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
