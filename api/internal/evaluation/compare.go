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

	// Entity scores
	b.WriteString("Entity Metrics:\n")
	b.WriteString(fmt.Sprintf("  Recall:    %.1f%% (%d found / %d expected)\n",
		result.EntityRecall*100,
		int(float64(len(result.EntityDetails))*result.EntityRecall),
		len(result.EntityDetails)))
	b.WriteString(fmt.Sprintf("  Precision:  %.1f%% (%d correct / %d extracted)\n",
		result.EntityPrecision*100,
		int(float64(len(result.EntityDetails))*result.EntityPrecision),
		len(result.EntityDetails)))
	b.WriteString(fmt.Sprintf("  F1:        %.1f%%\n\n", result.EntityF1*100))

	// Event scores
	b.WriteString("Event Metrics:\n")
	b.WriteString(fmt.Sprintf("  Recall:    %.1f%% (%d found / %d expected)\n",
		result.EventRecall*100,
		int(float64(len(result.EventDetails))*result.EventRecall),
		len(result.EventDetails)))
	b.WriteString(fmt.Sprintf("  Precision:  %.1f%% (%d correct / %d extracted)\n",
		result.EventPrecision*100,
		int(float64(len(result.EventDetails))*result.EventPrecision),
		len(result.EventDetails)))
	b.WriteString(fmt.Sprintf("  F1:        %.1f%%\n\n", result.EventF1*100))

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
	b.WriteString("  False positive <20%%: ")
	if result.FalsePositiveRate < 0.20 {
		b.WriteString("✓ PASS\n")
	} else {
		b.WriteString("✗ FAIL\n")
	}

	b.WriteString("\n")

	// Entity breakdown (incorrect ones only)
	b.WriteString("Entity Extraction Issues:\n")
	incorrectCount := 0
	for _, match := range result.EntityDetails {
		if !match.IsCorrect && match.ManifestID != "" {
			incorrectCount++
			b.WriteString(fmt.Sprintf("  %s (%s): %s\n", match.ManifestID, match.ManifestLabel, match.MatchMethod))
		}
	}
	if incorrectCount == 0 {
		b.WriteString("  None! All entities correctly extracted.\n")
	}
	b.WriteString("\n")

	// Event breakdown (incorrect ones only)
	b.WriteString("Event Extraction Issues:\n")
	incorrectCount = 0
	for _, match := range result.EventDetails {
		if !match.IsCorrect && match.ManifestID != "" {
			incorrectCount++
			b.WriteString(fmt.Sprintf("  %s (%s): %s\n", match.ManifestID, match.ManifestLabel, match.MatchMethod))
		}
	}
	if incorrectCount == 0 {
		b.WriteString("  None! All events correctly extracted.\n")
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
