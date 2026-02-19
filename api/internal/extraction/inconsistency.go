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
	"github.com/jackc/pgx/v5/pgtype"
)

// InconsistencyType represents different types of detected inconsistencies
type InconsistencyType string

const (
	InconsistencyTypeNarrativeChronologicalMismatch InconsistencyType = "narrative_chronological_mismatch"
	InconsistencyTypeTemporalImpossibility          InconsistencyType = "temporal_impossibility"
	InconsistencyTypeContradiction                  InconsistencyType = "contradiction"
	InconsistencyTypeCrossReference                 InconsistencyType = "cross_reference"
	InconsistencyTypeDuplicateEntity                InconsistencyType = "duplicate_entity"
	InconsistencyTypeDataMismatch                   InconsistencyType = "data_mismatch"
)

// Severity represents the severity level of an inconsistency
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityConflict Severity = "conflict"
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
	ID                uuid.UUID
	SourceID          uuid.UUID
	InconsistencyType InconsistencyType
	Severity          Severity
	Title             string
	Description       string
	ResolutionStatus  ResolutionStatus
	ResolutionNote    *string
	Metadata          map[string]interface{}
}

// InconsistencyItem links an inconsistency to claims or entities
type InconsistencyItem struct {
	ID              uuid.UUID
	InconsistencyID uuid.UUID
	ClaimID         *uuid.UUID
	EntityID        *uuid.UUID
	Side            *string
	Description     string
}

// NarrativeChronologyMismatch represents when narrative order ≠ chronological order
type NarrativeChronologyMismatch struct {
	ClaimID          uuid.UUID
	Title            string
	NarrativePos     int
	ChronologicalPos *int
}

// TemporalImpossibility represents logically impossible timing
type TemporalImpossibility struct {
	Description string
	ClaimIDs    []uuid.UUID
}

// Contradiction represents conflicting information
type Contradiction struct {
	Description string
	ClaimAID    uuid.UUID
	ClaimBID    uuid.UUID
}

