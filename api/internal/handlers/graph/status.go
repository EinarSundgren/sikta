package graph

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
)

// StatusHandler handles extraction status requests using the graph model
type StatusHandler struct {
	db     *database.Queries
	logger *slog.Logger
}

// NewStatusHandler creates a new graph-based status handler
func NewStatusHandler(db *database.Queries, logger *slog.Logger) *StatusHandler {
	return &StatusHandler{
		db:     db,
		logger: logger,
	}
}

// GetExtractionStatus handles GET /api/documents/:id/extract/status using graph model
func (h *StatusHandler) GetExtractionStatus(w http.ResponseWriter, r *http.Request) {
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

	// Count chunks for this source
	chunkCount, _ := h.db.CountChunksBySource(r.Context(), pgID)

	// For graph model, we need to find the document node first
	// then count its associated nodes
	eventCount := int32(0)
	entityCount := int32(0)
	relationshipCount := int32(0)

	// Find document node by source_id in properties
	allDocNodes, _ := h.db.ListNodesByType(r.Context(), database.ListNodesByTypeParams{
		NodeType: "document",
		Limit:    100,
	})

	var docNode *database.Node
	for _, node := range allDocNodes {
		if node.GetProperty("source_id", "") == parsedUUID.String() {
			docNode = node
			break
		}
	}

	if docNode != nil {
		// Get all nodes for this document and count by type
		allNodes, err := h.db.ListNodesBySource(r.Context(), docNode.ID)
		if err == nil {
			for _, node := range allNodes {
				switch node.NodeType {
				case "event", "attribute", "relation":
					eventCount++
				case "person", "place", "organization", "object":
					entityCount++
				}
			}
		}

		// Count edges for relationships
		edges, err := h.db.ListEdgesBySource(r.Context(), docNode.ID)
		if err == nil {
			relationshipCount = int32(len(edges))
		}
	}

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
