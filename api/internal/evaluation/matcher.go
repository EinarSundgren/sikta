package evaluation

import (
	"strings"
	"unicode"
)

// EntityMatcher handles fuzzy entity matching
type EntityMatcher struct {
	manifestEntities map[string]*ManifestEntity // ID -> entity
	extractedNodes   map[string]*ExtractedNode  // Label -> node (for lookup)
}

// NewEntityMatcher creates a new entity matcher
func NewEntityMatcher(manifest *Manifest, extraction *Extraction) *EntityMatcher {
	em := &EntityMatcher{
		manifestEntities: make(map[string]*ManifestEntity),
		extractedNodes:   make(map[string]*ExtractedNode),
	}

	// Index manifest entities by ID
	for i := range manifest.Entities {
		em.manifestEntities[manifest.Entities[i].ID] = &manifest.Entities[i]
	}

	// Index entity nodes by label (for fuzzy lookup)
	// Events, values, obligations, etc. are not entities and should not be matched
	entityTypes := map[string]bool{
		"person":       true,
		"organization": true,
		"place":        true,
		"address":      true,
		"vehicle":      true,
		"technology":   true,
	}
	for i := range extraction.Nodes {
		node := &extraction.Nodes[i]
		if entityTypes[node.NodeType] {
			label := normalizeText(node.Label)
			em.extractedNodes[label] = node
		}
	}

	return em
}

// MatchEntities matches extracted entities against manifest entities
func (em *EntityMatcher) MatchEntities() []EntityMatch {
	results := make([]EntityMatch, 0, len(em.manifestEntities))

	// Track which extracted nodes have been matched
	matchedExtracted := make(map[string]bool)

	// First pass: try to match each manifest entity to an extracted node
	for _, manifestEnt := range em.manifestEntities {
		match := em.findBestMatch(manifestEnt, matchedExtracted)
		results = append(results, match)
		if match.MatchedNodeID != "" {
			matchedExtracted[match.MatchedNodeLabel] = true
		}
	}

	// Second pass: identify hallucinations (extracted nodes that match nothing)
	for label, extractedNode := range em.extractedNodes {
		if !matchedExtracted[label] {
			results = append(results, EntityMatch{
				ManifestID:       "",
				ManifestLabel:    "",
				MatchedNodeID:    extractedNode.ID,
				MatchedNodeLabel: extractedNode.Label,
				MatchMethod:      "none",
				MatchScore:       0.0,
				IsCorrect:        false,
				IsHallucination:  true,
			})
		}
	}

	return results
}

