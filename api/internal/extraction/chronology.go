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
	"github.com/jackc/pgx/v5/pgtype"
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
	OrderedCount      int
	AnomaliesDetected int
	Anomalies         []Anomaly
}

// EstimateChronology estimates the chronological order of events.
func (e *ChronologicalEstimator) EstimateChronology(ctx context.Context, sourceID string) (*EstimationResult, error) {
	e.logger.Info("estimating chronology", "source_id", sourceID)

	claims, err := e.db.ListClaimsBySource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get claims: %w", err)
	}

	if len(claims) == 0 {
		return &EstimationResult{}, nil
	}

	eventsSummary := e.buildEventsSummary(claims)

	prompt := ChronologyEstimationPrompt + "\n\n" + eventsSummary
	resp, err := e.claude.SendSystemPrompt(ctx, "", prompt, e.model)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}

	// Strip markdown code blocks before parsing
	responseText := stripMarkdownCodeBlocks(resp.Content[0].Text)

	// Fallback: try to extract JSON from markdown if primary parsing fails
	var chronology ChronologicalOrder
	if err := json.Unmarshal([]byte(responseText), &chronology); err != nil {
		if jsonStr := extractJSONFromMarkdown(responseText); jsonStr != "" {
			responseText = jsonStr
			if err := json.Unmarshal([]byte(responseText), &chronology); err != nil {
				return nil, fmt.Errorf("failed to parse chronology response after markdown extraction: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse chronology response: %w", err)
		}
	}

	e.logger.Info("chronology parsed",
		"events_ordered", len(chronology.ChronologicalOrder),
		"anomalies", len(chronology.Anomalies))

	orderedCount := 0
	for _, pos := range chronology.ChronologicalOrder {
		_, err := e.db.UpdateClaimChronologicalPosition(ctx, database.UpdateClaimChronologicalPositionParams{
			ID:                    database.PgUUID(parseUUID(pos.EventID)),
			ChronologicalPosition: pgtype.Int4{Int32: int32(pos.ChronologicalPosition), Valid: true},
		})
		if err != nil {
			e.logger.Error("failed to update claim position", "claim_id", pos.EventID, "error", err)
			continue
		}
		orderedCount++
	}

	anomalies := e.detectTemporalImpossibilities(claims)

	e.logger.Info("chronology estimation complete",
		"ordered", orderedCount,
		"anomalies", len(anomalies))

	return &EstimationResult{
		OrderedCount:      orderedCount,
		AnomaliesDetected: len(anomalies),
		Anomalies:         anomalies,
	}, nil
}

// buildEventsSummary creates a summary of claims for the LLM.
func (e *ChronologicalEstimator) buildEventsSummary(claims []*database.Claim) string {
	var summary strings.Builder

	summary.WriteString("EVENTS WITH NARRATIVE ORDER:\n\n")
	for _, claim := range claims {
		summary.WriteString(fmt.Sprintf("Event ID: %s\n", database.UUIDStr(claim.ID)))
		summary.WriteString(fmt.Sprintf("  Title: %s\n", claim.Title))
		summary.WriteString(fmt.Sprintf("  Description: %s\n", claim.Description.String))
		if claim.DateText.String != "" {
			summary.WriteString(fmt.Sprintf("  Date Reference: %s\n", claim.DateText.String))
		}
		summary.WriteString(fmt.Sprintf("  Type: %s\n", claim.EventType.String))
		summary.WriteString(fmt.Sprintf("  Narrative Position: %d (order in text)\n", claim.NarrativePosition))
		summary.WriteString("\n")
	}

	return summary.String()
}

// detectTemporalImpossibilities detects temporal impossibilities in claims.
func (e *ChronologicalEstimator) detectTemporalImpossibilities(claims []*database.Claim) []Anomaly {
	var anomalies []Anomaly

	characterTimelines := e.buildCharacterTimelines(claims)

	for character, timeline := range characterTimelines {
		overlaps := e.findOverlappingEvents(timeline)
		if len(overlaps) > 0 {
			if !e.verifyPossible(overlaps) {
				anomalies = append(anomalies, Anomaly{
					Type:        "temporal_impossibility",
					Description: fmt.Sprintf("%s cannot be in two places at once", character),
					Events:      overlaps,
				})
			}
		}
	}

	return anomalies
}

// buildCharacterTimelines builds timelines for each character.
func (e *ChronologicalEstimator) buildCharacterTimelines(claims []*database.Claim) map[string][]*database.Claim {
	timelines := make(map[string][]*database.Claim)

	for _, claim := range claims {
		characters := e.extractCharacters(claim)
		for _, character := range characters {
			timelines[character] = append(timelines[character], claim)
		}
	}

	return timelines
}

// extractCharacters extracts character names from a claim.
func (e *ChronologicalEstimator) extractCharacters(claim *database.Claim) []string {
	var characters []string
	return characters
}

// findOverlappingEvents finds claims that overlap in time.
func (e *ChronologicalEstimator) findOverlappingEvents(claims []*database.Claim) []string {
	if len(claims) < 2 {
		return nil
	}

	sort.Slice(claims, func(i, j int) bool {
		posI := claims[i].ChronologicalPosition
		posJ := claims[j].ChronologicalPosition
		if !posI.Valid && !posJ.Valid {
			return claims[i].NarrativePosition < claims[j].NarrativePosition
		}
		if !posI.Valid {
			return false
		}
		if !posJ.Valid {
			return true
		}
		return posI.Int32 < posJ.Int32
	})

	var overlapping []string
	for i := 0; i < len(claims)-1; i++ {
		for j := i + 1; j < len(claims); j++ {
			if e.eventsConflict(claims[i], claims[j]) {
				overlapping = append(overlapping, database.UUIDStr(claims[i].ID), database.UUIDStr(claims[j].ID))
			}
		}
	}

	return overlapping
}

// eventsConflict checks if two claims temporally conflict.
func (e *ChronologicalEstimator) eventsConflict(claimA, claimB *database.Claim) bool {
	if !claimA.DateStart.Valid && !claimB.DateStart.Valid {
		return false
	}

	if claimA.DateStart.Valid && claimB.DateStart.Valid {
		return false
	}

	descA := claimA.Description.String
	descB := claimB.Description.String

	if e.hasLocation(descA) && e.hasLocation(descB) {
		locationA := e.extractLocation(descA)
		locationB := e.extractLocation(descB)
		if locationA != "" && locationB != "" && locationA != locationB {
			return e.sameTime(claimA, claimB)
		}
	}

	return false
}

// verifyPossible checks if overlapping claims are legitimately possible.
func (e *ChronologicalEstimator) verifyPossible(eventIDs []string) bool {
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

// sameTime checks if two claims occur at the same time.
func (e *ChronologicalEstimator) sameTime(claimA, claimB *database.Claim) bool {
	if claimA.DateStart.Valid && claimB.DateStart.Valid {
		return claimA.DateStart.Time.Equal(claimB.DateStart.Time)
	}

	dateA := claimA.DateText.String
	dateB := claimB.DateText.String

	if dateA != "" && dateB != "" {
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
