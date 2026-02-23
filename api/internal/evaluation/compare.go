package evaluation

import (
	"fmt"
	"strings"
	"time"
)

// Compare compares two ScoreResult structs and produces a diff result
func Compare(resultA, resultB *ScoreResult) *DiffResult {
	if resultA.Corpus != resultB.Corpus {
		return nil // Cannot compare different corpora
	}

	diff := &DiffResult{
		Corpus:    resultA.Corpus,
		VersionA:  resultA.PromptVersion,
		VersionB:  resultB.PromptVersion,
		Timestamp: time.Now(),
	}

	// Calculate entity metric deltas
	diff.EntityRecallDelta = resultB.EntityRecall - resultA.EntityRecall
	diff.EntityPrecisionDelta = resultB.EntityPrecision - resultA.EntityPrecision
	diff.EntityF1Delta = resultB.EntityF1 - resultA.EntityF1

	// Calculate event metric deltas
	diff.EventRecallDelta = resultB.EventRecall - resultA.EventRecall
	diff.EventPrecisionDelta = resultB.EventPrecision - resultA.EventPrecision
	diff.EventF1Delta = resultB.EventF1 - resultA.EventF1

	// Identify improved and regressed entities
	entityStatusA := getEntityStatus(resultA.EntityDetails)
	entityStatusB := getEntityStatus(resultB.EntityDetails)

	for id, statusA := range entityStatusA {
		statusB, ok := entityStatusB[id]
		if !ok {
			continue // Entity only in A
		}

		// Check if status improved (false -> true)
		if !statusA && statusB {
			diff.ImprovedEntities = append(diff.ImprovedEntities, id)
		}

		// Check if status regressed (true -> false)
		if statusA && !statusB {
			diff.RegressedEntities = append(diff.RegressedEntities, id)
		}
	}

	// Identify improved and regressed events
	eventStatusA := getEventStatus(resultA.EventDetails)
	eventStatusB := getEventStatus(resultB.EventDetails)

	for id, statusA := range eventStatusA {
		statusB, ok := eventStatusB[id]
		if !ok {
			continue // Event only in A
		}

		if !statusA && statusB {
			diff.ImprovedEvents = append(diff.ImprovedEvents, id)
		}

		if statusA && !statusB {
			diff.RegressedEvents = append(diff.RegressedEvents, id)
		}
	}

	return diff
}

// getEntityStatus returns a map of entity ID -> whether it was correctly extracted
func getEntityStatus(matches []EntityMatch) map[string]bool {
	status := make(map[string]bool)
	for _, match := range matches {
		if match.ManifestID != "" {
			status[match.ManifestID] = match.IsCorrect
		}
	}
	return status
}

// getEventStatus returns a map of event ID -> whether it was correctly extracted
func getEventStatus(matches []EventMatch) map[string]bool {
	status := make(map[string]bool)
	for _, match := range matches {
		if match.ManifestID != "" {
			status[match.ManifestID] = match.IsCorrect
		}
	}
	return status
}