// findBestMatch finds the best matching extracted node for a manifest entity
func (em *EntityMatcher) findBestMatch(manifestEnt *ManifestEntity, alreadyMatched map[string]bool) EntityMatch {
	bestMatch := EntityMatch{
		ManifestID:    manifestEnt.ID,
		ManifestLabel: manifestEnt.Label,
		MatchMethod:   "none",
		MatchScore:    0.0,
		IsCorrect:     false,
	}
	_ = bestMatch // Suppress unused warning temporarily

	// Try exact match with label first
	if node, ok := em.extractedNodes[normalizeText(manifestEnt.Label)]; ok && !alreadyMatched[node.Label] {
		return EntityMatch{
			ManifestID:       manifestEnt.ID,
			ManifestLabel:    manifestEnt.Label,
			MatchedNodeID:    node.ID,
			MatchedNodeLabel: node.Label,
			MatchMethod:      "exact",
			MatchScore:       1.0,
			IsCorrect:        true,
			IsHallucination:  false,
		}
	}

	// Try exact match with aliases
	for _, alias := range manifestEnt.Aliases {
		if node, ok := em.extractedNodes[normalizeText(alias)]; ok && !alreadyMatched[node.Label] {
			return EntityMatch{
				ManifestID:       manifestEnt.ID,
				ManifestLabel:    manifestEnt.Label,
				MatchedNodeID:    node.ID,
				MatchedNodeLabel: node.Label,
				MatchMethod:      "exact_alias",
				MatchScore:       1.0,
				IsCorrect:        true,
				IsHallucination:  false,
			}
		}
	}

	// No exact match found, try fuzzy matching
	for extractedLabel, extractedNode := range em.extractedNodes {
		if alreadyMatched[extractedNode.Label] {
			continue
		}

		// Check Levenshtein distance against label
		if dist := levenshteinDistance(normalizeText(manifestEnt.Label), extractedLabel); dist <= 3 {
			_ = dist // Use dist in condition
			if dist < 3 {
				// Close match, high confidence
				return EntityMatch{
					ManifestID:       manifestEnt.ID,
					ManifestLabel:    manifestEnt.Label,
					MatchedNodeID:    extractedNode.ID,
					MatchedNodeLabel: extractedNode.Label,
					MatchMethod:      "levenshtein",
					MatchScore:       1.0 - float64(dist)/10.0,
					IsCorrect:        true,
					IsHallucination:  false,
				}
			}
			// Update best match if this is better
			if 1.0-float64(dist)/10.0 > bestMatch.MatchScore {
				bestMatch = EntityMatch{
					ManifestID:       manifestEnt.ID,
					ManifestLabel:    manifestEnt.Label,
					MatchedNodeID:    extractedNode.ID,
					MatchedNodeLabel: extractedNode.Label,
					MatchMethod:      "levenshtein",
					MatchScore:       1.0 - float64(dist)/10.0,
					IsCorrect:        true,
					IsHallucination:  false,
				}
			}
		}

		// Check Levenshtein distance against aliases
		for _, alias := range manifestEnt.Aliases {
			if dist := levenshteinDistance(normalizeText(alias), extractedLabel); dist <= 3 {
				score := 1.0 - float64(dist)/10.0
				if score > bestMatch.MatchScore {
					bestMatch = EntityMatch{
						ManifestID:       manifestEnt.ID,
						ManifestLabel:    manifestEnt.Label,
						MatchedNodeID:    extractedNode.ID,
						MatchedNodeLabel: extractedNode.Label,
						MatchMethod:      "levenshtein_alias",
						MatchScore:       score,
						IsCorrect:        true,
						IsHallucination:  false,
					}
				}
			}
		}

		// Check substring match as fallback
		if strings.Contains(extractedLabel, normalizeText(manifestEnt.Label)) ||
		   strings.Contains(normalizeText(manifestEnt.Label), extractedLabel) {
			if 0.7 > bestMatch.MatchScore {
				bestMatch = EntityMatch{
					ManifestID:       manifestEnt.ID,
					ManifestLabel:    manifestEnt.Label,
					MatchedNodeID:    extractedNode.ID,
					MatchedNodeLabel: extractedNode.Label,
					MatchMethod:      "substring",
					MatchScore:       0.7,
					IsCorrect:        true,
					IsHallucination:  false,
				}
			}
		}

		// Check substring against aliases
		for _, alias := range manifestEnt.Aliases {
			if strings.Contains(extractedLabel, normalizeText(alias)) ||
			   strings.Contains(normalizeText(alias), extractedLabel) {
				if 0.7 > bestMatch.MatchScore {
					bestMatch = EntityMatch{
						ManifestID:       manifestEnt.ID,
						ManifestLabel:    manifestEnt.Label,
						MatchedNodeID:    extractedNode.ID,
						MatchedNodeLabel: extractedNode.Label,
						MatchMethod:      "substring_alias",
						MatchScore:       0.7,
						IsCorrect:        true,
						IsHallucination:  false,
					}
				}
			}
		}
	}

	return bestMatch
}

// EventMatcher handles event matching based on entity involvement and source
type EventMatcher struct {
	manifestEvents   map[string]*ManifestEvent // ID -> event
	extractedNodes   map[string]*ExtractedNode  // Label -> node (for entity lookup)
	entityMatches    map[string]string         // Manifest entity ID -> Extracted node label
}

// NewEventMatcher creates a new event matcher
func NewEventMatcher(manifest *Manifest, extraction *Extraction, entityMatches []EntityMatch) *EventMatcher {
	em := &EventMatcher{
		manifestEvents: make(map[string]*ManifestEvent),
		extractedNodes: make(map[string]*ExtractedNode),
		entityMatches:  make(map[string]string),
	}

	// Index manifest events by ID
	for i := range manifest.Events {
		em.manifestEvents[manifest.Events[i].ID] = &manifest.Events[i]
	}

	// Index extracted nodes by label
	for i := range extraction.Nodes {
		label := normalizeText(extraction.Nodes[i].Label)
		em.extractedNodes[label] = &extraction.Nodes[i]
	}

	// Build entity match lookup from entity match results
	for _, match := range entityMatches {
		if match.IsCorrect && match.MatchedNodeLabel != "" {
			em.entityMatches[match.ManifestID] = match.MatchedNodeLabel
		}
	}

	return em
}

