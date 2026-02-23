package evaluation

import (
	"context"
	"log/slog"
	"time"
)

// Scorer calculates metrics from match results
type Scorer struct {
	manifest           *Manifest
	extraction         *Extraction
	judge              *EventJudge
	inconsistencyJudge *InconsistencyJudge
	logger             *slog.Logger
}

// NewScorer creates a new scorer
func NewScorer(manifest *Manifest, extraction *Extraction) *Scorer {
	return &Scorer{
		manifest:           manifest,
		extraction:         extraction,
		judge:              nil,
		inconsistencyJudge: nil,
		logger:             nil,
	}
}

// NewScorerWithJudge creates a new scorer with an LLM judge for semantic matching
func NewScorerWithJudge(manifest *Manifest, extraction *Extraction, judge *EventJudge, logger *slog.Logger) *Scorer {
	return &Scorer{
		manifest:           manifest,
		extraction:         extraction,
		judge:              judge,
		inconsistencyJudge: nil,
		logger:             logger,
	}
}

// NewScorerWithJudges creates a new scorer with both event and inconsistency judges
func NewScorerWithJudges(manifest *Manifest, extraction *Extraction, judge *EventJudge, inconsistencyJudge *InconsistencyJudge, logger *slog.Logger) *Scorer {
	return &Scorer{
		manifest:           manifest,
		extraction:         extraction,
		judge:              judge,
		inconsistencyJudge: inconsistencyJudge,
		logger:             logger,
	}
}

// Score computes all metrics for an extraction run
func (s *Scorer) Score() *ScoreResult {
	return s.ScoreWithContext(context.Background())
}

// ScoreWithContext computes all metrics with optional LLM judge support
func (s *Scorer) ScoreWithContext(ctx context.Context) *ScoreResult {
	// Match entities
	entityMatcher := NewEntityMatcher(s.manifest, s.extraction)
	entityMatches := entityMatcher.MatchEntities()

	// Match events
	eventMatcher := NewEventMatcher(s.manifest, s.extraction, entityMatches)
	eventMatches := eventMatcher.MatchEvents(s.extraction)

	// If judge is available, run LLM matching on unmatched events
	if s.judge != nil {
		eventMatches = s.runJudgePass(ctx, eventMatches)
	}

	// Match inconsistencies
	inconsistencyMatches := s.matchInconsistencies(ctx)

	// Calculate entity metrics
	entityRecall, entityPrecision, entityF1 := s.calculateEntityMetrics(entityMatches)

	// Calculate event metrics
	eventRecall, eventPrecision, eventF1 := s.calculateEventMetrics(eventMatches)

	// Calculate inconsistency metrics
	inconsistencyRecall, inconsistencyPrecision := s.calculateInconsistencyMetrics(inconsistencyMatches)

	// Calculate false positive rate
	falsePositiveRate := s.calculateFalsePositiveRate(entityMatches, eventMatches)

	// Calculate confidence accuracy
	avgConfidenceAccuracy := s.calculateConfidenceAccuracy(entityMatches, eventMatches)

	return &ScoreResult{
		Corpus:                 s.manifest.Corpus,
		PromptVersion:          s.extraction.PromptVersion,
		Timestamp:              time.Now(),
		EntityRecall:           entityRecall,
		EntityPrecision:        entityPrecision,
		EntityF1:               entityF1,
		EntityDetails:          entityMatches,
		EventRecall:            eventRecall,
		EventPrecision:         eventPrecision,
		EventF1:                eventF1,
		EventDetails:           eventMatches,
		InconsistencyRecall:    inconsistencyRecall,
		InconsistencyPrecision: inconsistencyPrecision,
		InconsistencyDetails:   inconsistencyMatches,
		FalsePositiveRate:      falsePositiveRate,
		AvgConfidenceAccuracy:  avgConfidenceAccuracy,
	}
}

