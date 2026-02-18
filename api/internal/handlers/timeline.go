package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
	"github.com/einarsundgren/sikta/internal/database"
)

// TimelineHandler handles timeline-related HTTP requests.
type TimelineHandler struct {
	db     *database.Queries
	logger *slog.Logger
}

// NewTimelineHandler creates a new timeline handler.
func NewTimelineHandler(db *database.Queries, logger *slog.Logger) *TimelineHandler {
	return &TimelineHandler{
		db:     db,
		logger: logger,
	}
}

// TimelineEvent represents an event on the timeline with all related data.
type TimelineEvent struct {
	ID                  string                 `json:"id"`
	DocumentID          string                 `json:"document_id"`
	Title               string                 `json:"title"`
	Description         *string                `json:"description"`
	EventType           string                 `json:"event_type"`
	DateText            *string                `json:"date_text"`
	DateStart           *string                `json:"date_start"`
	DateEnd             *string                `json:"date_end"`
	DatePrecision       *string                `json:"date_precision"`
	NarrativePosition   int32                  `json:"narrative_position"`
	ChronologicalPos    *int32                 `json:"chronological_position"`
	Confidence          float64                `json:"confidence"`
	ConfidenceReason    *string                `json:"confidence_reason"`
	ReviewStatus        string                 `json:"review_status"`
	Metadata            map[string]interface{} `json:"metadata"`
	Entities            []TimelineEntity       `json:"entities"`
	SourceReferences     []TimelineSource       `json:"source_references"`
	Inconsistencies     []TimelineInconsistency `json:"inconsistencies"`
	ChapterTitle        *string                `json:"chapter_title"`
	ChapterNumber       *int32                 `json:"chapter_number"`
}

// TimelineEntity represents an entity connected to an event.
type TimelineEntity struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	EntityType  string  `json:"entity_type"`
	Role        *string `json:"role"`
}

// TimelineSource represents a source reference for an event.
type TimelineSource struct {
	ID          string  `json:"id"`
	ChunkID     string  `json:"chunk_id"`
	Excerpt     string  `json:"excerpt"`
	ChapterTitle *string `json:"chapter_title"`
	ChapterNumber *int32 `json:"chapter_number"`
	PageStart   *int32  `json:"page_start"`
	PageEnd     *int32  `json:"page_end"`
}

// TimelineInconsistency represents an inconsistency linked to an event.
type TimelineInconsistency struct {
	ID               string `json:"id"`
	InconsistencyType string `json:"inconsistency_type"`
	Severity         string `json:"severity"`
	Title            string `json:"title"`
}

// GetTimeline handles GET /api/documents/:id/timeline
func (h *TimelineHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract document ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	idStr = strings.TrimSuffix(idStr, "/timeline")

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get all events for the document
	events, err := h.db.ListEventsByDocument(r.Context(), database.UUID(parsedUUID))
	if err != nil {
		h.logger.Error("failed to get events", "error", err)
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	// Convert to timeline events with related data
	timelineEvents := make([]TimelineEvent, 0, len(events))
	for _, event := range events {
		timelineEvent := TimelineEvent{
			ID:                event.ID.String(),
			DocumentID:        event.DocumentID.String(),
			Title:             event.Title,
			Description:       event.Description,
			EventType:         event.EventType,
			DateText:          event.DateText,
			DateStart:         dateString(event.DateStart),
			DateEnd:           dateString(event.DateEnd),
			DatePrecision:     event.DatePrecision,
			NarrativePosition: event.NarrativePosition,
			ChronologicalPos:  event.ChronologicalPosition,
			Confidence:        event.Confidence,
			ConfidenceReason:  event.ConfidenceReason,
			ReviewStatus:      event.ReviewStatus,
			// ChapterTitle and ChapterNumber not in Event struct yet
			// Will be added when we link events to chunks
			Entities:        []TimelineEntity{},
			SourceReferences: []TimelineSource{},
		}

		// Parse metadata if present
		// pgtype.Map stores the raw data - skip for now
		// Will need to enhance when we want to display metadata

		// Get inconsistencies for this event
		inconsistencies, err := h.db.ListInconsistenciesByEvent(r.Context(), event.ID)
		if err == nil {
			timelineEvent.Inconsistencies = make([]TimelineInconsistency, len(inconsistencies))
			for i, inc := range inconsistencies {
				timelineEvent.Inconsistencies[i] = TimelineInconsistency{
					ID:               inc.ID.String(),
					InconsistencyType: inc.InconsistencyType,
					Severity:         inc.Severity,
					Title:            inc.Title,
				}
			}
		}

		timelineEvents = append(timelineEvents, timelineEvent)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timelineEvents)
}

// dateString converts pgtype.Date to ISO string pointer
func dateString(date pgtype.Date) *string {
	if !date.Valid {
		return nil
	}
	s := date.Time.Format("2006-01-02")
	return &s
}
