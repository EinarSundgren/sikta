package extraction

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
)

// Deduplicator handles entity deduplication.
type Deduplicator struct {
	db     *database.Queries
	claude *claude.Client
	logger *slog.Logger
	model  string
}

// NewDeduplicator creates a new entity deduplicator.
func NewDeduplicator(db *database.Queries, claude *claude.Client, logger *slog.Logger, model string) *Deduplicator {
	return &Deduplicator{
		db:     db,
		claude: claude,
		logger: logger,
		model:  model,
	}
}

// DeduplicationResult contains the results of deduplication.
type DeduplicationResult struct {
	MergedCount   int
	TotalEntities int
	MergedPairs   []EntityPair
}

// EntityPair represents two entities that were merged.
type EntityPair struct {
	KeptEntityID   string
	MergedEntityID string
	Reasoning      string
}

// DeduplicateEntities deduplicates entities in a document.
func (d *Deduplicator) DeduplicateEntities(ctx context.Context, sourceID string) (*DeduplicationResult, error) {
	d.logger.Info("starting entity deduplication", "source_id", sourceID)

	entities, err := d.db.ListEntitiesBySource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	result := &DeduplicationResult{
		TotalEntities: len(entities),
	}

	// Pass 1: Exact name matches
	exactMatches := d.findExactMatches(entities)
	for _, match := range exactMatches {
		if err := d.mergeEntities(ctx, match.entityA, match.entityB, "exact name match"); err != nil {
			d.logger.Error("failed to merge entities", "error", err)
			continue
		}
		result.MergedCount++
		result.MergedPairs = append(result.MergedPairs, EntityPair{
			KeptEntityID:   database.UUIDStr(match.entityA.ID),
			MergedEntityID: database.UUIDStr(match.entityB.ID),
			Reasoning:      "exact name match",
		})
	}

	// Refresh entity list after merges
	entities, err = d.db.ListEntitiesBySource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return nil, fmt.Errorf("failed to refresh entities: %w", err)
	}

	// Pass 2: Alias matches
	aliasMatches := d.findAliasMatches(entities)
	for _, match := range aliasMatches {
		if match.confidence < 0.85 {
			continue
		}
		reason := fmt.Sprintf("alias match (confidence: %.2f)", match.confidence)
		if err := d.mergeEntities(ctx, match.entityA, match.entityB, reason); err != nil {
			d.logger.Error("failed to merge entities", "error", err)
			continue
		}
		result.MergedCount++
		result.MergedPairs = append(result.MergedPairs, EntityPair{
			KeptEntityID:   database.UUIDStr(match.entityA.ID),
			MergedEntityID: database.UUIDStr(match.entityB.ID),
			Reasoning:      reason,
		})
	}

	// Pass 3: Similarity matches with LLM confirmation for ambiguous cases
	similarityMatches := d.findSimilarityMatches(entities)
	for _, match := range similarityMatches {
		if match.confidence < 0.70 {
			continue
		}

		if match.confidence >= 0.85 {
			reason := fmt.Sprintf("similarity match (confidence: %.2f)", match.confidence)
			if err := d.mergeEntities(ctx, match.entityA, match.entityB, reason); err != nil {
				d.logger.Error("failed to merge entities", "error", err)
				continue
			}
			result.MergedCount++
			result.MergedPairs = append(result.MergedPairs, EntityPair{
				KeptEntityID:   database.UUIDStr(match.entityA.ID),
				MergedEntityID: database.UUIDStr(match.entityB.ID),
				Reasoning:      reason,
			})
			continue
		}

		confirmed, reasoning, err := d.confirmWithLLM(ctx, match.entityA, match.entityB)
		if err != nil {
			d.logger.Error("LLM confirmation failed", "error", err)
			continue
		}

		if confirmed {
			if err := d.mergeEntities(ctx, match.entityA, match.entityB, reasoning); err != nil {
				d.logger.Error("failed to merge entities", "error", err)
				continue
			}
			result.MergedCount++
			result.MergedPairs = append(result.MergedPairs, EntityPair{
				KeptEntityID:   database.UUIDStr(match.entityA.ID),
				MergedEntityID: database.UUIDStr(match.entityB.ID),
				Reasoning:      reasoning,
			})
		}
	}

	d.logger.Info("deduplication complete", "merged", result.MergedCount, "total", result.TotalEntities)

	return result, nil
}

// entityMatch represents two entities that might be the same.
type entityMatch struct {
	entityA    *database.Entity
	entityB    *database.Entity
	confidence float64
}

// findExactMatches finds entities with exact name matches.
func (d *Deduplicator) findExactMatches(entities []*database.Entity) []entityMatch {
	var matches []entityMatch
	seen := make(map[string]bool)

	for _, entityA := range entities {
		idA := database.UUIDStr(entityA.ID)
		if seen[idA] {
			continue
		}

		for _, entityB := range entities {
			idB := database.UUIDStr(entityB.ID)
			if entityA.ID == entityB.ID || seen[idB] {
				continue
			}

			if strings.EqualFold(entityA.Name, entityB.Name) {
				matches = append(matches, entityMatch{
					entityA:    entityA,
					entityB:    entityB,
					confidence: 1.0,
				})
				seen[idB] = true
			}
		}
		seen[idA] = true
	}

	return matches
}

