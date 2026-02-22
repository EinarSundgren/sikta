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
	InconsistencyRecall  float64               // found / expected
	InconsistencyDetails []InconsistencyMatch // Per-inconsistency breakdown

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
	ID          string            // Entity ID (e.g., "E1")
	Label       string            // Entity label (e.g., "Anna Lindqvist")
	Type        string            // Entity type (person, organization, etc.)
	Aliases     []string          // Alternative names/references
	Properties  map[string]interface{} // Additional properties
	MentionedIn []string          // Document IDs where this entity appears
}

// ManifestEvent represents an expected event extraction
type ManifestEvent struct {
	ID                string   // Event ID (e.g., "V1")
	Label             string   // Event label (e.g., "Fuktsinspektion")
	Type              string   // Event type (decision, inspection, etc.)
	ClaimedTimeText   string   // Raw time text from document
	ClaimedTimeStart  string   // ISO date if available
	ClaimedTimeEnd    string   // ISO date if available
	Entities          []string // Entity IDs involved in this event
	SourceDoc         string   // Document ID where this event appears
	SourceSection     string   // Section reference (e.g., "§5")
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

// Extraction represents the output from an extraction run
type Extraction struct {
	Corpus        string            // Corpus identifier
	PromptVersion string            // Prompt version used
	Nodes         []ExtractedNode  // Extracted nodes
	Edges         []ExtractedEdge  // Extracted edges
	Timestamp     time.Time         // When extraction was run
}

// ExtractedNode represents a node from extraction output
type ExtractedNode struct {
	ID              string                 // Node ID
	NodeType        string                 // Node type
	Label           string                 // Node label
	Properties      map[string]interface{} // Node properties
	Confidence      float32                // Confidence score
	Modality        string                 // Modality (asserted, etc.)
	Excerpt         string                 // Source excerpt
	ClaimedTimeText string                 // Raw time text
	ClaimedGeoText  string                 // Raw location text
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
