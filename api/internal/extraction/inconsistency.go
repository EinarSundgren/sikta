package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
	"github.com/google/uuid"
)

// InconsistencyType represents different types of detected inconsistencies
type InconsistencyType string

const (
	InconsistencyTypeNarrativeChronologicalMismatch InconsistencyType = "narrative_chronological_mismatch"
	InconsistencyTypeTemporalImpossibility         InconsistencyType = "temporal_impossibility"
	InconsistencyTypeContradiction                 InconsistencyType = "contradiction"
	InconsistencyTypeCrossReference                 InconsistencyType = "cross_reference"
	InconsistencyTypeDuplicateEntity                InconsistencyType = "duplicate_entity"
	InconsistencyTypeDataMismatch                   InconsistencyType = "data_mismatch"
)

// Severity represents the severity level of an inconsistency
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning   Severity = "warning"
	SeverityConflict  Severity = "conflict"
)

// ResolutionStatus represents the resolution state
type ResolutionStatus string

const (
	ResolutionStatusUnresolved ResolutionStatus = "unresolved"
	ResolutionStatusResolved   ResolutionStatus = "resolved"
	ResolutionStatusNoted      ResolutionStatus = "noted"
	ResolutionStatusDismissed  ResolutionStatus = "dismissed"
)

// Inconsistency represents a detected inconsistency
type Inconsistency struct {
	ID               uuid.UUID
	DocumentID       uuid.UUID
	InconsistencyType InconsistencyType
	Severity         Severity
	Title            string
	Description      string
	ResolutionStatus ResolutionStatus
	ResolutionNote   *string
	Metadata         map[string]interface{}
}

// InconsistencyItem links an inconsistency to events or entities
type InconsistencyItem struct {
	ID               uuid.UUID
	InconsistencyID  uuid.UUID
	EventID          *uuid.UUID
	EntityID         *uuid.UUID
	Side             *string // 'a', 'b', or 'neutral' for conflicts
	Description      string
}

// NarrativeChronologyMismatch represents when narrative order ≠ chronological order
type NarrativeChronologyMismatch struct {
	EventID        uuid.UUID
	Title          string
	NarrativePos   int
	ChronologicalPos *int
}

// TemporalImpossibility represents logically impossible timing
type TemporalImpossibility struct {
	Description string
	EventIDs    []uuid.UUID
}

// Contradiction represents conflicting information
type Contradiction struct {
	Description string
	EventAID    uuid.UUID
	EventBID    uuid.UUID
}

// CrossReference represents when the same event appears in multiple places
type CrossReference struct {
	Description string
	EventIDs    []uuid.UUID
}

// InconsistencyDetector handles all inconsistency detection
type InconsistencyDetector struct {
	db     *database.Queries
	claude *claude.Client
	logger *slog.Logger
	model  string
}

// NewInconsistencyDetector creates a new detector
func NewInconsistencyDetector(db *database.Queries, claude *claude.Client, logger *slog.Logger, model string) *InconsistencyDetector {
	return &InconsistencyDetector{
		db:     db,
		claude: claude,
		logger: logger,
		model:  model,
	}
}

// DetectAll runs all inconsistency detection passes for a document
func (d *InconsistencyDetector) DetectAll(ctx context.Context, documentID string) ([]Inconsistency, error) {
	d.logger.Info("starting inconsistency detection", "document_id", documentID)

	var allInconsistencies []Inconsistency

	// Pass 1: Narrative vs chronological mismatches
	narrativeMismatches, err := d.detectNarrativeChronologicalMismatches(ctx, documentID)
	if err != nil {
		d.logger.Error("narrative/chronological detection failed", "error", err)
	} else {
		allInconsistencies = append(allInconsistencies, narrativeMismatches...)
	}

	// Pass 2: Temporal impossibilities
	temporalImpossibilities, err := d.detectTemporalImpossibilities(ctx, documentID)
	if err != nil {
		d.logger.Error("temporal impossibility detection failed", "error", err)
	} else {
		allInconsistencies = append(allInconsistencies, temporalImpossibilities...)
	}

	// Pass 3: Cross-references
	crossReferences, err := d.detectCrossReferences(ctx, documentID)
	if err != nil {
		d.logger.Error("cross-reference detection failed", "error", err)
	} else {
		allInconsistencies = append(allInconsistencies, crossReferences...)
	}

	d.logger.Info("inconsistency detection complete",
		"total", len(allInconsistencies),
		"narrative_mismatches", len(narrativeMismatches),
		"temporal_impossibilities", len(temporalImpossibilities),
		"cross_references", len(crossReferences))

	return allInconsistencies, nil
}