// findAliasMatches finds entities where one is an alias of another.
func (d *Deduplicator) findAliasMatches(entities []*database.Entity) []entityMatch {
	var matches []entityMatch

	for _, entityA := range entities {
		for _, entityB := range entities {
			if entityA.ID == entityB.ID {
				continue
			}

			if entityB.Aliases != nil {
				for _, alias := range entityB.Aliases {
					if strings.EqualFold(entityA.Name, alias) {
						matches = append(matches, entityMatch{
							entityA:    entityB,
							entityB:    entityA,
							confidence: 0.90,
						})
						break
					}
				}
			}

			if entityA.Aliases != nil {
				for _, alias := range entityA.Aliases {
					if strings.EqualFold(entityB.Name, alias) {
						matches = append(matches, entityMatch{
							entityA:    entityA,
							entityB:    entityB,
							confidence: 0.90,
						})
						break
					}
				}
			}
		}
	}

	return matches
}

// findSimilarityMatches finds entities with similar names.
func (d *Deduplicator) findSimilarityMatches(entities []*database.Entity) []entityMatch {
	var matches []entityMatch

	for _, entityA := range entities {
		for _, entityB := range entities {
			if entityA.ID == entityB.ID {
				continue
			}

			if entityA.EntityType != entityB.EntityType {
				continue
			}

			similarity := calculateSimilarity(entityA.Name, entityB.Name)

			if similarity >= 0.70 {
				matches = append(matches, entityMatch{
					entityA:    entityA,
					entityB:    entityB,
					confidence: similarity,
				})
			}
		}
	}

	return matches
}

// confirmWithLLM asks LLM if two entities are the same.
func (d *Deduplicator) confirmWithLLM(ctx context.Context, entityA, entityB *database.Entity) (bool, string, error) {
	context := fmt.Sprintf("Entity A: %s (%s), Entity B: %s (%s)",
		entityA.Name, entityA.EntityType, entityB.Name, entityB.EntityType)

	prompt := fmt.Sprintf("You are determining if two entity names refer to the same person in a novel.\n\n"+
		"Given:\n"+
		"- Entity A: %s (type: %s)\n"+
		"- Entity B: %s (type: %s)\n"+
		"- Context: %s\n\n"+
		"Consider:\n"+
		"- Are these the same person? (Y/N)\n"+
		"- Is one a nickname or alias of the other?\n"+
		"- Could they be different people with similar names?\n"+
		"- Does the text support them being the same?\n\n"+
		"Return JSON:\n"+
		`{"same_entity": boolean, "confidence": float (0.0-1.0), "reasoning": string}`,
		entityA.Name, entityA.EntityType, entityB.Name, entityB.EntityType, context)

	resp, err := d.claude.SendSystemPrompt(ctx, "", prompt, d.model)
	if err != nil {
		return false, "", err
	}

	if len(resp.Content) == 0 {
		return false, "", fmt.Errorf("empty response from LLM")
	}

	responseText := resp.Content[0].Text
	sameEntity := strings.Contains(strings.ToLower(responseText), `"same_entity": true`)

	reasoning := "LLM confirmed"
	if strings.Contains(responseText, `"reasoning"`) {
		reasoning = "LLM confirmed based on context"
	}

	return sameEntity, reasoning, nil
}

// mergeEntities merges two entities.
func (d *Deduplicator) mergeEntities(ctx context.Context, keep, merge *database.Entity, reason string) error {
	var combinedAliases []string
	if keep.Aliases != nil {
		combinedAliases = append(combinedAliases, keep.Aliases...)
	}
	if merge.Aliases != nil {
		combinedAliases = append(combinedAliases, merge.Aliases...)
	}
	combinedAliases = append(combinedAliases, merge.Name)

	// Update appearance range
	firstAppearance := keep.FirstAppearanceChunk
	lastAppearance := keep.LastAppearanceChunk
	if merge.FirstAppearanceChunk.Valid && (!firstAppearance.Valid || merge.FirstAppearanceChunk.Int32 < firstAppearance.Int32) {
		firstAppearance = merge.FirstAppearanceChunk
	}
	if merge.LastAppearanceChunk.Valid && (!lastAppearance.Valid || merge.LastAppearanceChunk.Int32 > lastAppearance.Int32) {
		lastAppearance = merge.LastAppearanceChunk
	}
	_ = firstAppearance
	_ = lastAppearance

	_, err := d.db.UpdateEntityAliases(ctx, database.UpdateEntityAliasesParams{
		ID:      keep.ID,
		Aliases: combinedAliases,
	})
	if err != nil {
		return fmt.Errorf("failed to update aliases: %w", err)
	}

	d.logger.Info("merged entities",
		"kept", keep.Name,
		"merged", merge.Name,
		"reason", reason)

	return nil
}

// calculateSimilarity calculates string similarity using Levenshtein distance.
func calculateSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}

	distance := levenshteinDistance(a, b)
	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - (float64(distance) / float64(maxLen))
}

// levenshteinDistance calculates the edit distance between two strings.
func levenshteinDistance(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	lenA := len(a)
	lenB := len(b)

	dp := make([][]int, lenA+1)
	for i := range dp {
		dp[i] = make([]int, lenB+1)
		dp[i][0] = i
	}
	for j := range dp[0] {
		dp[0][j] = j
	}

	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			dp[i][j] = minOf3(
				dp[i-1][j]+1,
				dp[i][j-1]+1,
				dp[i-1][j-1]+cost,
			)
		}
	}

	return dp[lenA][lenB]
}

func minOf3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