// FormatTerminal formats a ScoreResult for terminal output
func FormatTerminal(result *ScoreResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("=== Extraction Validation Results: %s (prompt %s) ===\n\n", result.Corpus, result.PromptVersion))
	b.WriteString(fmt.Sprintf("Timestamp: %s\n\n", result.Timestamp.Format("2006-01-02 15:04:05")))

	// Count manifest items vs hallucinations
	manifestEntities := 0
	foundEntities := 0
	manifestEvents := 0
	foundEvents := 0
	var entityHallucinations []EntityMatch
	var eventHallucinations []EventMatch

	for _, m := range result.EntityDetails {
		if m.ManifestID != "" {
			manifestEntities++
			if m.IsCorrect {
				foundEntities++
			}
		} else if m.IsHallucination {
			entityHallucinations = append(entityHallucinations, m)
		}
	}
	for _, m := range result.EventDetails {
		if m.ManifestID != "" {
			manifestEvents++
			if m.IsCorrect {
				foundEvents++
			}
		} else if m.IsHallucination {
			eventHallucinations = append(eventHallucinations, m)
		}
	}

	// Entity scores
	b.WriteString("Entity Metrics:\n")
	b.WriteString(fmt.Sprintf("  Recall:    %.1f%% (%d found / %d expected)\n",
		result.EntityRecall*100, foundEntities, manifestEntities))
	b.WriteString(fmt.Sprintf("  Precision:  %.1f%% (%d correct / %d extracted)\n",
		result.EntityPrecision*100, foundEntities, manifestEntities+len(entityHallucinations)))
	b.WriteString(fmt.Sprintf("  F1:        %.1f%%\n\n", result.EntityF1*100))

	// Event scores
	b.WriteString("Event Metrics:\n")
	b.WriteString(fmt.Sprintf("  Recall:    %.1f%% (%d found / %d expected)\n",
		result.EventRecall*100, foundEvents, manifestEvents))
	b.WriteString(fmt.Sprintf("  Precision:  %.1f%% (%d correct / %d extracted)\n",
		result.EventPrecision*100, foundEvents, manifestEvents+len(eventHallucinations)))
	b.WriteString(fmt.Sprintf("  F1:        %.1f%%\n\n", result.EventF1*100))

	// Inconsistency scores
	if len(result.InconsistencyDetails) > 0 {
		foundInconsistencies := 0
		for _, m := range result.InconsistencyDetails {
			if m.IsCorrect {
				foundInconsistencies++
			}
		}
		b.WriteString("Inconsistency Metrics:\n")
		b.WriteString(fmt.Sprintf("  Recall:    %.1f%% (%d found / %d expected)\n",
			result.InconsistencyRecall*100, foundInconsistencies, len(result.InconsistencyDetails)))
		b.WriteString(fmt.Sprintf("  Precision: %.1f%%\n\n", result.InconsistencyPrecision*100))
	}

	// Quality metrics
	b.WriteString("Quality Metrics:\n")
	b.WriteString(fmt.Sprintf("  False Positive Rate:  %.1f%%\n", result.FalsePositiveRate*100))
	b.WriteString(fmt.Sprintf("  Confidence Accuracy: %.1f%%\n\n", result.AvgConfidenceAccuracy*100))

	// Go/Kill thresholds
	b.WriteString("Go/Kill Thresholds:\n")
	b.WriteString("  Entity recall ≥85%%:  ")
	if result.EntityRecall >= 0.85 {
		b.WriteString("✓ PASS\n")
	} else {
		b.WriteString("✗ FAIL\n")
	}
	b.WriteString("  Event recall ≥70%%:   ")
	if result.EventRecall >= 0.70 {
		b.WriteString("✓ PASS\n")
	} else {
		b.WriteString("✗ FAIL\n")
	}
	if len(result.InconsistencyDetails) > 0 {
		b.WriteString("  Inconsistency ≥50%%:  ")
		if result.InconsistencyRecall >= 0.50 {
			b.WriteString("✓ PASS\n")
		} else {
			b.WriteString("✗ FAIL\n")
		}
	}
	b.WriteString("  False positive <20%%: ")
	if result.FalsePositiveRate < 0.20 {
		b.WriteString("✓ PASS\n\n")
	} else {
		b.WriteString("✗ FAIL\n\n")
	}

	// Entity matches section
	b.WriteString("=== Entity Matches ===\n")
	for _, m := range result.EntityDetails {
		if m.ManifestID != "" {
			if m.IsCorrect {
				b.WriteString(fmt.Sprintf("  ✓ %s: \"%s\" → \"%s\" [%s]\n",
					m.ManifestID, m.ManifestLabel, m.MatchedNodeLabel, m.MatchMethod))
			} else {
				b.WriteString(fmt.Sprintf("  ✗ %s: \"%s\" → NOT FOUND\n",
					m.ManifestID, m.ManifestLabel))
			}
		}
	}
	b.WriteString("\n")

	// Event matches section
	b.WriteString("=== Event Matches ===\n")
	for _, m := range result.EventDetails {
		if m.ManifestID != "" {
			if m.IsCorrect {
				b.WriteString(fmt.Sprintf("  ✓ %s: \"%s\" → \"%s\" [%s]\n",
					m.ManifestID, m.ManifestLabel, m.MatchedNodeLabel, m.MatchMethod))
			} else {
				b.WriteString(fmt.Sprintf("  ✗ %s: \"%s\" → NOT FOUND\n",
					m.ManifestID, m.ManifestLabel))
			}
		}
	}
	b.WriteString("\n")

	// Inconsistency matches section
	if len(result.InconsistencyDetails) > 0 {
		b.WriteString("=== Inconsistency Matches ===\n")
		for _, m := range result.InconsistencyDetails {
			if m.IsCorrect {
				matchMethod := "deterministic"
				if m.MatchedByLLM {
					matchMethod = "llm_judge"
				}
				b.WriteString(fmt.Sprintf("  ✓ %s: [%s/%s] → \"%s\" [%s]\n",
					m.ManifestID, m.ManifestType, m.ManifestSeverity, m.MatchedID, matchMethod))
			} else {
				b.WriteString(fmt.Sprintf("  ✗ %s: [%s/%s] → NOT FOUND\n",
					m.ManifestID, m.ManifestType, m.ManifestSeverity))
			}
		}
		b.WriteString("\n")
	}

	// False positives section
	b.WriteString("=== False Positives (Hallucinations) ===\n")
	if len(entityHallucinations) == 0 && len(eventHallucinations) == 0 {
		b.WriteString("  None!\n")
	} else {
		if len(entityHallucinations) > 0 {
			b.WriteString(fmt.Sprintf("  Entities (%d):\n", len(entityHallucinations)))
			for _, m := range entityHallucinations {
				b.WriteString(fmt.Sprintf("    - \"%s\"\n", m.MatchedNodeLabel))
			}
		}
		if len(eventHallucinations) > 0 {
			b.WriteString(fmt.Sprintf("  Events (%d):\n", len(eventHallucinations)))
			for _, m := range eventHallucinations {
				b.WriteString(fmt.Sprintf("    - \"%s\"\n", m.MatchedNodeLabel))
			}
		}
	}

	return b.String()
}

