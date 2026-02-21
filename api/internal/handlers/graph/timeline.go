package graph

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/graph"
	"github.com/google/uuid"
)

// TimelineHandler handles timeline requests using the graph model
type TimelineHandler struct {
	db     *database.Queries
	graph  *graph.Service
	views  *graph.Views
	logger *slog.Logger
}

// NewTimelineHandler creates a new graph-based timeline handler
func NewTimelineHandler(db *database.Queries, logger *slog.Logger) *TimelineHandler {
	graphService := graph.NewService(db, logger)
	viewsService := graph.NewViews(db, logger)

	return &TimelineHandler{
		db:     db,
		graph:  graphService,
		views:  viewsService,
		logger: logger,
	}
}

// TimelineEvent represents a claim on the timeline (matches legacy format)
type TimelineEvent struct {
	ID                string                  `json:"id"`
	SourceID          string                  `json:"source_id"`
	Title             string                  `json:"title"`
	Description       *string                 `json:"description"`
	EventType         *string                 `json:"event_type"`
	DateText          *string                 `json:"date_text"`
	DateStart         *string                 `json:"date_start"`
	DateEnd           *string                 `json:"date_end"`
	DatePrecision     *string                 `json:"date_precision"`
	NarrativePosition int32                   `json:"narrative_position"`
	ChronologicalPos  *int32                  `json:"chronological_position"`
	Confidence        float64                 `json:"confidence"`
	ConfidenceReason  *string                 `json:"confidence_reason"`
	ReviewStatus      string                  `json:"review_status"`
	Metadata          map[string]interface{}  `json:"metadata"`
	Entities          []TimelineEntity        `json:"entities"`
	SourceReferences  []TimelineSource        `json:"source_references"`
	Inconsistencies   []TimelineInconsistency `json:"inconsistencies"`
	ChapterTitle      *string                 `json:"chapter_title"`
	ChapterNumber     *int32                  `json:"chapter_number"`
}

// TimelineEntity represents an entity connected to a claim
type TimelineEntity struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	EntityType string  `json:"entity_type"`
	Role       *string `json:"role"`
}

// TimelineSource represents a source reference for a claim
type TimelineSource struct {
	ID            string  `json:"id"`
	ChunkID       string  `json:"chunk_id"`
	Excerpt       string  `json:"excerpt"`
	ChapterTitle  *string `json:"chapter_title"`
	ChapterNumber *int32  `json:"chapter_number"`
	PageStart     *int32  `json:"page_start"`
	PageEnd       *int32  `json:"page_end"`
}

// TimelineInconsistency represents an inconsistency linked to a claim
type TimelineInconsistency struct {
	ID                string `json:"id"`
	InconsistencyType string `json:"inconsistency_type"`
	Severity          string `json:"severity"`
	Title             string `json:"title"`
}

// GetTimeline handles GET /api/documents/:id/timeline using graph model
func (h *TimelineHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/timeline")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Use view strategy from query param (default: trust_weighted)
	strategy := graph.ViewStrategy(r.URL.Query().Get("view_strategy"))
	if strategy == "" {
		strategy = graph.ViewStrategyTrustWeighted
	}

	// Get events from graph using view strategy
	events, err := h.views.GetEventsForTimeline(r.Context(), parsedUUID, strategy)
	if err != nil {
		h.logger.Error("failed to get timeline events", "error", err)
		http.Error(w, "Failed to get timeline", http.StatusInternalServerError)
		return
	}

	// Convert to legacy format
	timelineEvents := make([]TimelineEvent, len(events))
	for i, event := range events {
		timelineEvents[i] = TimelineEvent{
			ID:                event.ID,
			SourceID:          idStr,
			Title:             event.Title,
			Description:       stringPtr(event.Description),
			EventType:         stringPtr(event.Type),
			DateText:          stringPtr(event.DateText),
			DateStart:         timeStringPtr(event.DateStart),
			DateEnd:           timeStringPtr(event.DateEnd),
			DatePrecision:     nil, // Would come from node properties
			NarrativePosition: event.NarrativePosition,
			ChronologicalPos:  int32Ptr(event.ChronologicalPosition),
			Confidence:        float64(event.Confidence),
			ReviewStatus:      event.ReviewStatus,
			Entities:          convertEntities(event.Entities),
			SourceReferences:  []TimelineSource{}, // TODO: populate from provenance
			Inconsistencies:   []TimelineInconsistency{}, // TODO: map from inconsistency nodes
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timelineEvents)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timeStringPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func convertEntities(entities []graph.TimelineEntity) []TimelineEntity {
	result := make([]TimelineEntity, len(entities))
	for i, e := range entities {
		result[i] = TimelineEntity{
			ID:         e.ID,
			Name:       e.Name,
			EntityType: e.Type,
			Role:       stringPtr(e.Role),
		}
	}
	return result
}
