package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

// TimelineEvent represents a claim on the timeline with all related data.
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

// TimelineEntity represents an entity connected to a claim.
type TimelineEntity struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	EntityType string  `json:"entity_type"`
	Role       *string `json:"role"`
}

// TimelineSource represents a source reference for a claim.
type TimelineSource struct {
	ID            string  `json:"id"`
	ChunkID       string  `json:"chunk_id"`
	Excerpt       string  `json:"excerpt"`
	ChapterTitle  *string `json:"chapter_title"`
	ChapterNumber *int32  `json:"chapter_number"`
	PageStart     *int32  `json:"page_start"`
	PageEnd       *int32  `json:"page_end"`
}

// TimelineInconsistency represents an inconsistency linked to a claim.
type TimelineInconsistency struct {
	ID                string `json:"id"`
	InconsistencyType string `json:"inconsistency_type"`
	Severity          string `json:"severity"`
	Title             string `json:"title"`
}

// GetTimeline handles GET /api/documents/:id/timeline
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

	sourceID := database.PgUUID(parsedUUID)

	claims, err := h.db.ListClaimsBySource(r.Context(), sourceID)
	if err != nil {
		h.logger.Error("failed to get claims", "error", err)
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	// Load all claim-entity associations in one query, group by claim ID.
	claimEntityRows, err := h.db.ListClaimEntitiesBySource(r.Context(), sourceID)
	if err != nil {
		h.logger.Error("failed to get claim entities", "error", err)
		// Non-fatal: continue without entity data
		claimEntityRows = nil
	}
	entitiesByClaim := make(map[string][]TimelineEntity)
	for _, row := range claimEntityRows {
		claimID := database.UUIDStr(row.ClaimID)
		entitiesByClaim[claimID] = append(entitiesByClaim[claimID], TimelineEntity{
			ID:         database.UUIDStr(row.EntityID),
			Name:       row.Name,
			EntityType: row.EntityType,
			Role:       database.TextPtr(row.Role),
		})
	}

	timelineEvents := make([]TimelineEvent, 0, len(claims))
	for _, claim := range claims {
		claimIDStr := database.UUIDStr(claim.ID)
		entities := entitiesByClaim[claimIDStr]
		if entities == nil {
			entities = []TimelineEntity{}
		}

		timelineEvent := TimelineEvent{
			ID:                claimIDStr,
			SourceID:          database.UUIDStr(claim.SourceID),
			Title:             claim.Title,
			Description:       database.TextPtr(claim.Description),
			EventType:         database.TextPtr(claim.EventType),
			DateText:          database.TextPtr(claim.DateText),
			DateStart:         dateString(claim.DateStart),
			DateEnd:           dateString(claim.DateEnd),
			DatePrecision:     database.TextPtr(claim.DatePrecision),
			NarrativePosition: claim.NarrativePosition,
			ChronologicalPos:  database.Int4Ptr(claim.ChronologicalPosition),
			Confidence:        float64(claim.Confidence),
			ConfidenceReason:  database.TextPtr(claim.ConfidenceReason),
			ReviewStatus:      claim.ReviewStatus,
			Entities:          entities,
			SourceReferences:  []TimelineSource{},
		}

		inconsistencies, err := h.db.ListInconsistenciesByClaim(r.Context(), claim.ID)
		if err == nil {
			timelineEvent.Inconsistencies = make([]TimelineInconsistency, len(inconsistencies))
			for i, inc := range inconsistencies {
				timelineEvent.Inconsistencies[i] = TimelineInconsistency{
					ID:                database.UUIDStr(inc.ID),
					InconsistencyType: inc.InconsistencyType,
					Severity:          inc.Severity,
					Title:             inc.Title,
				}
			}
		}

		timelineEvents = append(timelineEvents, timelineEvent)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timelineEvents)
}

// dateString converts pgtype.Date to ISO string pointer.
func dateString(date pgtype.Date) *string {
	if !date.Valid {
		return nil
	}
	s := date.Time.Format("2006-01-02")
	return &s
}
