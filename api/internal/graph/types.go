package graph

import (
	"time"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
)

// ViewStrategy determines how to resolve conflicting claims
type ViewStrategy string

const (
	ViewStrategySingleSource  ViewStrategy = "single_source"  // Use one provenance record
	ViewStrategyTrustWeighted ViewStrategy = "trust_weighted" // Highest trust * confidence
	ViewStrategyMajority       ViewStrategy = "majority"        // Most sources agree
	ViewStrategyHumanDecided   ViewStrategy = "human_decided"  // status='approved' wins
	ViewStrategyConflict        ViewStrategy = "conflict"       // Show all claims
)

// CreateNodeParams contains parameters for creating a node
type CreateNodeParams struct {
	NodeType   string // open text — any value is valid
	Label      string
	Properties map[string]interface{}
}

// CreateEdgeParams contains parameters for creating an edge
type CreateEdgeParams struct {
	EdgeType   string // open text — any value is valid
	SourceNode uuid.UUID
	TargetNode uuid.UUID
	Properties map[string]interface{}
	IsNegated  bool
}

// CreateProvenanceParams contains parameters for creating provenance
type CreateProvenanceParams struct {
	TargetType string // "node" or "edge"
	TargetID   uuid.UUID
	SourceID   uuid.UUID
	Excerpt    string
	Location   database.Location
	Confidence float32
	Trust      float32
	Modality   string               // open text — any value is valid
	Status     database.ReviewStatus

	// Temporal claims
	ClaimedTimeStart *time.Time
	ClaimedTimeEnd   *time.Time
	ClaimedTimeText  string

	// Geo claims
	ClaimedGeoRegion string
	ClaimedGeoText   string

	// Attribution
	ClaimedBy *uuid.UUID // e.g., a reviewer node
}

// TimelineEvent represents an event for the timeline view (compatible with frontend)
type TimelineEvent struct {
	ID                    string
	Title                 string
	Description           string
	Type                  string
	DateText              string
	DateStart             *time.Time
	DateEnd               *time.Time
	DatePrecision         string
	NarrativePosition     int32
	ChronologicalPosition int32
	Confidence            float32
	ReviewStatus          string
	Entities              []TimelineEntity
	SourceReferences      []SourceReference
	Inconsistencies       []InconsistencyRef
}

// TimelineEntity represents an entity in a timeline event
type TimelineEntity struct {
	ID   string
	Name string
	Type string
	Role string
}

// SourceReference represents a link to source text
type SourceReference struct {
	ID       string
	Excerpt  string
	Location database.Location
}

// InconsistencyRef represents a reference to an inconsistency
type InconsistencyRef struct {
	ID       string
	Type     string
	Severity string
}

// GraphEntity represents an entity for the graph view
type GraphEntity struct {
	ID              string
	Name            string
	Type            string
	Aliases         []string
	Description     string
	FirstAppearance int32
	LastAppearance  int32
	Confidence      float32
	ReviewStatus    string
	RelationshipCount int
	// Identity claims
	SameAsEdges []SameAsEdge
}

// SameAsEdge represents an identity claim
type SameAsEdge struct {
	TargetEntityID string
	Confidence     float32
	Reasoning      string
}

// GraphRelationship represents a relationship for the graph view
type GraphRelationship struct {
	ID               string
	EntityAID        string
	EntityBID        string
	RelationshipType string
	Description      string
	Confidence       float32
	ReviewStatus     string
	StartClaimID     *string
	EndClaimID       *string
}