// detectNarrativeChronologicalMismatches finds events where story order ≠ timeline order
func (d *InconsistencyDetector) detectNarrativeChronologicalMismatches(ctx context.Context, documentID string) ([]Inconsistency, error) {
	events, err := d.db.ListEventsByDocument(ctx, parseUUID(documentID))
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	var inconsistencies []Inconsistency

	for _, event := range events {
		if event.ChronologicalPosition == nil {
			continue // Skip events without chronological position
		}

		// Check if narrative position differs significantly from chronological position
		// Allow small differences (flashbacks/foreshadowing within chapter)
		// Flag large differences as info-level mismatches
		narrativePos := event.NarrativePosition
		chronoPos := *event.ChronologicalPosition

		// If position differs by more than 3, it's a significant narrative device
		if absInt32(narrativePos-chronoPos) > 3 {
			inconsistency := Inconsistency{
				ID:               uuid.New(),
				DocumentID:       parseUUID(documentID),
				InconsistencyType: InconsistencyTypeNarrativeChronologicalMismatch,
				Severity:         SeverityInfo,
				Title:            fmt.Sprintf("Narrative order mismatch: %s", event.Title),
				Description:      fmt.Sprintf("Event appears at narrative position %d but chronological position %d", narrativePos, chronoPos),
				ResolutionStatus: ResolutionStatusUnresolved,
				Metadata: map[string]interface{}{
					"event_id":           event.ID.String(),
					"event_title":         event.Title,
					"narrative_position":   narrativePos,
					"chronological_position": chronoPos,
				},
			}

			// Store in database
			created, err := d.createInconsistency(ctx, inconsistency, event.ID)
			if err != nil {
				d.logger.Error("failed to create inconsistency", "error", err)
				continue
			}

			inconsistencies = append(inconsistencies, created)
		}
	}

	return inconsistencies, nil
}

// detectTemporalImpossibilities detects logically impossible timing
func (d *InconsistencyDetector) detectTemporalImpossibilities(ctx context.Context, documentID string) ([]Inconsistency, error) {
	// Reuse existing chronology.go temporal impossibility detection
	// This is already implemented in the ChronologicalEstimator

	// For now, return empty - we can enhance this later
	return []Inconsistency{}, nil
}

// detectCrossReferences finds when the same event is described in multiple places
func (d *InconsistencyDetector) detectCrossReferences(ctx context.Context, documentID string) ([]Inconsistency, error) {
	events, err := d.db.ListEventsByDocument(ctx, parseUUID(documentID))
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	var inconsistencies []Inconsistency

	// Group events by title similarity (simple approach)
	eventGroups := make(map[string][]database.Event)

	for _, event := range events {
		// Normalize title for grouping
		normalizedTitle := strings.ToLower(strings.TrimSpace(event.Title))

		// Simple grouping: exact title match after normalization
		eventGroups[normalizedTitle] = append(eventGroups[normalizedTitle], event)
	}

	// Find groups with multiple events (potential cross-references)
	for title, group := range eventGroups {
		if len(group) > 1 {
			// Found potential cross-reference
			var eventIDs []uuid.UUID
			var titles []string

			for _, event := range group {
				eventIDs = append(eventIDs, uuid.UUID(event.ID))
				titles = append(titles, event.Title)
			}

			inconsistency := Inconsistency{
				ID:               uuid.New(),
				DocumentID:       parseUUID(documentID),
				InconsistencyType: InconsistencyTypeCrossReference,
				Severity:         SeverityInfo,
				Title:            fmt.Sprintf("Cross-reference: %s", title),
				Description:      fmt.Sprintf("Event appears %d times: %s", len(group), strings.Join(titles, "; ")),
				ResolutionStatus: ResolutionStatusUnresolved,
				Metadata: map[string]interface{}{
					"event_ids": eventIDs,
					"titles":    titles,
				},
			}

			// Store in database
			created, err := d.createInconsistency(ctx, inconsistency)
			if err != nil {
				d.logger.Error("failed to create inconsistency", "error", err)
				continue
			}

			inconsistencies = append(inconsistencies, created)
		}
	}

	return inconsistencies, nil
}