// runJudgePass runs the LLM judge on unmatched events
func (s *Scorer) runJudgePass(ctx context.Context, eventMatches []EventMatch) []EventMatch {
	// Collect unmatched manifest events
	var unmatchedManifest []ManifestEvent
	for _, match := range eventMatches {
		if match.ManifestID != "" && !match.IsCorrect {
			// Find the manifest event
			for _, evt := range s.manifest.Events {
				if evt.ID == match.ManifestID {
					unmatchedManifest = append(unmatchedManifest, evt)
					break
				}
			}
		}
	}

	// Collect unmatched extracted events
	// Only count labels from CORRECT matches (not hallucinations)
	matchedLabels := make(map[string]bool)
	for _, match := range eventMatches {
		if match.MatchedNodeLabel != "" && match.IsCorrect {
			matchedLabels[match.MatchedNodeLabel] = true
		}
	}
	var unmatchedExtracted []ExtractedNode
	for i := range s.extraction.Nodes {
		node := &s.extraction.Nodes[i]
		if node.NodeType == "event" && !matchedLabels[node.Label] {
			unmatchedExtracted = append(unmatchedExtracted, *node)
		}
	}

	// Always log judge status at Info level
	if s.logger != nil {
		s.logger.Info("judge pass analysis",
			"unmatched_manifest", len(unmatchedManifest),
			"unmatched_extracted", len(unmatchedExtracted),
			"total_events", len(eventMatches))
	}

	if len(unmatchedManifest) == 0 || len(unmatchedExtracted) == 0 {
		return eventMatches
	}

	// Run judge
	judgeMatches, err := s.judge.JudgeUnmatchedEvents(ctx, unmatchedManifest, unmatchedExtracted)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("judge pass failed", "error", err)
		}
		return eventMatches
	}

	// Merge judge results back into event matches
	for _, judgeMatch := range judgeMatches {
		for i := range eventMatches {
			if eventMatches[i].ManifestID == judgeMatch.ManifestID {
				eventMatches[i] = judgeMatch
				break
			}
		}
	}

	return eventMatches
}

// calculateEntityMetrics computes entity recall, precision, and F1
func (s *Scorer) calculateEntityMetrics(matches []EntityMatch) (recall, precision, f1 float64) {
	var found, correct, total int

	for _, match := range matches {
		if match.ManifestID != "" {
			total++ // This is a manifest entity
			if match.IsCorrect {
				found++
				correct++
			}
		}
	}

	extractedCount := len(matches)

	// Recall: found / expected (avoid division by zero)
	if total > 0 {
		recall = float64(found) / float64(total)
	}

	// Precision: correct / extracted (avoid division by zero)
	if extractedCount > 0 {
		precision = float64(correct) / float64(extractedCount)
	}

	// F1: harmonic mean
	if recall+precision > 0 {
		f1 = 2 * (recall * precision) / (recall + precision)
	}

	return recall, precision, f1
}

// calculateEventMetrics computes event recall, precision, and F1
func (s *Scorer) calculateEventMetrics(matches []EventMatch) (recall, precision, f1 float64) {
	var found, correct, total int

	for _, match := range matches {
		if match.ManifestID != "" {
			total++ // This is a manifest event
			if match.IsCorrect {
				found++
				correct++
			}
		}
	}

	extractedCount := len(matches)

	// Recall: found / expected
	if total > 0 {
		recall = float64(found) / float64(total)
	}

	// Precision: correct / extracted
	if extractedCount > 0 {
		precision = float64(correct) / float64(extractedCount)
	}

	// F1: harmonic mean
	if recall+precision > 0 {
		f1 = 2 * (recall * precision) / (recall + precision)
	}

	return recall, precision, f1
}

// calculateFalsePositiveRate computes the rate of hallucinated extractions
func (s *Scorer) calculateFalsePositiveRate(entityMatches []EntityMatch, eventMatches []EventMatch) float64 {
	var hallucinations, total int

	for _, match := range entityMatches {
		total++
		if match.IsHallucination {
			hallucinations++
		}
	}

	for _, match := range eventMatches {
		total++
		if match.IsHallucination {
			hallucinations++
		}
	}

	if total == 0 {
		return 0.0
	}

	return float64(hallucinations) / float64(total)
}

// calculateConfidenceAccuracy measures how well confidence scores predict correctness
func (s *Scorer) calculateConfidenceAccuracy(entityMatches []EntityMatch, eventMatches []EventMatch) float64 {
	// This is a simplified metric: average confidence of correct extractions vs. incorrect
	// In practice, we'd want to measure calibration (do high-confidence items have higher accuracy?)
	// For now, return a placeholder as we don't have the node data in match results
	// This would be enhanced in a full implementation

	return 0.0 // Placeholder
}

// matchInconsistencies attempts to match detected inconsistencies against manifest
func (s *Scorer) matchInconsistencies(ctx context.Context) []InconsistencyMatch {
	// If no inconsistencies in extraction or manifest, return empty
	if len(s.extraction.Inconsistencies) == 0 || len(s.manifest.Inconsistencies) == 0 {
		if s.logger != nil {
			s.logger.Info("inconsistency matching skipped",
				"manifest_count", len(s.manifest.Inconsistencies),
				"extracted_count", len(s.extraction.Inconsistencies))
		}
		return nil
	}

	// If inconsistency judge is available, use LLM matching
	if s.inconsistencyJudge != nil {
		return s.runInconsistencyJudgePass(ctx)
	}

	// Otherwise, do simple deterministic matching by description similarity
	return s.matchInconsistenciesDeterministic()
}