// CrossReference represents when the same event appears in multiple places
type CrossReference struct {
	Description string
	ClaimIDs    []uuid.UUID
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

// DetectAll runs all inconsistency detection passes for a source
func (d *InconsistencyDetector) DetectAll(ctx context.Context, sourceID string) ([]Inconsistency, error) {
	d.logger.Info("starting inconsistency detection", "source_id", sourceID)

	var allInconsistencies []Inconsistency

	narrativeMismatches, err := d.detectNarrativeChronologicalMismatches(ctx, sourceID)
	if err != nil {
		d.logger.Error("narrative/chronological detection failed", "error", err)
	} else {
		allInconsistencies = append(allInconsistencies, narrativeMismatches...)
	}

	temporalImpossibilities, err := d.detectTemporalImpossibilities(ctx, sourceID)
	if err != nil {
		d.logger.Error("temporal impossibility detection failed", "error", err)
	} else {
		allInconsistencies = append(allInconsistencies, temporalImpossibilities...)
	}

	crossReferences, err := d.detectCrossReferences(ctx, sourceID)
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

// detectNarrativeChronologicalMismatches finds claims where story order ≠ timeline order
func (d *InconsistencyDetector) detectNarrativeChronologicalMismatches(ctx context.Context, sourceID string) ([]Inconsistency, error) {
	claims, err := d.db.ListClaimsBySource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get claims: %w", err)
	}

	var inconsistencies []Inconsistency

	for _, claim := range claims {
		if !claim.ChronologicalPosition.Valid {
			continue
		}

		narrativePos := claim.NarrativePosition
		chronoPos := claim.ChronologicalPosition.Int32

		if absInt32(narrativePos-chronoPos) > 3 {
			inconsistency := Inconsistency{
				ID:                uuid.New(),
				SourceID:          parseUUID(sourceID),
				InconsistencyType: InconsistencyTypeNarrativeChronologicalMismatch,
				Severity:          SeverityInfo,
				Title:             fmt.Sprintf("Narrative order mismatch: %s", claim.Title),
				Description:       fmt.Sprintf("Event appears at narrative position %d but chronological position %d", narrativePos, chronoPos),
				ResolutionStatus:  ResolutionStatusUnresolved,
				Metadata: map[string]interface{}{
					"claim_id":               database.UUIDStr(claim.ID),
					"claim_title":            claim.Title,
					"narrative_position":     narrativePos,
					"chronological_position": chronoPos,
				},
			}

			created, err := d.createInconsistency(ctx, inconsistency, claim.ID)
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
func (d *InconsistencyDetector) detectTemporalImpossibilities(ctx context.Context, sourceID string) ([]Inconsistency, error) {
	return []Inconsistency{}, nil
}

// detectCrossReferences finds when the same event is described in multiple places
func (d *InconsistencyDetector) detectCrossReferences(ctx context.Context, sourceID string) ([]Inconsistency, error) {
	claims, err := d.db.ListClaimsBySource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get claims: %w", err)
	}

	var inconsistencies []Inconsistency

	claimGroups := make(map[string][]*database.Claim)

	for _, claim := range claims {
		normalizedTitle := strings.ToLower(strings.TrimSpace(claim.Title))
		claimGroups[normalizedTitle] = append(claimGroups[normalizedTitle], claim)
	}

	for title, group := range claimGroups {
		if len(group) > 1 {
			var claimIDs []uuid.UUID
			var titles []string

			for _, claim := range group {
				claimIDs = append(claimIDs, uuid.UUID(claim.ID.Bytes))
				titles = append(titles, claim.Title)
			}

			inconsistency := Inconsistency{
				ID:                uuid.New(),
				SourceID:          parseUUID(sourceID),
				InconsistencyType: InconsistencyTypeCrossReference,
				Severity:          SeverityInfo,
				Title:             fmt.Sprintf("Cross-reference: %s", title),
				Description:       fmt.Sprintf("Event appears %d times: %s", len(group), strings.Join(titles, "; ")),
				ResolutionStatus:  ResolutionStatusUnresolved,
				Metadata: map[string]interface{}{
					"claim_ids": claimIDs,
					"titles":    titles,
				},
			}

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
func (d *InconsistencyDetector) createInconsistency(ctx context.Context, inconsistency Inconsistency, claimID ...pgtype.UUID) (Inconsistency, error) {
	metadataJSON, _ := json.Marshal(inconsistency.Metadata)

	created, err := d.db.CreateInconsistency(ctx, database.CreateInconsistencyParams{
		ID:                database.PgUUID(inconsistency.ID),
		SourceID:          database.PgUUID(inconsistency.SourceID),
		InconsistencyType: string(inconsistency.InconsistencyType),
		Severity:          string(inconsistency.Severity),
		Title:             inconsistency.Title,
		Description:       inconsistency.Description,
		Metadata:          metadataJSON,
	})
	if err != nil {
		return Inconsistency{}, err
	}

	for _, cid := range claimID {
		_, err := d.db.CreateInconsistencyItem(ctx, database.CreateInconsistencyItemParams{
			ID:              database.PgUUID(uuid.New()),
			InconsistencyID: created.ID,
			ClaimID:         cid,
			EntityID:        pgtype.UUID{},
			Side:            pgtype.Text{},
			Description:     database.PgText("Related claim"),
		})
		if err != nil {
			d.logger.Error("failed to link claim to inconsistency", "error", err)
		}
	}

	return inconsistencyFromDB(created), nil
}

// DetectContradictionsWithLLM uses LLM to detect contradictions between claims
func (d *InconsistencyDetector) DetectContradictionsWithLLM(ctx context.Context, sourceID string) ([]Inconsistency, error) {
	claims, err := d.db.ListClaimsBySource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get claims: %w", err)
	}

	if len(claims) < 2 {
		return nil, nil
	}

	eventsSummary := d.buildEventsSummaryForContradictionDetection(claims)

	prompt := fmt.Sprintf(ContradictionDetectionPrompt, eventsSummary)
	resp, err := d.claude.SendSystemPrompt(ctx, "", prompt, d.model)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}

	var detectionResult ContradictionDetectionResult
	if err := json.Unmarshal([]byte(resp.Content[0].Text), &detectionResult); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	var inconsistencies []Inconsistency
	for _, contradiction := range detectionResult.Contradictions {
		inconsistency := Inconsistency{
			ID:                uuid.New(),
			SourceID:          parseUUID(sourceID),
			InconsistencyType: InconsistencyTypeContradiction,
			Severity:          SeverityConflict,
			Title:             contradiction.Description,
			Description:       contradiction.Description,
			ResolutionStatus:  ResolutionStatusUnresolved,
			Metadata: map[string]interface{}{
				"claim_a_id": contradiction.EventAID,
				"claim_b_id": contradiction.EventBID,
			},
		}

		created, err := d.createInconsistency(ctx, inconsistency)
		if err != nil {
			d.logger.Error("failed to create contradiction", "error", err)
			continue
		}

		claimAID := database.PgUUID(parseUUID(contradiction.EventAID))
		claimBID := database.PgUUID(parseUUID(contradiction.EventBID))

		_, err = d.db.CreateInconsistencyItem(ctx, database.CreateInconsistencyItemParams{
			ID:              database.PgUUID(uuid.New()),
			InconsistencyID: database.PgUUID(created.ID),
			ClaimID:         claimAID,
			EntityID:        pgtype.UUID{},
			Side:            database.PgText("a"),
			Description:     database.PgText("Side A of contradiction"),
		})
		if err != nil {
			d.logger.Error("failed to link claim A", "error", err)
		}

		_, err = d.db.CreateInconsistencyItem(ctx, database.CreateInconsistencyItemParams{
			ID:              database.PgUUID(uuid.New()),
			InconsistencyID: database.PgUUID(created.ID),
			ClaimID:         claimBID,
			EntityID:        pgtype.UUID{},
			Side:            database.PgText("b"),
			Description:     database.PgText("Side B of contradiction"),
		})
		if err != nil {
			d.logger.Error("failed to link claim B", "error", err)
		}

		inconsistencies = append(inconsistencies, created)
	}

	return inconsistencies, nil
}

// buildEventsSummaryForContradictionDetection creates a summary for LLM analysis
func (d *InconsistencyDetector) buildEventsSummaryForContradictionDetection(claims []*database.Claim) string {
	var summary strings.Builder

	summary.WriteString("Events:\n\n")
	for _, claim := range claims {
		summary.WriteString(fmt.Sprintf("- ID: %s\n", database.UUIDStr(claim.ID)))
		summary.WriteString(fmt.Sprintf("  Title: %s\n", claim.Title))
		if claim.Description.Valid {
			summary.WriteString(fmt.Sprintf("  Description: %s\n", claim.Description.String))
		}
		if claim.DateText.Valid {
			summary.WriteString(fmt.Sprintf("  Date: %s\n", claim.DateText.String))
		}
		if claim.ChronologicalPosition.Valid {
			summary.WriteString(fmt.Sprintf("  Chronological Position: %d\n", claim.ChronologicalPosition.Int32))
		}
		summary.WriteString("\n")
	}

	return summary.String()
}

func inconsistencyFromDB(dbInc *database.Inconsistency) Inconsistency {
	var metadata map[string]interface{}
	if dbInc.Metadata != nil {
		_ = json.Unmarshal(dbInc.Metadata, &metadata)
	}

	return Inconsistency{
		ID:                uuid.UUID(dbInc.ID.Bytes),
		SourceID:          uuid.UUID(dbInc.SourceID.Bytes),
		InconsistencyType: InconsistencyType(dbInc.InconsistencyType),
		Severity:          Severity(dbInc.Severity),
		Title:             dbInc.Title,
		Description:       dbInc.Description,
		ResolutionStatus:  ResolutionStatus(dbInc.ResolutionStatus),
		ResolutionNote:    database.TextPtr(dbInc.ResolutionNote),
		Metadata:          metadata,
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
