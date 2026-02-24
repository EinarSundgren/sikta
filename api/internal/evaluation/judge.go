package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/einarsundgren/sikta/internal/extraction/claude"
)

// EventJudge uses an LLM to determine if unmatched manifest events
// have semantically equivalent extractions that deterministic matching missed.
type EventJudge struct {
	client *claude.Client
	logger *slog.Logger
	model  string
}

// NewEventJudge creates a new LLM event judge.
func NewEventJudge(client *claude.Client, logger *slog.Logger, model string) *EventJudge {
	return &EventJudge{
		client: client,
		logger: logger,
		model:  model,
	}
}

// JudgeDecision represents the LLM's decision about an event match
type JudgeDecision struct {
	Match        bool    `json:"match"`
	MatchedLabel string  `json:"matched_label"`
	Confidence   float64 `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
}

// JudgeUnmatchedEvents takes unmatched manifest events and unmatched extracted events,
// asks the LLM to find semantic matches. Returns updated EventMatch entries for
// any events that the judge determines are matches.
func (j *EventJudge) JudgeUnmatchedEvents(
	ctx context.Context,
	unmatchedManifest []ManifestEvent,
	unmatchedExtracted []ExtractedNode,
) ([]EventMatch, error) {
	if len(unmatchedManifest) == 0 || len(unmatchedExtracted) == 0 {
		return nil, nil
	}

	var results []EventMatch

	for _, manifestEvt := range unmatchedManifest {
		decision, err := j.judgeSingleEvent(ctx, manifestEvt, unmatchedExtracted)
		if err != nil {
			j.logger.Error("judge failed for event", "event_id", manifestEvt.ID, "error", err)
			continue
		}

		if decision.Match && decision.MatchedLabel != "" {
			// Find the matched node to get its ID
			var matchedNode *ExtractedNode
			for i := range unmatchedExtracted {
				if unmatchedExtracted[i].Label == decision.MatchedLabel {
					matchedNode = &unmatchedExtracted[i]
					break
				}
			}

			if matchedNode != nil {
				results = append(results, EventMatch{
					ManifestID:        manifestEvt.ID,
					ManifestLabel:     manifestEvt.Label,
					MatchedNodeID:     matchedNode.ID,
					MatchedNodeLabel:  matchedNode.Label,
					MatchMethod:       "llm_judge",
					MatchScore:        decision.Confidence,
					EntitiesMatched:   len(manifestEvt.Entities),
					EntitiesTotal:     len(manifestEvt.Entities),
					SourceMatch:       true,
					TemporalOverlap:   true,
					IsCorrect:         true,
					IsHallucination:   false,
				})

				j.logger.Info("LLM judge matched event",
					"manifest_id", manifestEvt.ID,
					"manifest_label", manifestEvt.Label,
					"matched_label", matchedNode.Label,
					"confidence", decision.Confidence,
					"reasoning", decision.Reasoning)
			}
		} else {
			j.logger.Debug("LLM judge: no match found",
				"manifest_id", manifestEvt.ID,
				"manifest_label", manifestEvt.Label,
				"reasoning", decision.Reasoning)
		}
	}

	return results, nil
}

// judgeSingleEvent asks the LLM to judge a single manifest event against all candidates
func (j *EventJudge) judgeSingleEvent(
	ctx context.Context,
	manifestEvt ManifestEvent,
	candidates []ExtractedNode,
) (*JudgeDecision, error) {
	systemPrompt := j.buildSystemPrompt()
	userMessage := j.buildUserMessage(manifestEvt, candidates)

	j.logger.Debug("calling LLM judge",
		"manifest_id", manifestEvt.ID,
		"candidates", len(candidates),
		"model", j.model)

	resp, err := j.client.SendSystemPrompt(ctx, systemPrompt, userMessage, j.model)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}

	responseText := resp.Content[0].Text
	j.logger.Debug("LLM judge response",
		"input_tokens", resp.Usage.InputTokens,
		"output_tokens", resp.Usage.OutputTokens,
		"response_length", len(responseText))

	// Parse JSON response
	decision, err := j.parseJudgeResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse judge response: %w", err)
	}

	return decision, nil
}

// buildSystemPrompt creates the judge system prompt
func (j *EventJudge) buildSystemPrompt() string {
	return `You are an evaluation judge comparing extracted events against ground truth.

Given a MANIFEST EVENT (the expected extraction) and a list of CANDIDATE events
(actually extracted from the document), determine if any candidate refers to the
same real-world occurrence as the manifest event.

MATCH CRITERIA:
- Both must refer to the SAME specific occurrence (same who/what/when)
- Different wording is acceptable if the underlying event is identical
- The same temporal scope is required (same date/period or overlapping time)
- Partial overlaps are NOT matches (e.g., "construction started" ≠ "construction completed")
- Different aspects of the same situation are NOT matches (e.g., "inspection done" ≠ "damage found during inspection")

EXAMPLES OF MATCHES:
- "Fuktsinspektion utförd" = "Fasadsinspektion med fuktskador konstaterade" (same inspection event)
- "Byggstart fasadrenovering" = "Beräknad byggstart v.32" (same construction start, different specificity)
- "Val av entreprenör — NorrBygg Fasad AB antagen" = "Beslut att anta offert från NorrBygg Fasad AB" (same decision)

EXAMPLES OF NON-MATCHES:
- "Budget fastställd" ≠ "Upphandling beslutad" (different events)
- "Inspektion utförd" ≠ "Fuktskador konstaterade" (inspection vs. finding from inspection)

Respond with JSON only, no markdown:
{"match": boolean, "matched_label": "exact label from candidates" or null, "confidence": 0.0-1.0, "reasoning": "brief explanation"}

If no candidate matches, return:
{"match": false, "matched_label": null, "confidence": 0.0, "reasoning": "why no match found"}`
}

// buildUserMessage creates the user message with manifest event and candidates
func (j *EventJudge) buildUserMessage(manifestEvt ManifestEvent, candidates []ExtractedNode) string {
	var sb strings.Builder

	sb.WriteString("MANIFEST EVENT:\n")
	sb.WriteString(fmt.Sprintf("  Label: %s\n", manifestEvt.Label))
	sb.WriteString(fmt.Sprintf("  Type: %s\n", manifestEvt.Type))
	if manifestEvt.ClaimedTimeText != "" {
		sb.WriteString(fmt.Sprintf("  Time: %s\n", manifestEvt.ClaimedTimeText))
	}
	if manifestEvt.SourceDoc != "" {
		sb.WriteString(fmt.Sprintf("  Source: %s", manifestEvt.SourceDoc))
		if manifestEvt.SourceSection != "" {
			sb.WriteString(fmt.Sprintf(" %s", manifestEvt.SourceSection))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("CANDIDATE EXTRACTED EVENTS:\n")
	for i, candidate := range candidates {
		sb.WriteString(fmt.Sprintf("%d. Label: %s\n", i+1, candidate.Label))
		if candidate.ClaimedTimeText != "" {
			sb.WriteString(fmt.Sprintf("   Time: %s\n", candidate.ClaimedTimeText))
		}
		if candidate.Excerpt != "" {
			// Truncate long excerpts
			excerpt := candidate.Excerpt
			if len(excerpt) > 200 {
				excerpt = excerpt[:197] + "..."
			}
			sb.WriteString(fmt.Sprintf("   Excerpt: %s\n", excerpt))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nDoes any candidate match the manifest event? Return JSON.")

	return sb.String()
}

// parseJudgeResponse extracts the JudgeDecision from LLM response text
func (j *EventJudge) parseJudgeResponse(responseText string) (*JudgeDecision, error) {
	// Strip markdown code blocks if present
	text := stripMarkdownCodeBlocks(responseText)

	// Fix trailing commas
	text = fixTrailingCommas(text)

	// Try to parse as JSON directly
	var decision JudgeDecision
	if err := json.Unmarshal([]byte(text), &decision); err == nil {
		return &decision, nil
	}

	// Try to extract JSON from the response
	jsonStart := strings.Index(text, "{")
	jsonEnd := strings.LastIndex(text, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := text[jsonStart : jsonEnd+1]
		// Fix any remaining issues
		jsonStr = fixTrailingCommas(jsonStr)
		if err := json.Unmarshal([]byte(jsonStr), &decision); err == nil {
			return &decision, nil
		}
	}

	// Fallback: parse natural language response
	// If the response clearly indicates no match
	lowerText := strings.ToLower(text)
	if strings.Contains(lowerText, "no match") ||
		strings.Contains(lowerText, "no candidate") ||
		strings.Contains(lowerText, "none of the") ||
		strings.Contains(lowerText, "does not match") {
		return &JudgeDecision{
			Match:        false,
			MatchedLabel: "",
			Confidence:   0.0,
			Reasoning:    "LLM indicated no match in natural language",
		}, nil
	}

	// Last resort: look for match:true or match:false pattern
	if strings.Contains(text, "match") {
		// Try to extract match boolean
		isMatch := strings.Contains(text, "true") && !strings.Contains(text, "false")

		// Try to extract matched_label by looking for quoted strings after "label" or similar
		matchedLabel := ""
		reasoning := ""

		// Look for patterns like "matched_label": "..." or label: "..."
		labelPatterns := []string{"matched_label", "matched label", "label"}
		for _, pattern := range labelPatterns {
			if idx := strings.Index(strings.ToLower(text), pattern); idx >= 0 {
				sub := text[idx:]
				if quoteStart := strings.Index(sub, "\""); quoteStart >= 0 && quoteStart < 50 {
					quoteEnd := strings.Index(sub[quoteStart+1:], "\"")
					if quoteEnd >= 0 && quoteEnd < 200 {
						matchedLabel = sub[quoteStart+1 : quoteStart+1+quoteEnd]
						break
					}
				}
			}
		}

		// Look for reasoning
		if idx := strings.Index(strings.ToLower(text), "reasoning"); idx >= 0 {
			sub := text[idx:]
			if quoteStart := strings.Index(sub, "\""); quoteStart >= 0 && quoteStart < 20 {
				quoteEnd := strings.Index(sub[quoteStart+1:], "\"")
				if quoteEnd >= 0 {
					reasoning = sub[quoteStart+1 : quoteStart+1+quoteEnd]
				}
			}
		}

		if isMatch && matchedLabel != "" {
			return &JudgeDecision{
				Match:        true,
				MatchedLabel: matchedLabel,
				Confidence:   0.8,
				Reasoning:    reasoning,
			}, nil
		}

		return &JudgeDecision{
			Match:        false,
			MatchedLabel: "",
			Confidence:   0.0,
			Reasoning:    "Parsed from natural language: " + text[:minInt(200, len(text))],
		}, nil
	}

	return nil, fmt.Errorf("no JSON found in response: %s", text[:minInt(100, len(text))])
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// stripMarkdownCodeBlocks removes ```json and ``` markers
func stripMarkdownCodeBlocks(s string) string {
	// Remove ```json or ``` at the start
	s = regexp.MustCompile("(?m)```(?:json)?\\s*").ReplaceAllString(s, "")
	// Remove trailing ```
	s = regexp.MustCompile("(?m)\\s*```\\s*$").ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// fixTrailingCommas removes trailing commas before ] and }
func fixTrailingCommas(s string) string {
	result := regexp.MustCompile(`,\s*([}\]])`).ReplaceAllString(s, "$1")
	return result
}
