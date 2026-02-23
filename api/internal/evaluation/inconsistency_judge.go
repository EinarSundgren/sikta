package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/einarsundgren/sikta/internal/extraction/claude"
)

// InconsistencyJudge uses an LLM to determine if detected inconsistencies
// match the ground truth inconsistencies from the manifest.
type InconsistencyJudge struct {
	client *claude.Client
	logger *slog.Logger
	model  string
}

// NewInconsistencyJudge creates a new LLM inconsistency judge.
func NewInconsistencyJudge(client *claude.Client, logger *slog.Logger, model string) *InconsistencyJudge {
	return &InconsistencyJudge{
		client: client,
		logger: logger,
		model:  model,
	}
}

// InconsistencyJudgeDecision represents the LLM's decision about an inconsistency match
type InconsistencyJudgeDecision struct {
	Match        bool    `json:"match"`
	MatchedID    string  `json:"matched_id"`
	Confidence   float64 `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
}

// JudgeInconsistencies takes manifest inconsistencies and detected inconsistencies,
// asks the LLM to find semantic matches. Returns updated InconsistencyMatch entries
// for any inconsistencies that the judge determines are matches.
func (j *InconsistencyJudge) JudgeInconsistencies(
	ctx context.Context,
	manifestInconsistencies []ManifestInconsistency,
	detectedInconsistencies []DetectedInconsistencyForJudge,
) ([]InconsistencyMatch, error) {
	if len(manifestInconsistencies) == 0 || len(detectedInconsistencies) == 0 {
		return nil, nil
	}

	var results []InconsistencyMatch

	for _, manifestInc := range manifestInconsistencies {
		decision, err := j.judgeSingleInconsistency(ctx, manifestInc, detectedInconsistencies)
		if err != nil {
			j.logger.Error("judge failed for inconsistency", "inconsistency_id", manifestInc.ID, "error", err)
			continue
		}

		if decision.Match && decision.MatchedID != "" {
			// Find the matched detected inconsistency
			var matchedInc *DetectedInconsistencyForJudge
			for i := range detectedInconsistencies {
				if detectedInconsistencies[i].ID == decision.MatchedID {
					matchedInc = &detectedInconsistencies[i]
					break
				}
			}

			if matchedInc != nil {
				results = append(results, InconsistencyMatch{
					ManifestID:       manifestInc.ID,
					ManifestType:     manifestInc.Type,
					ManifestSeverity: manifestInc.Severity,
					MatchedID:        matchedInc.ID,
					MatchedByLLM:     true,
					MatchConfidence:  decision.Confidence,
					Reasoning:        decision.Reasoning,
					IsCorrect:        true,
					IsHallucination:  false,
				})

				j.logger.Info("LLM judge matched inconsistency",
					"manifest_id", manifestInc.ID,
					"manifest_description", manifestInc.Description,
					"matched_id", matchedInc.ID,
					"matched_description", matchedInc.Description,
					"confidence", decision.Confidence,
					"reasoning", decision.Reasoning)
			}
		} else {
			j.logger.Debug("LLM judge: no match found for inconsistency",
				"manifest_id", manifestInc.ID,
				"manifest_description", manifestInc.Description,
				"reasoning", decision.Reasoning)
		}
	}

	return results, nil
}

// DetectedInconsistencyForJudge is a format suitable for the judge
// (matches the extraction output format)
type DetectedInconsistencyForJudge struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Severity         string                 `json:"severity"`
	Description      string                 `json:"description"`
	Documents        []string               `json:"documents"`
	EntitiesInvolved []string               `json:"entities_involved"`
	Evidence         map[string]interface{} `json:"evidence"`
	Confidence       float32                `json:"confidence"`
}

// judgeSingleInconsistency asks the LLM to judge a single manifest inconsistency against all candidates
func (j *InconsistencyJudge) judgeSingleInconsistency(
	ctx context.Context,
	manifestInc ManifestInconsistency,
	candidates []DetectedInconsistencyForJudge,
) (*InconsistencyJudgeDecision, error) {
	systemPrompt := j.buildSystemPrompt()
	userMessage := j.buildUserMessage(manifestInc, candidates)

	j.logger.Debug("calling LLM inconsistency judge",
		"manifest_id", manifestInc.ID,
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
	j.logger.Debug("LLM inconsistency judge response",
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

// buildSystemPrompt creates the judge system prompt for inconsistency matching
func (j *InconsistencyJudge) buildSystemPrompt() string {
	return `You are an evaluation judge comparing detected inconsistencies against ground truth.

Given a MANIFEST INCONSISTENCY (the expected detection) and a list of CANDIDATE inconsistencies
(actually detected by the system), determine if any candidate refers to the same underlying issue.

MATCH CRITERIA:
- Both must refer to the SAME underlying contradiction or violation
- Different wording and different severity levels are acceptable
- The same documents and entities should typically be involved
- The core issue must match (e.g., both about budget discrepancy, both about authority violation)
- Different aspects of the same problem may or may not match (use judgment)

INCONSISTENCY TYPES:
- amount: numeric discrepancies between documents
- temporal: impossible timelines, missed deadlines
- authority: actions without authorization
- procedural: process violations, role conflicts
- obligation: untracked or violated commitments
- provenance: citation issues, reference problems
- contradiction: direct factual conflicts

EXAMPLES OF MATCHES:
- Manifest: "Budget set to 650k in A1, reduced to 600k in A3 without justification"
  Candidate: "Budget discrepancy: 650,000 kr in A1 vs 600,000 kr in A3" → MATCH
  (Same budget discrepancy between same documents)

- Manifest: "Per Sandberg ordered work without board authorization"
  Candidate: "Authority violation: suppleant ordered additional work without approval" → MATCH
  (Same authority violation by same person)

EXAMPLES OF NON-MATCHES:
- Manifest: "Budget discrepancy"
  Candidate: "Deadline missed by 3 weeks" → NO MATCH
  (Completely different issues)

Respond with JSON only, no markdown:
{"match": boolean, "matched_id": "exact ID from candidates" or null, "confidence": 0.0-1.0, "reasoning": "brief explanation"}

If no candidate matches, return:
{"match": false, "matched_id": null, "confidence": 0.0, "reasoning": "why no match found"}`
}

// buildUserMessage creates the user message with manifest inconsistency and candidates
func (j *InconsistencyJudge) buildUserMessage(manifestInc ManifestInconsistency, candidates []DetectedInconsistencyForJudge) string {
	var sb strings.Builder

	sb.WriteString("MANIFEST INCONSISTENCY:\n")
	sb.WriteString(fmt.Sprintf("  ID: %s\n", manifestInc.ID))
	sb.WriteString(fmt.Sprintf("  Type: %s\n", manifestInc.Type))
	sb.WriteString(fmt.Sprintf("  Severity: %s\n", manifestInc.Severity))
	sb.WriteString(fmt.Sprintf("  Description: %s\n", manifestInc.Description))
	sb.WriteString(fmt.Sprintf("  Documents: %v\n", manifestInc.Documents))
	sb.WriteString(fmt.Sprintf("  Entities: %v\n", manifestInc.EntitiesInvolved))
	if manifestInc.Evidence != nil {
		sb.WriteString("  Evidence:\n")
		for key, val := range manifestInc.Evidence {
			sb.WriteString(fmt.Sprintf("    %s: %v\n", key, val))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("CANDIDATE DETECTED INCONSISTENCIES:\n")
	for i, candidate := range candidates {
		sb.WriteString(fmt.Sprintf("%d. ID: %s\n", i+1, candidate.ID))
		sb.WriteString(fmt.Sprintf("   Type: %s\n", candidate.Type))
		sb.WriteString(fmt.Sprintf("   Severity: %s\n", candidate.Severity))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", candidate.Description))
		sb.WriteString(fmt.Sprintf("   Documents: %v\n", candidate.Documents))
		sb.WriteString(fmt.Sprintf("   Entities: %v\n", candidate.EntitiesInvolved))
		if candidate.Evidence != nil {
			sb.WriteString("   Evidence:\n")
			for key, val := range candidate.Evidence {
				sb.WriteString(fmt.Sprintf("     %s: %v\n", key, val))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nDoes any candidate match the manifest inconsistency? Return JSON with matched_id being the exact ID from the candidate list.")

	return sb.String()
}

// parseJudgeResponse extracts the InconsistencyJudgeDecision from LLM response text
func (j *InconsistencyJudge) parseJudgeResponse(responseText string) (*InconsistencyJudgeDecision, error) {
	// Strip markdown code blocks if present
	text := stripMarkdownCodeBlocks(responseText)

	// Fix trailing commas
	text = fixTrailingCommas(text)

	// Try to parse as JSON directly
	var decision InconsistencyJudgeDecision
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
	lowerText := strings.ToLower(text)
	if strings.Contains(lowerText, "no match") ||
		strings.Contains(lowerText, "no candidate") ||
		strings.Contains(lowerText, "none of the") ||
		strings.Contains(lowerText, "does not match") {
		return &InconsistencyJudgeDecision{
			Match:      false,
			MatchedID:  "",
			Confidence: 0.0,
			Reasoning:  "LLM indicated no match in natural language",
		}, nil
	}

	// Last resort: look for match:true pattern
	if strings.Contains(text, "match") {
		isMatch := strings.Contains(text, "true") && !strings.Contains(text, "false")

		matchedID := ""
		reasoning := ""

		// Look for matched_id pattern
		idPatterns := []string{"matched_id", "matched id", "id"}
		for _, pattern := range idPatterns {
			if idx := strings.Index(strings.ToLower(text), pattern); idx >= 0 {
				sub := text[idx:]
				if quoteStart := strings.Index(sub, "\""); quoteStart >= 0 && quoteStart < 50 {
					quoteEnd := strings.Index(sub[quoteStart+1:], "\"")
					if quoteEnd >= 0 && quoteEnd < 100 {
						matchedID = sub[quoteStart+1 : quoteStart+1+quoteEnd]
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

		if isMatch && matchedID != "" {
			return &InconsistencyJudgeDecision{
				Match:      true,
				MatchedID:  matchedID,
				Confidence: 0.8,
				Reasoning:  reasoning,
			}, nil
		}

		return &InconsistencyJudgeDecision{
			Match:      false,
			MatchedID:  "",
			Confidence: 0.0,
			Reasoning:  "Parsed from natural language: " + text[:minInt(200, len(text))],
		}, nil
	}

	return nil, fmt.Errorf("no JSON found in response: %s", text[:minInt(100, len(text))])
}
