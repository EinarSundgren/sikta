package evaluation

import (
	"time"
)

// Scorer calculates metrics from match results
type Scorer struct {
	manifest  *Manifest
	extraction *Extraction
}

// NewScorer creates a new scorer
func NewScorer(manifest *Manifest, extraction *Extraction) *Scorer {
	return &Scorer{
		manifest:  manifest,
		extraction: extraction,
	}
}

// Score computes all metrics for an extraction run
func (s *Scorer) Score() *ScoreResult {
	// Match entities
	entityMatcher := NewEntityMatcher(s.manifest, s.extraction)
	entityMatches := entityMatcher.MatchEntities()

	// Match events
	eventMatcher := NewEventMatcher(s.manifest, s.extraction, entityMatches)
	eventMatches := eventMatcher.MatchEvents(s.extraction)

	// Calculate entity metrics
	entityRecall, entityPrecision, entityF1 := s.calculateEntityMetrics(entityMatches)

	// Calculate event metrics
	eventRecall, eventPrecision, eventF1 := s.calculateEventMetrics(eventMatches)

	// Calculate false positive rate
	falsePositiveRate := s.calculateFalsePositiveRate(entityMatches, eventMatches)

	// Calculate confidence accuracy
	avgConfidenceAccuracy := s.calculateConfidenceAccuracy(entityMatches, eventMatches)

	return &ScoreResult{
		Corpus:              s.manifest.Corpus,
		PromptVersion:       s.extraction.PromptVersion,
		Timestamp:           time.Now(),
		EntityRecall:        entityRecall,
		EntityPrecision:     entityPrecision,
		EntityF1:           entityF1,
		EntityDetails:       entityMatches,
		EventRecall:         eventRecall,
		EventPrecision:      eventPrecision,
		EventF1:            eventF1,
		EventDetails:        eventMatches,
		FalsePositiveRate:   falsePositiveRate,
		AvgConfidenceAccuracy: avgConfidenceAccuracy,
	}
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
