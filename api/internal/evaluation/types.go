package evaluation

import (
	"time"
)

// ScoreResult contains all scoring metrics for an extraction run
type ScoreResult struct {
	Corpus       string            // Corpus identifier (e.g., "brf", "mna", "police")
	PromptVersion string           // Prompt version (e.g., "v1")
	Timestamp    time.Time         // When the scoring was run

	// Entity scores
	EntityRecall    float64       // found / expected
	EntityPrecision float64       // correct / found
	EntityF1       float64       // harmonic mean of recall and precision
	EntityDetails   []EntityMatch // Per-entity breakdown

	// Event scores
	EventRecall    float64      // found / expected
	EventPrecision float64      // correct / found
	EventF1       float64      // harmonic mean of recall and precision
	EventDetails   []EventMatch // Per-event breakdown

	// Inconsistency scores (only in --full mode with LLM-as-judge)
	InconsistencyRecall    float64               // found / expected
	InconsistencyPrecision float64               // correct / found
	InconsistencyDetails   []InconsistencyMatch  // Per-inconsistency breakdown

	// Quality metrics
	FalsePositiveRate     float64 // hallucinations / total extracted
	AvgConfidenceAccuracy float64 // how well confidence scores predict correctness
}

// EntityMatch tracks matching status for a single entity
type EntityMatch struct {
	ManifestID      string   // ID from manifest (e.g., "E1")
	ManifestLabel   string   // Label from manifest (e.g., "Anna Lindqvist")
	MatchedNodeID   string   // ID of matched extracted node (empty if no match)
	MatchedNodeLabel string // Label of matched extracted node (empty if no match)
	MatchMethod     string   // How the match was found: "exact", "levenshtein", "substring", "none"
	MatchScore      float64  // Confidence in this match (1.0 = exact, 0.0 = no match)
	IsCorrect       bool     // True if this entity was correctly extracted
	IsHallucination bool     // True if an extracted entity matches nothing in manifest
}

// EventMatch tracks matching status for a single event
type EventMatch struct {
	ManifestID       string   // ID from manifest (e.g., "V1")
	ManifestLabel    string   // Label from manifest (e.g., "Fuktsinspektion")
	MatchedNodeID    string   // ID of matched extracted node (empty if no match)
	MatchedNodeLabel string   // Label of matched extracted node (empty if no match)
	MatchMethod      string   // How the match was found
	MatchScore       float64  // Confidence in this match
	EntitiesMatched  int      // Number of entities involved in both manifest and extraction
	EntitiesTotal    int      // Total entities involved in manifest event
	SourceMatch      bool     // True if source document matches
	TemporalOverlap  bool     // True if temporal claims overlap
	IsCorrect        bool     // True if event was correctly extracted
	IsHallucination  bool     // True if extracted event matches nothing in manifest
}

// InconsistencyMatch tracks matching status for a single inconsistency
type InconsistencyMatch struct {
	ManifestID       string   // ID from manifest (e.g., "I1")
	ManifestType     string   // Type from manifest (e.g., "amount", "temporal")
	ManifestSeverity string   // Severity from manifest (e.g., "high", "medium")
	MatchedID        string   // ID of matched detected inconsistency (empty if no match)
	MatchedByLLM      bool     // True if matched via LLM-as-judge
	MatchConfidence  float64  // LLM confidence in the match (0.0-1.0)
	Reasoning        string   // LLM reasoning for the match (if applicable)
	IsCorrect        bool     // True if inconsistency was correctly detected
	IsHallucination  bool     // True if detected inconsistency doesn't match any manifest inconsistency
}

// DiffResult compares two ScoreResult structs
type DiffResult struct {
	Corpus         string
	VersionA       string
	VersionB       string
	Timestamp      time.Time

	EntityRecallDelta    float64
	EntityPrecisionDelta float64
	EntityF1Delta        float64

	EventRecallDelta    float64
	EventPrecisionDelta float64
	EventF1Delta        float64

	ImprovedEntities []string // Entity IDs that improved from A to B
	RegressedEntities []string // Entity IDs that regressed from A to B

	ImprovedEvents []string // Event IDs that improved from A to B
	RegressedEvents []string // Event IDs that regressed from A to B
}