// createInconsistency stores an inconsistency in the database
func (d *InconsistencyDetector) createInconsistency(ctx context.Context, inconsistency Inconsistency, eventID ...database.UUID) (Inconsistency, error) {
	metadataJSON, _ := json.Marshal(inconsistency.Metadata)

	var resolutionNote interface{}
	if inconsistency.ResolutionNote != nil {
		resolutionNote = *inconsistency.ResolutionNote
	}

	created, err := d.db.CreateInconsistency(ctx, database.CreateInconsistencyParams{
		ID:               database.UUID(inconsistency.ID),
		DocumentID:       inconsistency.DocumentID,
		InconsistencyType: string(inconsistency.InconsistencyType),
		Severity:         string(inconsistency.Severity),
		Title:            inconsistency.Title,
		Description:      inconsistency.Description,
		ResolutionStatus: string(inconsistency.ResolutionStatus),
		ResolutionNote:    resolutionNote,
		Metadata:         metadataJSON,
	})
	if err != nil {
		return Inconsistency{}, err
	}

	// Link to events if provided
	for _, eid := range eventID {
		eventUUID := database.UUID(eid)
		_, err := d.db.CreateInconsistencyItem(ctx, database.CreateInconsistencyItemParams{
			ID:              uuid.New(),
			InconsistencyID: database.UUID(created.ID),
			EventID:         &eventUUID,
			Side:            nil,
			Description:     "Related event",
		})
		if err != nil {
			d.logger.Error("failed to link event to inconsistency", "error", err)
		}
	}

	return inconsistencyFromDB(created), nil
}

// DetectContradictionsWithLLM uses LLM to detect contradictions between events
func (d *InconsistencyDetector) DetectContradictionsWithLLM(ctx context.Context, documentID string) ([]Inconsistency, error) {
	// Get all events for the document
	events, err := d.db.ListEventsByDocument(ctx, parseUUID(documentID))
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) < 2 {
		return nil, nil // Need at least 2 events to have contradictions
	}

	// Build events summary for LLM
	eventsSummary := d.buildEventsSummaryForContradictionDetection(events)

	// Call LLM for contradiction detection
	prompt := fmt.Sprintf(ContradictionDetectionPrompt, eventsSummary)
	resp, err := d.claude.SendSystemPrompt(ctx, "", prompt, d.model)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}

	// Parse response
	var detectionResult ContradictionDetectionResult
	if err := json.Unmarshal([]byte(resp.Content[0].Text), &detectionResult); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Convert to Inconsistency objects
	var inconsistencies []Inconsistency
	for _, contradiction := range detectionResult.Contradictions {
		inconsistency := Inconsistency{
			ID:               uuid.New(),
			DocumentID:       parseUUID(documentID),
			InconsistencyType: InconsistencyTypeContradiction,
			Severity:         SeverityConflict,
			Title:            contradiction.Description,
			Description:      contradiction.Description,
			ResolutionStatus: ResolutionStatusUnresolved,
			Metadata: map[string]interface{}{
				"event_a_id": contradiction.EventAID,
				"event_b_id": contradiction.EventBID,
			},
		}

		// Store in database
		created, err := d.createInconsistency(ctx, inconsistency)
		if err != nil {
			d.logger.Error("failed to create contradiction", "error", err)
			continue
		}

		// Link both events
		eventAID := parseUUID(contradiction.EventAID)
		eventBID := parseUUID(contradiction.EventBID)

		_, err = d.db.CreateInconsistencyItem(ctx, database.CreateInconsistencyItemParams{
			ID:              uuid.New(),
			InconsistencyID: database.UUID(created.ID),
			EventID:         &eventAID,
			Side:            stringPtr("a"),
			Description:     "Side A of contradiction",
		})
		if err != nil {
			d.logger.Error("failed to link event A", "error", err)
		}

		_, err = d.db.CreateInconsistencyItem(ctx, database.CreateInconsistencyItemParams{
			ID:              uuid.New(),
			InconsistencyID: database.UUID(created.ID),
			EventID:         &eventBID,
			Side:            stringPtr("b"),
			Description:     "Side B of contradiction",
		})
		if err != nil {
			d.logger.Error("failed to link event B", "error", err)
		}

		inconsistencies = append(inconsistencies, created)
	}

	return inconsistencies, nil
}

