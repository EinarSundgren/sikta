package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/einarsundgren/sikta/internal/evaluation"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: compare-events <extraction-result.json> <manifest.json>")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  compare-events results/brf-v1-fixed.json corpora/brf/manifest.json")
		os.Exit(1)
	}

	resultPath := os.Args[1]
	manifestPath := os.Args[2]

	// Load extraction result
	resultData, err := os.ReadFile(resultPath)
	if err != nil {
		log.Fatalf("Failed to read result file: %v", err)
	}

	var resultExt evaluation.ExtractionResult
	if err := json.Unmarshal(resultData, &resultExt); err != nil {
		log.Fatalf("Failed to parse result JSON: %v", err)
	}

	// Flatten to Extraction
	extraction := resultExt.Flatten()

	// Load manifest
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Fatalf("Failed to read manifest file: %v", err)
	}

	var manifest evaluation.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		log.Fatalf("Failed to parse manifest JSON: %v", err)
	}

	// Run matchers
	entityMatcher := evaluation.NewEntityMatcher(&manifest, extraction)
	entityMatches := entityMatcher.MatchEntities()

	eventMatcher := evaluation.NewEventMatcher(&manifest, extraction, entityMatches)
	eventMatches := eventMatcher.MatchEvents(extraction)

	// Print verbose comparison
	printVerboseComparison(&manifest, extraction, eventMatches, resultPath, manifestPath)
}