// Manifest represents the ground truth for a corpus
type Manifest struct {
	Corpus      string          // Corpus identifier
	Description string          // Human-readable description
	Language    string          // "en", "sv", etc.
	Domain      string          // "brf_protocol", "due_diligence", etc.
	Documents   []ManifestDoc   // Documents in the corpus
	Entities    []ManifestEntity // Expected entities
	Events      []ManifestEvent  // Expected events
	Inconsistencies []ManifestInconsistency // Expected inconsistencies
}

// ManifestDoc represents a document in the manifest
type ManifestDoc struct {
	ID     string   // Document ID (e.g., "A1", "B1")
	Filename string  // Original filename
	Type   string   // Document type (protocol, invoice, etc.)
	Trust  float64  // Source trust score (0.0-1.0)
	Date   string   // Document date (ISO format if available)
}

// ManifestEntity represents an expected entity extraction
type ManifestEntity struct {
	ID          string                 `json:"id"`          // Entity ID (e.g., "E1")
	Label       string                 `json:"label"`       // Entity label (e.g., "Anna Lindqvist")
	Type        string                 `json:"type"`        // Entity type (person, organization, etc.)
	Aliases     []string               `json:"aliases"`     // Alternative names/references
	Properties  map[string]interface{} `json:"properties"`  // Additional properties
	MentionedIn []string               `json:"mentioned_in"` // Document IDs where this entity appears
}

// ManifestEvent represents an expected event extraction
type ManifestEvent struct {
	ID                string   `json:"id"`                // Event ID (e.g., "V1")
	Label             string   `json:"label"`             // Event label (e.g., "Fuktsinspektion")
	Type              string   `json:"type"`              // Event type (decision, inspection, etc.)
	ClaimedTimeText   string   `json:"claimed_time_text"`   // Raw time text from document
	ClaimedTimeStart  string   `json:"claimed_time_start"`  // ISO date if available
	ClaimedTimeEnd    string   `json:"claimed_time_end"`    // ISO date if available
	Entities          []string `json:"entities"`          // Entity IDs involved in this event
	SourceDoc         string   `json:"source_doc"`         // Document ID where this event appears
	SourceSection     string   `json:"source_section"`     // Section reference (e.g., "§5")
}

// ManifestInconsistency represents an expected inconsistency
type ManifestInconsistency struct {
	ID               string                 // Inconsistency ID (e.g., "I1")
	Type             string                 // Inconsistency type (amount, temporal, etc.)
	Severity         string                 // Severity (high, medium, low)
	Description      string                 // Human-readable description
	Documents        []string               // Document IDs involved
	EntitiesInvolved []string               // Entity IDs involved
	Evidence         map[string]interface{} // Evidence from each side (flexible for nested objects)
}

// ExtractionResult represents the raw output from sikta-eval (with Documents array)
type ExtractionResult struct {
	Corpus          string                   // Corpus identifier
	PromptVersion   string                   // Prompt version used
	Documents       []DocumentExtraction     // Document-level extractions
	Inconsistencies []ExtractedInconsistency // Cross-document inconsistencies (optional)
	Metadata        ExtractionMetadata       // Metadata
}

// DocumentExtraction represents nodes/edges extracted from a single document
type DocumentExtraction struct {
	DocumentID string          // Document ID (e.g., "A1")
	Filename   string          // Original filename
	Nodes      []ExtractedNode // Nodes from this document
	Edges      []ExtractedEdge // Edges from this document
	Error      string          // Error message if extraction failed
}

// ExtractionMetadata contains metadata about the extraction run
type ExtractionMetadata struct {
	Model       string // LLM model used
	Timestamp   string // ISO timestamp
	TotalDocs   int    // Total documents processed
	TotalNodes  int    // Total nodes extracted
	TotalEdges  int    // Total edges extracted
	FailedDocs  int    // Number of failed documents
}