// MatchEvents matches extracted events against manifest events
func (em *EventMatcher) MatchEvents(extraction *Extraction) []EventMatch {
	results := make([]EventMatch, 0, len(em.manifestEvents))

	// Track which extracted nodes have been matched as events
	matchedExtracted := make(map[string]bool)

	// Match each manifest event to an extracted event node
	for _, manifestEvt := range em.manifestEvents {
		match := em.findBestEventMatch(manifestEvt, extraction, matchedExtracted)
		results = append(results, match)
		if match.MatchedNodeID != "" {
			matchedExtracted[match.MatchedNodeLabel] = true
		}
	}

	// Identify hallucinated events (event nodes that match nothing)
	for i := range extraction.Nodes {
		node := &extraction.Nodes[i]
		if node.NodeType == "event" && !matchedExtracted[node.Label] {
			results = append(results, EventMatch{
				ManifestID:        "",
				ManifestLabel:     "",
				MatchedNodeID:     node.ID,
				MatchedNodeLabel:  node.Label,
				MatchMethod:       "none",
				MatchScore:        0.0,
				EntitiesMatched:   0,
				EntitiesTotal:     0,
				SourceMatch:       false,
				TemporalOverlap:   false,
				IsCorrect:         false,
				IsHallucination:   true,
			})
		}
	}

	return results
}

// findBestEventMatch finds the best matching extracted event for a manifest event
func (em *EventMatcher) findBestEventMatch(manifestEvt *ManifestEvent, extraction *Extraction, alreadyMatched map[string]bool) EventMatch {
	bestMatch := EventMatch{
		ManifestID:  manifestEvt.ID,
		ManifestLabel: manifestEvt.Label,
		MatchMethod: "none",
		MatchScore:  0.0,
		EntitiesTotal: len(manifestEvt.Entities),
		IsCorrect:   false,
	}

	// Find extracted event nodes with simplified matching (label-based for now)
	for i := range extraction.Nodes {
		node := &extraction.Nodes[i]
		if node.NodeType != "event" {
			continue
		}
		if alreadyMatched[node.Label] {
			continue
		}

		// Try exact label match first
		if normalizeText(node.Label) == normalizeText(manifestEvt.Label) {
			bestMatch = EventMatch{
				ManifestID:        manifestEvt.ID,
				ManifestLabel:     manifestEvt.Label,
				MatchedNodeID:     node.ID,
				MatchedNodeLabel:  node.Label,
				MatchMethod:       "exact_label",
				MatchScore:        1.0,
				EntitiesMatched:   len(manifestEvt.Entities),
				EntitiesTotal:     len(manifestEvt.Entities),
				SourceMatch:       true, // Assume source matches if labels match
				TemporalOverlap:   true,
				IsCorrect:         true,
				IsHallucination:   false,
			}
			break // Found exact match, stop looking
		}

		// Try Levenshtein distance on labels (within 3 edits)
		if dist := levenshteinDistance(normalizeText(manifestEvt.Label), normalizeText(node.Label)); dist <= 3 {
			score := 1.0 - float64(dist)/10.0
			if score > bestMatch.MatchScore {
				bestMatch = EventMatch{
					ManifestID:        manifestEvt.ID,
					ManifestLabel:     manifestEvt.Label,
					MatchedNodeID:     node.ID,
					MatchedNodeLabel:  node.Label,
					MatchMethod:       "levenshtein_label",
					MatchScore:        score,
					EntitiesMatched:   len(manifestEvt.Entities),
					EntitiesTotal:     len(manifestEvt.Entities),
					SourceMatch:       true,
					TemporalOverlap:   true,
					IsCorrect:         true,
					IsHallucination:   false,
				}
			}
		}

		// Try keyword overlap matching
		// Split both labels into words, count shared significant words
		normManifest := normalizeText(manifestEvt.Label)
		normExtracted := normalizeText(node.Label)
		keywordScore := keywordOverlapScore(normManifest, normExtracted)
		if keywordScore >= 0.6 && keywordScore > bestMatch.MatchScore {
			bestMatch = EventMatch{
				ManifestID:        manifestEvt.ID,
				ManifestLabel:     manifestEvt.Label,
				MatchedNodeID:     node.ID,
				MatchedNodeLabel:  node.Label,
				MatchMethod:       "keyword_overlap",
				MatchScore:        keywordScore,
				EntitiesMatched:   len(manifestEvt.Entities),
				EntitiesTotal:     len(manifestEvt.Entities),
				SourceMatch:       true,
				TemporalOverlap:   true,
				IsCorrect:         true,
				IsHallucination:   false,
			}
		}

		// Try substring containment (one label contains the other)
		if strings.Contains(normExtracted, normManifest) || strings.Contains(normManifest, normExtracted) {
			score := 0.85
			if score > bestMatch.MatchScore {
				bestMatch = EventMatch{
					ManifestID:        manifestEvt.ID,
					ManifestLabel:     manifestEvt.Label,
					MatchedNodeID:     node.ID,
					MatchedNodeLabel:  node.Label,
					MatchMethod:       "substring",
					MatchScore:        score,
					EntitiesMatched:   len(manifestEvt.Entities),
					EntitiesTotal:     len(manifestEvt.Entities),
					SourceMatch:       true,
					TemporalOverlap:   true,
					IsCorrect:         true,
					IsHallucination:   false,
				}
			}
		}
	}

	return bestMatch
}