// FormatDiffTerminal formats a DiffResult for terminal output
func FormatDiffTerminal(diff *DiffResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("=== Comparison: %s → %s (corpus %s) ===\n\n", diff.VersionA, diff.VersionB, diff.Corpus))

	// Entity metric changes
	b.WriteString("Entity Metrics:\n")
	formatDelta(&b, "Recall", diff.EntityRecallDelta)
	formatDelta(&b, "Precision", diff.EntityPrecisionDelta)
	formatDelta(&b, "F1", diff.EntityF1Delta)
	b.WriteString("\n")

	// Event metric changes
	b.WriteString("Event Metrics:\n")
	formatDelta(&b, "Recall", diff.EventRecallDelta)
	formatDelta(&b, "Precision", diff.EventPrecisionDelta)
	formatDelta(&b, "F1", diff.EventF1Delta)
	b.WriteString("\n")

	// Improved entities
	if len(diff.ImprovedEntities) > 0 {
		b.WriteString(fmt.Sprintf("Improved Entities (%d): %s\n", len(diff.ImprovedEntities), joinIDs(diff.ImprovedEntities)))
	}

	// Regressed entities
	if len(diff.RegressedEntities) > 0 {
		b.WriteString(fmt.Sprintf("Regressed Entities (%d): %s\n", len(diff.RegressedEntities), joinIDs(diff.RegressedEntities)))
	}

	// Improved events
	if len(diff.ImprovedEvents) > 0 {
		b.WriteString(fmt.Sprintf("Improved Events (%d): %s\n", len(diff.ImprovedEvents), joinIDs(diff.ImprovedEvents)))
	}

	// Regressed events
	if len(diff.RegressedEvents) > 0 {
		b.WriteString(fmt.Sprintf("Regressed Events (%d): %s\n", len(diff.RegressedEvents), joinIDs(diff.RegressedEvents)))
	}

	return b.String()
}

// formatDelta formats a metric delta with appropriate symbols
func formatDelta(b *strings.Builder, name string, delta float64) {
	symbol := "±"
	if delta > 0 {
		symbol = "+"
	} else if delta < 0 {
		symbol = ""
	}
	b.WriteString(fmt.Sprintf("  %s: %s%.1f%%\n", name, symbol, delta*100))
}

// joinIDs joins a slice of IDs into a comma-separated string
func joinIDs(ids []string) string {
	if len(ids) == 0 {
		return "none"
	}
	if len(ids) <= 5 {
		return strings.Join(ids, ", ")
	}
	return strings.Join(ids[:5], ", ") + fmt.Sprintf(" ... (%d more)", len(ids)-5)
}