// runInconsistencyJudgePass uses the LLM judge to match inconsistencies
func (s *Scorer) runInconsistencyJudgePass(ctx context.Context) []InconsistencyMatch {
	// Convert extracted inconsistencies to judge format
	candidates := make([]DetectedInconsistencyForJudge, len(s.extraction.Inconsistencies))
	for i, inc := range s.extraction.Inconsistencies {
		candidates[i] = DetectedInconsistencyForJudge{
			ID:               inc.ID,
			Type:             inc.Type,
			Severity:         inc.Severity,
			Description:      inc.Description,
			Documents:        inc.Documents,
			EntitiesInvolved: inc.EntitiesInvolved,
			Evidence:         evidenceToMap(inc.Evidence),
			Confidence:       inc.Confidence,
		}
	}

	if s.logger != nil {
		s.logger.Info("running inconsistency judge",
			"manifest_count", len(s.manifest.Inconsistencies),
			"candidate_count", len(candidates))
	}

	matches, err := s.inconsistencyJudge.JudgeInconsistencies(ctx, s.manifest.Inconsistencies, candidates)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("inconsistency judge failed", "error", err)
		}
		return nil
	}

	// Add unmatched manifest inconsistencies as failed matches
	matchedIDs := make(map[string]bool)
	for _, match := range matches {
		matchedIDs[match.ManifestID] = true
	}

	for _, inc := range s.manifest.Inconsistencies {
		if !matchedIDs[inc.ID] {
			matches = append(matches, InconsistencyMatch{
				ManifestID:       inc.ID,
				ManifestType:     inc.Type,
				ManifestSeverity: inc.Severity,
				MatchedID:        "",
				MatchedByLLM:     false,
				MatchConfidence:  0.0,
				Reasoning:        "No matching inconsistency found",
				IsCorrect:        false,
				IsHallucination:  false,
			})
		}
	}

	return matches
}

// matchInconsistenciesDeterministic does simple string-based matching
func (s *Scorer) matchInconsistenciesDeterministic() []InconsistencyMatch {
	var matches []InconsistencyMatch

	for _, manifestInc := range s.manifest.Inconsistencies {
		matched := false
		for _, extractedInc := range s.extraction.Inconsistencies {
			// Simple heuristic: same type and overlapping documents
			if manifestInc.Type == extractedInc.Type {
				docOverlap := hasOverlap(manifestInc.Documents, extractedInc.Documents)
				if docOverlap {
					matches = append(matches, InconsistencyMatch{
						ManifestID:       manifestInc.ID,
						ManifestType:     manifestInc.Type,
						ManifestSeverity: manifestInc.Severity,
						MatchedID:        extractedInc.ID,
						MatchedByLLM:     false,
						MatchConfidence:  0.7, // Lower confidence for deterministic match
						Reasoning:        "Matched by type and document overlap",
						IsCorrect:        true,
						IsHallucination:  false,
					})
					matched = true
					break
				}
			}
		}

		if !matched {
			matches = append(matches, InconsistencyMatch{
				ManifestID:       manifestInc.ID,
				ManifestType:     manifestInc.Type,
				ManifestSeverity: manifestInc.Severity,
				MatchedID:        "",
				MatchedByLLM:     false,
				MatchConfidence:  0.0,
				Reasoning:        "No matching inconsistency found",
				IsCorrect:        false,
				IsHallucination:  false,
			})
		}
	}

	return matches
}

// calculateInconsistencyMetrics computes inconsistency recall and precision
func (s *Scorer) calculateInconsistencyMetrics(matches []InconsistencyMatch) (recall, precision float64) {
	if len(s.manifest.Inconsistencies) == 0 {
		return 0.0, 0.0
	}

	var found int
	for _, match := range matches {
		if match.IsCorrect {
			found++
		}
	}

	// Recall: found / expected
	recall = float64(found) / float64(len(s.manifest.Inconsistencies))

	// Precision: correct / detected (if any detected)
	if len(s.extraction.Inconsistencies) > 0 {
		precision = float64(found) / float64(len(s.extraction.Inconsistencies))
	}

	return recall, precision
}

// evidenceToMap converts InconsistencyEvidence to map for judge format
func evidenceToMap(evidence InconsistencyEvidence) map[string]interface{} {
	return map[string]interface{}{
		"side_a": map[string]interface{}{
			"doc":     evidence.SideA.Doc,
			"section": evidence.SideA.Section,
			"claim":   evidence.SideA.Claim,
		},
		"side_b": map[string]interface{}{
			"doc":     evidence.SideB.Doc,
			"section": evidence.SideB.Section,
			"claim":   evidence.SideB.Claim,
		},
	}
}

// hasOverlap checks if two string slices have any common elements
func hasOverlap(a, b []string) bool {
	set := make(map[string]bool)
	for _, s := range a {
		set[s] = true
	}
	for _, s := range b {
		if set[s] {
			return true
		}
	}
	return false
}