// normalizeText normalizes text for comparison: lowercase, trim, remove titles
func normalizeText(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Trim whitespace
	s = strings.TrimSpace(s)

	// Remove common titles (Swedish and English)
	titles := []string{
		"mr.", "mrs.", "ms.", "dr.", "prof.", "herr", "fru", "fröken", "doktor",
		"ordföranden", "ordförande", "kassören", "kassör", "sekreteraren", "sekreterare",
	}
	for _, title := range titles {
		if strings.HasPrefix(s, title+" ") {
			s = strings.TrimPrefix(s, title+" ")
		}
	}

	// Remove punctuation (keep spaces and alphanumeric)
	var result []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			result = append(result, r)
		}
	}

	// Collapse multiple spaces
	return strings.Join(strings.Fields(string(result)), " ")
}

// levenshteinDistance computes the edit distance between two strings
func levenshteinDistance(a, b string) int {
	a = normalizeText(a)
	b = normalizeText(b)

	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Use a single row to save space
	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)

	for i := 0; i <= len(b); i++ {
		previous[i] = i
	}

	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			current[j] = min(
				previous[j]+1,      // deletion
				current[j-1]+1,     // insertion
				previous[j-1]+cost, // substitution
			)
		}
		previous, current = current, previous
	}

	return previous[len(b)]
}

// keywordOverlapScore computes the fraction of significant words in label A
// that also appear in label B. Stop words are excluded.
func keywordOverlapScore(a, b string) float64 {
	stopWords := map[string]bool{
		"i": true, "och": true, "av": true, "för": true, "till": true,
		"med": true, "på": true, "om": true, "den": true, "det": true,
		"en": true, "ett": true, "att": true, "ska": true, "har": true,
		"är": true, "var": true, "inte": true, "från": true, "som": true,
		"vid": true, "utan": true, "efter": true, "under": true,
		"the": true, "a": true, "an": true, "in": true, "of": true,
		"to": true, "for": true, "and": true, "with": true, "by": true,
		"kr": true, "exkl": true, "inkl": true, "moms": true,
	}

	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)

	// Build set of significant words in B
	setB := make(map[string]bool)
	for _, w := range wordsB {
		if !stopWords[w] && len(w) > 1 {
			setB[w] = true
		}
	}

	// Count significant words in A that appear in B
	significantA := 0
	matched := 0
	for _, w := range wordsA {
		if !stopWords[w] && len(w) > 1 {
			significantA++
			if setB[w] {
				matched++
			}
		}
	}

	if significantA == 0 {
		return 0.0
	}

	return float64(matched) / float64(significantA)
}

func min(a, b, c int) int {
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