// Extraction represents a flattened extraction for scoring
type Extraction struct {
	Corpus         string                   `json:"corpus"`          // Corpus identifier
	PromptVersion  string                   `json:"prompt_version"`  // Prompt version used
	Nodes          []ExtractedNode          `json:"nodes"`           // Extracted nodes (flattened from all documents)
	Edges          []ExtractedEdge          `json:"edges"`           // Extracted edges (flattened from all documents)
	Inconsistencies []ExtractedInconsistency `json:"inconsistencies"` // Detected inconsistencies (cross-document)
	Timestamp      time.Time                `json:"timestamp"`       // When extraction was run
}

// ExtractedNode represents a node from extraction output
type ExtractedNode struct {
	ID              string                 `json:"id"`              // Node ID
	NodeType        string                 `json:"node_type"`      // Node type
	Label           string                 `json:"label"`           // Node label
	Properties      map[string]interface{} `json:"properties"`    // Node properties
	Confidence      float32                `json:"confidence"`      // Confidence score
	Modality        string                 `json:"modality"`        // Modality (asserted, etc.)
	Excerpt         string                 `json:"excerpt"`         // Source excerpt
	ClaimedTimeText string                 `json:"claimed_time_text"` // Raw time text
	ClaimedGeoText  string                 `json:"claimed_geo_text"`  // Raw location text
}

// ExtractedEdge represents an edge from extraction output
type ExtractedEdge struct {
	ID         string                 // Edge ID
	EdgeType   string                 // Edge type
	SourceNode string                 // Source node label
	TargetNode string                 // Target node label
	Properties map[string]interface{} // Edge properties
	Confidence float32                // Confidence score
	Modality   string                 // Modality
	Excerpt    string                 // Source excerpt
}

// ExtractedInconsistency represents an inconsistency detected during extraction
type ExtractedInconsistency struct {
	ID               string                      `json:"id"`                // Inconsistency ID
	Type             string                      `json:"type"`              // Type: amount, temporal, authority, procedural, obligation, provenance
	Severity         string                      `json:"severity"`          // Severity: high, medium, low
	Description      string                      `json:"description"`       // Human-readable description
	Documents        []string                    `json:"documents"`         // Document IDs involved
	EntitiesInvolved []string                    `json:"entities_involved"` // Entity IDs involved
	Evidence         InconsistencyEvidence       `json:"evidence"`          // Evidence from both sides
	Confidence       float32                     `json:"confidence"`        // Detection confidence (0.0-1.0)
}

// InconsistencyEvidence contains evidence from both sides of a contradiction
type InconsistencyEvidence struct {
	SideA InconsistencySide `json:"side_a"` // Evidence from side A
	SideB InconsistencySide `json:"side_b"` // Evidence from side B
}

// InconsistencySide represents evidence from one side of a contradiction
type InconsistencySide struct {
	Doc     string `json:"doc"`     // Document ID
	Section string `json:"section"` // Section reference (e.g., "§5")
	Claim   string `json:"claim"`   // The specific claim made
}

// Flatten converts an ExtractionResult to a flat Extraction for scoring
func (er *ExtractionResult) Flatten() *Extraction {
	nodes := make([]ExtractedNode, 0)
	edges := make([]ExtractedEdge, 0)

	for _, doc := range er.Documents {
		nodes = append(nodes, doc.Nodes...)
		edges = append(edges, doc.Edges...)
	}

	// Parse timestamp or use current time
	var timestamp time.Time
	if er.Metadata.Timestamp != "" {
		timestamp, _ = time.Parse(time.RFC3339, er.Metadata.Timestamp)
	} else {
		timestamp = time.Now()
	}

	return &Extraction{
		Corpus:         er.Corpus,
		PromptVersion:  er.PromptVersion,
		Nodes:          nodes,
		Edges:          edges,
		Inconsistencies: er.Inconsistencies,
		Timestamp:      timestamp,
	}
}
