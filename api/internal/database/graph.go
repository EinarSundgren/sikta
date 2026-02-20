package database

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Graph helper types and methods for working with Node, Edge, Provenance

// NodeType represents the type of a node
type NodeType string

const (
	NodeTypeDocument   NodeType = "document"
	NodeTypeEvent      NodeType = "event"
	NodeTypeAttribute  NodeType = "attribute"
	NodeTypeRelation   NodeType = "relation"
	NodeTypePerson     NodeType = "person"
	NodeTypePlace      NodeType = "place"
	NodeTypeOrganization NodeType = "organization"
	NodeTypeObject     NodeType = "object"
	NodeTypeChunk      NodeType = "chunk"
	NodeTypeValue      NodeType = "value"
	NodeTypeInconsistency NodeType = "inconsistency"
	NodeTypeReviewAction NodeType = "review_action"
)

// EdgeType represents the type of an edge
type EdgeType string

const (
	EdgeTypeInvolvedIn EdgeType = "involved_in"
	EdgeTypeSameAs     EdgeType = "same_as"
	EdgeTypeRelatedTo  EdgeType = "related_to"
	EdgeTypeLocatedAt  EdgeType = "located_at"
	EdgeTypeCauses     EdgeType = "causes"
	EdgeTypeAsserts    EdgeType = "asserts"
	EdgeTypeContradicts EdgeType = "contradicts"
	EdgeTypePerformedBy EdgeType = "performed_by"
	EdgeTypeApproves   EdgeType = "approves"
	EdgeTypeRejects    EdgeType = "rejects"
	EdgeTypeHasValue   EdgeType = "has_value"
)

// Modality represents the type of claim
type Modality string

const (
	ModalityAsserted    Modality = "asserted"
	ModalityHypothetical Modality = "hypothetical"
	ModalityDenied      Modality = "denied"
	ModalityConditional Modality = "conditional"
	ModalityInferred    Modality = "inferred"
	ModalityObligatory  Modality = "obligatory"
	ModalityPermitted   Modality = "permitted"
)

// ReviewStatus represents the review status of a provenance record
type ReviewStatus string

const (
	StatusPending  ReviewStatus = "pending"
	StatusApproved ReviewStatus = "approved"
	StatusRejected ReviewStatus = "rejected"
	StatusEdited   ReviewStatus = "edited"
)

// NodeWithProvenance extends Node with its provenance records
type NodeWithProvenance struct {
	Node       Node
	Provenance []Provenance
}

// EdgeWithProvenance extends Edge with its provenance records
type EdgeWithProvenance struct {
	Edge       Edge
	Provenance []Provenance
}

// Location represents where in a source something was found
type Location struct {
	Page        int    `json:"page,omitempty"`
	Section     string `json:"section,omitempty"`
	Chapter     string `json:"chapter,omitempty"`
	Paragraph   int    `json:"paragraph,omitempty"`
	CharStart   int    `json:"char_start,omitempty"`
	CharEnd     int    `json:"char_end,omitempty"`
}

// Properties helper methods for Node

// GetProperty retrieves a property value by key
func (n *Node) GetProperty(key string, defaultValue interface{}) interface{} {
	if n.Properties == nil {
		return defaultValue
	}
	var result map[string]interface{}
	if err := json.Unmarshal(n.Properties, &result); err != nil {
		return defaultValue
	}
	if val, ok := result[key]; ok {
		return val
	}
	return defaultValue
}

// SetProperty sets a property value (returns modified properties bytes)
func (n *Node) SetProperty(key string, value interface{}) ([]byte, error) {
	var props map[string]interface{}
	if n.Properties != nil {
		json.Unmarshal(n.Properties, &props)
	} else {
		props = make(map[string]interface{})
	}
	props[key] = value
	return json.Marshal(props)
}

// Properties helper methods for Edge

// GetProperty retrieves a property value by key
func (e *Edge) GetProperty(key string, defaultValue interface{}) interface{} {
	if e.Properties == nil {
		return defaultValue
	}
	var result map[string]interface{}
	if err := json.Unmarshal(e.Properties, &result); err != nil {
		return defaultValue
	}
	if val, ok := result[key]; ok {
		return val
	}
	return defaultValue
}

// Location helper methods for Provenance

// GetLocation parses the location JSONB into a Location struct
func (p *Provenance) GetLocation() Location {
	if p.Location == nil {
		return Location{}
	}
	var loc Location
	json.Unmarshal(p.Location, &loc)
	return loc
}

// SetLocation converts a Location struct to JSONB bytes
func (p *Provenance) SetLocation(loc Location) ([]byte, error) {
	return json.Marshal(loc)
}

// Effective confidence combines source trust and assertion confidence
func (p *Provenance) EffectiveConfidence() float32 {
	return p.Confidence * p.Trust
}

// IsPending returns true if the status is pending
func (p *Provenance) IsPending() bool {
	return p.Status == string(StatusPending)
}

// IsApproved returns true if the status is approved
func (p *Provenance) IsApproved() bool {
	return p.Status == string(StatusApproved)
}

// IsRejected returns true if the status is rejected
func (p *Provenance) IsRejected() bool {
	return p.Status == string(StatusRejected)
}

// HasTimeInfo returns true if any temporal information is present
func (p *Provenance) HasTimeInfo() bool {
	return p.ClaimedTimeStart.Valid || p.ClaimedTimeEnd.Valid || p.ClaimedTimeText.Valid
}

// HasGeoInfo returns true if any geographic information is present
func (p *Provenance) HasGeoInfo() bool {
	return p.ClaimedGeoRegion.Valid || p.ClaimedGeoText.Valid
}

// TimeRange returns the time range as Go time values
func (p *Provenance) TimeRange() (start, end *time.Time, hasRange bool) {
	if p.ClaimedTimeStart.Valid {
		start = &p.ClaimedTimeStart.Time
	}
	if p.ClaimedTimeEnd.Valid {
		end = &p.ClaimedTimeEnd.Time
	}
	hasRange = p.ClaimedTimeStart.Valid || p.ClaimedTimeEnd.Valid
	return
}

// PgTime converts a time.Time to pgtype.Timestamptz
func PgTime(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// JSONB converts a map to JSONB bytes
func JSONB(data map[string]interface{}) []byte {
	if data == nil {
		return []byte("{}")
	}
	b, _ := json.Marshal(data)
	return b
}
