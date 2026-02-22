package extraction

import (
	"strings"

	"github.com/google/uuid"
)

// ExtractedNode represents a node extracted from text by the LLM
type ExtractedNode struct {
	ID               string                 `json:"id"`
	NodeType         string                 `json:"node_type"`
	Label            string                 `json:"label"`
	Properties       map[string]interface{} `json:"properties"`
	Confidence       float32                `json:"confidence"`
	Modality         string                 `json:"modality,omitempty"`     // asserted, hypothetical, etc.
	Excerpt          string                 `json:"excerpt,omitempty"`      // source text excerpt
	ClaimedTimeStart string                 `json:"claimed_time_start,omitempty"` // Flexible date format (YYYY-MM-DD or RFC3339)
	ClaimedTimeEnd   string                 `json:"claimed_time_end,omitempty"`   // Flexible date format (YYYY-MM-DD or RFC3339)
	ClaimedTimeText  string                 `json:"claimed_time_text,omitempty"`
	ClaimedGeoRegion string                 `json:"claimed_geo_region,omitempty"`
	ClaimedGeoText   string                 `json:"claimed_geo_text,omitempty"`
}

// ExtractedEdge represents an edge extracted from text by the LLM
type ExtractedEdge struct {
	ID         string                 `json:"id"`
	EdgeType   string                 `json:"edge_type"`
	SourceNode string                 `json:"source_node"` // entity label
	TargetNode string                 `json:"target_node"` // entity label
	Properties map[string]interface{} `json:"properties"`
	IsNegated  bool                   `json:"is_negated"`
	Confidence float32                `json:"confidence"`
	Modality   string                 `json:"modality,omitempty"` // asserted, hypothetical, etc.
	Excerpt    string                 `json:"excerpt,omitempty"`  // source text excerpt
}

// GraphExtractionResponse is the response from the LLM for graph extraction
type GraphExtractionResponse struct {
	Nodes []ExtractedNode `json:"nodes"`
	Edges []ExtractedEdge `json:"edges"`
}

// Helper functions

// parseUUID parses a UUID from string
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// stripMarkdownCodeBlocks removes markdown code block markers from JSON responses
func stripMarkdownCodeBlocks(s string) string {
	// Remove opening ```json and closing ```
	start := 0
	end := len(s)

	// Check for ```json at start
	if len(s) >= 7 && s[0:7] == "```json" {
		start = 7
	} else if len(s) >= 3 && s[0:3] == "```" {
		start = 3
	}

	// Check for ``` at end
	if len(s) >= 3 && s[len(s)-3:] == "```" {
		end = len(s) - 3
	}

	// Trim whitespace
	result := strings.TrimSpace(s[start:end])
	return result
}

// extractJSONFromMarkdown attempts to extract JSON from markdown code blocks.
func extractJSONFromMarkdown(text string) string {
	// First try stripMarkdownCodeBlocks
	stripped := stripMarkdownCodeBlocks(text)
	if stripped != "" && stripped != text {
		return stripped
	}

	// Look for JSON object boundaries
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return ""
	}

	// Find first { and last }
	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return ""
	}

	return text[startIdx : endIdx+1]
}
