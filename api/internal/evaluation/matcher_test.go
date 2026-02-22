package evaluation

import (
	"testing"
)

// TestEntityMatcherBRF tests entity matching against BRF corpus manifest
func TestEntityMatcherBRF(t *testing.T) {
	// Create a mini manifest based on BRF corpus
	manifest := &Manifest{
		Corpus: "brf",
		Entities: []ManifestEntity{
			{
				ID:    "E1",
				Label: "Anna Lindqvist",
				Type:  "person",
				Aliases: []string{"ordföranden", "ordförande", "Anna", "Lindqvist"},
			},
			{
				ID:    "E8",
				Label: "Brf Stenbacken 3",
				Type:  "organization",
				Aliases: []string{"föreningen", "BRF Stenbacken 3", "Stenbacken", "brf"},
			},
		},
	}

	// Simulate extraction output with some correct and some incorrect extractions
	extraction := &Extraction{
		Corpus: "brf",
		PromptVersion: "v1",
		Nodes: []ExtractedNode{
			{
				ID:    "node-1",
				NodeType: "person",
				Label: "Anna Lindqvist",
				Confidence: 0.95,
				Modality: "asserted",
			},
			{
				ID:    "node-2",
				NodeType: "person",
				Label: "ordföranden", // Should match via alias
				Confidence: 0.90,
				Modality: "asserted",
			},
			{
				ID:    "node-3",
				NodeType: "organization",
				Label: "Brf Stenbacken 3",
				Confidence: 1.0,
				Modality: "asserted",
			},
			{
				ID:    "node-4",
				NodeType: "organization",
				Label: "Stenbacken", // Should match via alias
				Confidence: 0.85,
				Modality: "asserted",
			},
			{
				ID:    "node-5",
				NodeType: "person",
				Label: "Hallucinated Person", // Not in manifest
				Confidence: 0.5,
				Modality: "asserted",
			},
		},
	}

	// Run matcher
	matcher := NewEntityMatcher(manifest, extraction)
	matches := matcher.MatchEntities()

	// Verify results
	// We expect: 2 manifest entities + potentially multiple entries for duplicates/aliases + hallucinations
	// The matcher creates one entry per manifest entity, plus hallucination entries
	// So minimum is 2 (2 manifest entities), but we also have a hallucination
	if len(matches) < 2 { // At least 2 for manifest entities
		t.Errorf("Expected at least 2 matches, got %d", len(matches))
	}

	// Find matches for each manifest entity
	var annaMatch, brfMatch *EntityMatch
	var hallucination *EntityMatch

	for i := range matches {
		if matches[i].ManifestID == "E1" {
			annaMatch = &matches[i]
		}
		if matches[i].ManifestID == "E8" {
			brfMatch = &matches[i]
		}
		if matches[i].IsHallucination {
			hallucination = &matches[i]
		}
	}

	// Verify Anna Lindqvist match
	if annaMatch == nil {
		t.Fatal("Anna Lindqvist not found in matches")
	}
	if !annaMatch.IsCorrect {
		t.Error("Anna Lindqvist should be marked as correct")
	}
	if annaMatch.MatchMethod != "exact" {
		t.Errorf("Anna Lindqvist should match via exact method, got %s", annaMatch.MatchMethod)
	}

	// Verify Brf Stenbacken 3 match
	if brfMatch == nil {
		t.Fatal("Brf Stenbacken 3 not found in matches")
	}
	if !brfMatch.IsCorrect {
		t.Error("Brf Stenbacken 3 should be marked as correct")
	}

	// Verify hallucination detected
	if hallucination == nil {
		t.Fatal("Hallucination not detected")
	}
	if !hallucination.IsHallucination {
		t.Error("Should detect hallucinated entity")
	}

	// Test scorer
	scorer := NewScorer(manifest, extraction)
	score := scorer.Score()

	// Expected: 2/2 entities found (recall=1.0)
	// Precision depends on how many duplicates we count, but should be reasonable
	if score.EntityRecall != 1.0 {
		t.Errorf("Expected entity recall 1.0, got %.2f", score.EntityRecall)
	}
	// Just check precision is non-zero and reasonable
	if score.EntityPrecision < 0.1 || score.EntityPrecision > 1.0 {
		t.Errorf("Expected entity precision between 0.1 and 1.0, got %.2f", score.EntityPrecision)
	}
}

// TestLevenshteinDistance tests the edit distance calculation
func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b string
		expected int
	}{
		{"test", "test", 0},
		{"test", "tost", 1},
		{"katt", "hatt", 1},
		{"Anna", "Ana", 1}, // Delete n
		{"Lindqvist", "Lindquist", 1},
	}

	for _, tt := range tests {
		result := levenshteinDistance(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// TestNormalizeText tests text normalization
func TestNormalizeText(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"Anna Lindqvist", "anna lindqvist"},
		{"Ordföranden Anna Lindqvist", "anna lindqvist"},
		{"  Anna  Lindqvist  ", "anna lindqvist"}, // normalizeText now collapses spaces
	}

	for _, tt := range tests {
		result := normalizeText(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeText(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
