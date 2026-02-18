package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
)

// ChronologicalEstimator handles timeline ordering of events.
type ChronologicalEstimator struct {
	db     *database.Queries
	claude *claude.Client
	logger *slog.Logger
	model  string
}

// NewChronologicalEstimator creates a new chronological estimator.
func NewChronologicalEstimator(db *database.Queries, claude *claude.Client, logger *slog.Logger, model string) *ChronologicalEstimator {
	return &ChronologicalEstimator{
		db:     db,
		claude: claude,
		logger: logger,
		model:  model,
	}
}

// EstimationResult contains the results of chronological estimation.
type EstimationResult struct {
	OrderedCount    int
	AnomaliesDetected int
	Anomalies        []Anomaly
}

// EstimateChronology estimates the chronological order of events.
func (e *ChronologicalEstimator) EstimateChronology(ctx context.Context, documentID string) (*EstimationResult, error) {
	e.logger.Info("estimating chronology", "document_id", documentID)

	// Get all events for the document
	events, err := e.db.ListEventsByDocument(ctx, database.UUID(parseUUID(documentID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return &EstimationResult{}, nil
	}

	// Build events summary for LLM
	eventsSummary := e.buildEventsSummary(events)

	// Call LLM for ordering
	prompt := fmt.Sprintf(ChronologyEstimationPrompt, eventsSummary)
	resp, err := e.claude.SendSystemPrompt(ctx, "", prompt, e.model)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}

	// Parse response
	var chronology ChronologicalOrder
	if err := json.Unmarshal([]byte(resp.Content[0].Text), &chronology); err != nil {
		return nil, fmt.Errorf("failed to parse chronology response: %w", err)
	}

	// Update chronological positions in database
	orderedCount := 0
	for _, pos := range chronology.ChronologicalOrder {
		eventID := parseUUID(pos.EventID)
		_, err := e.db.UpdateEventChronologicalPosition(ctx, database.UpdateEventChronologicalPositionParams{
			ID:                    eventID,
			ChronologicalPosition: int32(pos.ChronologicalPosition),
		})
		if err != nil {
			e.logger.Error("failed to update event position", "event_id", pos.EventID, "error", err)
			continue
		}
		orderedCount++
	}

	// Detect temporal impossibilities
	anomalies := e.detectTemporalImpossibilities(events)

	e.logger.Info("chronology estimation complete",
		"ordered", orderedCount,
		"anomalies", len(anomalies))

	return &EstimationResult{
		OrderedCount:      orderedCount,
		AnomaliesDetected: len(anomalies),
		Anomalies:         anomalies,
	}, nil
}

// buildEventsSummary creates a summary of events for the LLM.
func (e *ChronologicalEstimator) buildEventsSummary(events []database.Event) string {
	var summary strings.Builder

	summary.WriteString("Events:\n")
	for _, event := range events {
		summary.WriteString(fmt.Sprintf("- ID: %s\n", event.ID))
		summary.WriteString(fmt.Sprintf("  Title: %s\n", event.Title))
		summary.WriteString(fmt.Sprintf("  Description: %s\n", safeString(event.Description)))
		summary.WriteString(fmt.Sprintf("  Date: %s\n", safeString(event.DateText)))
		summary.WriteString(fmt.Sprintf("  Narrative Position: %d\n", event.NarrativePosition))
		summary.WriteString("\n")
	}

	return summary.String()
}

// detectTemporalImpossibilities detects temporal impossibilities in events.
func (e *ChronologicalEstimator) detectTemporalImpossibilities(events []database.Event) []Anomaly {
	var anomalies []Anomaly

	// Build character timelines
	characterTimelines := e.buildCharacterTimelines(events)

	// Check for overlaps
	for character, timeline := range characterTimelines {
		overlaps := e.findOverlappingEvents(timeline)
		if len(overlaps) > 0 {
			if !e.verifyPossible(overlaps) {
				anomalies = append(anomalies, Anomaly{
					Type:        "temporal_impossibility",
					Description: fmt.Sprintf("%s cannot be in two places at once", character),
					Events: overlaps,
				})
			}
		}
	}

	return anomalies
}

// characterTimeline tracks events for a character.
type characterTimeline struct {
	Character string
	Events    []database.Event
}

// buildCharacterTimelines builds timelines for each character.
func (e *ChronologicalEstimator) buildCharacterTimelines(events []database.Event) map[string][]database.Event {
	timelines := make(map[string][]database.Event)

	for _, event := range events {
		// Get entities involved in this event
		// For now, we'll use a simplified approach
		// In production, you'd query event_entities table

		// Extract character names from title/description
		characters := e.extractCharacters(event)
		for _, character := range characters {
			timelines[character] = append(timelines[character], event)
		}
	}

	return timelines
}

// extractCharacters extracts character names from an event.
func (e *ChronologicalEstimator) extractCharacters(event database.Event) []string {
	// Simplified - in production, this would query the event_entities table
	// For now, extract from title and description
	var characters []string

	// Common name patterns (Mr., Mrs., Miss, etc.)
	// This is a very simplified implementation
	return characters
}

// findOverlappingEvents finds events that overlap in time.
func (e *ChronologicalEstimator) findOverlappingEvents(events []database.Event) []string {
	if len(events) < 2 {
		return nil
	}

	// Sort by chronological position
	sort.Slice(events, func(i, j int) bool {
		posI := events[i].ChronologicalPosition
		posJ := events[j].ChronologicalPosition
		if posI == nil && posJ == nil {
			return events[i].NarrativePosition < events[j].NarrativePosition
		}
		if posI == nil {
			return false
		}
		if posJ == nil {
			return true
		}
		return *posI < *posJ
	})

	// Check for impossible overlaps
	var overlapping []string
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if e.eventsConflict(events[i], events[j]) {
				overlapping = append(overlapping, events[i].ID.String(), events[j].ID.String())
			}
		}
	}

	return overlapping
}