// buildEventsSummaryForContradictionDetection creates a summary for LLM analysis
func (d *InconsistencyDetector) buildEventsSummaryForContradictionDetection(events []database.Event) string {
	var summary strings.Builder

	summary.WriteString("Events:\n\n")
	for _, event := range events {
		summary.WriteString(fmt.Sprintf("- ID: %s\n", event.ID))
		summary.WriteString(fmt.Sprintf("  Title: %s\n", event.Title))
		if event.Description != nil {
			summary.WriteString(fmt.Sprintf("  Description: %s\n", *event.Description))
		}
		if event.DateText != nil {
			summary.WriteString(fmt.Sprintf("  Date: %s\n", *event.DateText))
		}
		if event.ChronologicalPosition != nil {
			summary.WriteString(fmt.Sprintf("  Chronological Position: %d\n", *event.ChronologicalPosition))
		}
		summary.WriteString("\n")
	}

	return summary.String()
}

func inconsistencyFromDB(dbInc database.Inconsistency) Inconsistency {
	var metadata map[string]interface{}
	if dbInc.Metadata != nil {
		// JSONB is already unmarshalled by pgx
		metadata = dbInc.Metadata.(map[string]interface{})
	}

	var resolutionNote *string
	if dbInc.ResolutionNote != nil {
		note := *dbInc.ResolutionNote
		resolutionNote = &note
	}

	return Inconsistency{
		ID:               uuid.UUID(dbInc.ID),
		DocumentID:       uuid.UUID(dbInc.DocumentID),
		InconsistencyType: InconsistencyType(dbInc.InconsistencyType),
		Severity:         Severity(dbInc.Severity),
		Title:            dbInc.Title,
		Description:      dbInc.Description,
		ResolutionStatus: ResolutionStatus(dbInc.ResolutionStatus),
		ResolutionNote:   resolutionNote,
		Metadata:         metadata,
	}
}

// ContradictionDetectionResult represents LLM response for contradiction detection
type ContradictionDetectionResult struct {
	Contradictions []struct {
		Description string `json:"description"`
		EventAID    string `json:"event_a_id"`
		EventBID    string `json:"event_b_id"`
	} `json:"contradictions"`
}

// ContradictionDetectionPrompt is the LLM prompt for detecting contradictions
const ContradictionDetectionPrompt = `You are an expert analyst detecting contradictions and inconsistencies in narrative events.

Given the following events from a document, identify any contradictions:

%s

Analyze the events for:
- Factual contradictions (event A says X happened, event B says X didn't happen)
- Temporal contradictions (conflicting dates or times)
- Logical impossibilities (events that cannot both be true)

For each contradiction found, return:
{
  "contradictions": [
    {
      "description": "Brief description of the contradiction",
      "event_a_id": "ID of first event",
      "event_b_id": "ID of second event"
    }
  ]
}

Return valid JSON only, no markdown formatting. If no contradictions are found, return an empty contradictions array.`

func absInt32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