func printVerboseComparison(manifest *evaluation.Manifest, extraction *evaluation.Extraction, eventMatches []evaluation.EventMatch, resultPath, manifestPath string) {
	fmt.Println("╔════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    EVENT COMPARISON: MANIFEST vs EXTRACTION                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Print file paths
	fmt.Printf("Extraction Result: %s\n", resultPath)
	fmt.Printf("Manifest:          %s\n", manifestPath)
	fmt.Printf("Corpus:             %s\n", manifest.Corpus)
	fmt.Printf("Prompt Version:     %s\n\n", extraction.PromptVersion)

	// Collect all extracted events
	extractedEvents := make([]evaluation.ExtractedNode, 0)
	for _, node := range extraction.Nodes {
		if node.NodeType == "event" {
			extractedEvents = append(extractedEvents, node)
		}
	}

	// Print summary statistics
	printSummaryStatistics(manifest, extractedEvents, eventMatches)

	// Print detailed match table
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println("MANIFEST EVENTS (Ground Truth)")
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()

	for _, match := range eventMatches {
		if match.ManifestID == "" {
			continue // Skip hallucinations (shown later)
		}

		fmt.Printf("┌─ Manifest Event: %s\n", match.ManifestID)
		fmt.Printf("│  Label:          %s\n", match.ManifestLabel)

		// Find manifest event details
		var manifestEvent *evaluation.ManifestEvent
		for i := range manifest.Events {
			if manifest.Events[i].ID == match.ManifestID {
				manifestEvent = &manifest.Events[i]
				break
			}
		}

		if manifestEvent != nil {
			fmt.Printf("│  Type:           %s\n", manifestEvent.Type)
			fmt.Printf("│  Source Doc:     %s\n", manifestEvent.SourceDoc)
			if manifestEvent.ClaimedTimeText != "" {
				fmt.Printf("│  Time:           %s\n", manifestEvent.ClaimedTimeText)
			}
			if len(manifestEvent.Entities) > 0 {
				fmt.Printf("│  Entities:       %s\n", strings.Join(manifestEvent.Entities, ", "))
			}
		}

		if match.IsCorrect {
			// MATCHED
			fmt.Printf("│  ✅ MATCHED\n")
			fmt.Printf("│     Extracted Label:  %s\n", match.MatchedNodeLabel)
			fmt.Printf("│     Match Method:     %s\n", match.MatchMethod)
			fmt.Printf("│     Match Score:      %.2f\n", match.MatchScore)
			if match.EntitiesMatched > 0 {
				fmt.Printf("│     Entities Matched: %d/%d\n", match.EntitiesMatched, match.EntitiesTotal)
			}
		} else {
			// NOT MATCHED
			fmt.Printf("│  ❌ NOT MATCHED\n")
			fmt.Printf("│     Best Candidate:   %s\n", match.MatchedNodeLabel)
			if match.MatchedNodeLabel != "" {
				fmt.Printf("│     Match Method:     %s\n", match.MatchMethod)
				fmt.Printf("│     Match Score:      %.2f\n", match.MatchScore)
			} else {
				fmt.Printf("│     No close match found\n")
			}
		}
		fmt.Printf("└───────────────────────────────────────────────────────────────────────────────\n")
		fmt.Println()
	}

	// Print hallucinated events (extracted but not in manifest)
	hallucinations := getHallucinatedEvents(eventMatches)
	if len(hallucinations) > 0 {
		fmt.Println("════════════════════════════════════════════════════════════════════════════════")
		fmt.Println("HALLUCINATED EVENTS (Extracted but not in Manifest)")
		fmt.Println("════════════════════════════════════════════════════════════════════════════════")
		fmt.Println()

		for _, match := range hallucinations {
			fmt.Printf("┌─ Extracted Event: %s\n", match.MatchedNodeID)
			fmt.Printf("│  Label:           %s\n", match.MatchedNodeLabel)

			// Find the extracted node for more details
			var extractedNode *evaluation.ExtractedNode
			for i := range extractedEvents {
				if extractedEvents[i].ID == match.MatchedNodeID {
					extractedNode = &extractedEvents[i]
					break
				}
			}

			if extractedNode != nil {
				if extractedNode.Excerpt != "" {
					excerpt := extractedNode.Excerpt
					if len(excerpt) > 100 {
						excerpt = excerpt[:100] + "..."
					}
					fmt.Printf("│  Excerpt:         %s\n", excerpt)
				}
				if extractedNode.ClaimedTimeText != "" {
					fmt.Printf("│  Time:            %s\n", extractedNode.ClaimedTimeText)
				}
				if extractedNode.Confidence > 0 {
					fmt.Printf("│  Confidence:      %.2f\n", extractedNode.Confidence)
				}
			}

			fmt.Printf("│  ⚠️  NO MATCH IN MANIFEST\n")
			fmt.Printf("└───────────────────────────────────────────────────────────────────────────────\n")
			fmt.Println()
		}
	}

	// Print all extracted events for reference
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println("ALL EXTRACTED EVENTS (Reference)")
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()

	// Sort by document then label
	sort.Slice(extractedEvents, func(i, j int) bool {
		// Extract document ID from node ID or properties
		docI := getDocID(extractedEvents[i])
		docJ := getDocID(extractedEvents[j])
		if docI != docJ {
			return docI < docJ
		}
		return extractedEvents[i].Label < extractedEvents[j].Label
	})

	currentDoc := ""
	for _, node := range extractedEvents {
		docID := getDocID(node)
		if docID != currentDoc {
			if currentDoc != "" {
				fmt.Println()
			}
			fmt.Printf("Document: %s\n", docID)
			fmt.Println("────────────────────────────────────────────────────────────────────────────────")
			currentDoc = docID
		}

		fmt.Printf("  [%s] %s\n", node.ID, node.Label)
		if node.ClaimedTimeText != "" {
			fmt.Printf("       Time: %s\n", node.ClaimedTimeText)
		}
		if node.Confidence > 0 && node.Confidence < 1.0 {
			fmt.Printf("       Confidence: %.2f\n", node.Confidence)
		}
	}
	fmt.Println()
}

func printSummaryStatistics(manifest *evaluation.Manifest, extractedEvents []evaluation.ExtractedNode, eventMatches []evaluation.EventMatch) {
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println("SUMMARY STATISTICS")
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()

	// Count matches
	matchedCount := 0
	for _, match := range eventMatches {
		if match.IsCorrect && match.ManifestID != "" {
			matchedCount++
		}
	}

	recall := float64(matchedCount) / float64(len(manifest.Events)) * 100

	fmt.Printf("Manifest Events (Expected):     %d\n", len(manifest.Events))
	fmt.Printf("Extracted Events (Found):        %d\n", len(extractedEvents))
	fmt.Printf("Matched Events (Recall):         %d / %d (%.1f%%)\n", matchedCount, len(manifest.Events), recall)
	fmt.Printf("\n")

	// Count by match method
	matchMethods := make(map[string]int)
	for _, match := range eventMatches {
		if match.IsCorrect && match.ManifestID != "" {
			matchMethods[match.MatchMethod]++
		}
	}

	if len(matchMethods) > 0 {
		fmt.Println("Match Methods:")
		for method, count := range matchMethods {
			fmt.Printf("  %s: %d\n", method, count)
		}
		fmt.Println()
	}

	// Count hallucinations
	hallucinationCount := 0
	for _, match := range eventMatches {
		if match.IsHallucination {
			hallucinationCount++
		}
	}

	if hallucinationCount > 0 {
		fmt.Printf("Hallucinated Events:            %d (extracted but not in manifest)\n", hallucinationCount)
		fmt.Println()
	}

	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()
}

func getHallucinatedEvents(eventMatches []evaluation.EventMatch) []evaluation.EventMatch {
	hallucinations := make([]evaluation.EventMatch, 0)
	for _, match := range eventMatches {
		if match.IsHallucination {
			hallucinations = append(hallucinations, match)
		}
	}
	return hallucinations
}

func getDocID(node evaluation.ExtractedNode) string {
	// Try to get document ID from properties
	if docID, ok := node.Properties["source_doc"].(string); ok {
		return docID
	}
	// Fallback: extract from node ID if it contains a hyphen
	if idx := strings.Index(node.ID, "-"); idx > 0 {
		return node.ID[:idx]
	}
	return "unknown"
}