// eventsConflict checks if two events temporally conflict.
func (e *ChronologicalEstimator) eventsConflict(eventA, eventB database.Event) bool {
	// If no date information, can't determine conflict
	if !eventA.DateStart.Valid && !eventB.DateStart.Valid {
		return false
	}

	// If both have date information, check for overlap
	if eventA.DateStart.Valid && eventB.DateStart.Valid {
		// Same date and different locations could be a conflict
		// This is simplified - real implementation would need location info
		return false
	}

	// Check if descriptions suggest same time, different places
	descA := safeString(eventA.Description)
	descB := safeString(eventB.Description)

	// Look for location indicators in both
	if e.hasLocation(descA) && e.hasLocation(descB) {
		locationA := e.extractLocation(descA)
		locationB := e.extractLocation(descB)
		if locationA != "" && locationB != "" && locationA != locationB {
			// Different locations - check if same time
			return e.sameTime(eventA, eventB)
		}
	}

	return false
}

// verifyPossible checks if overlapping events are legitimately possible.
func (e *ChronologicalEstimator) verifyPossible(eventIDs []string) bool {
	// For now, assume overlaps are possible
	// In production, this would check travel times, etc.
	return true
}

// hasLocation checks if text contains location information.
func (e *ChronologicalEstimator) hasLocation(text string) bool {
	locationIndicators := []string{
		"at ", "in ", "to ", "from ", "near ",
		"Netherfield", "Longbourn", "London", "Hertfordshire",
	}
	for _, indicator := range locationIndicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}

// extractLocation extracts location from text.
func (e *ChronologicalEstimator) extractLocation(text string) string {
	// Simplified location extraction
	// In production, this would use NLP
	if strings.Contains(text, "Netherfield") {
		return "Netherfield"
	}
	if strings.Contains(text, "Longbourn") {
		return "Longbourn"
	}
	if strings.Contains(text, "London") {
		return "London"
	}
	return ""
}

// sameTime checks if two events occur at the same time.
func (e *ChronologicalEstimator) sameTime(eventA, eventB database.Event) bool {
	// If both have date information, compare
	if eventA.DateStart.Valid && eventB.DateStart.Valid {
		return eventA.DateStart.Time.Equal(eventB.DateStart.Time)
	}

	// If both have date text, compare
	dateA := safeString(eventA.DateText)
	dateB := safeString(eventB.DateText)

	if dateA != "" && dateB != "" {
		// Check for same-day indicators
		if strings.Contains(dateA, "that") && strings.Contains(dateB, "that") {
			return true
		}
		if strings.Contains(dateA, "same day") && strings.Contains(dateB, "same day") {
			return true
		}
	}

	return false
}

// safeString safely converts a string pointer to string.
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
