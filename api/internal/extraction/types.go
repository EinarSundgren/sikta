package extraction

import "time"

// ExtractionResponse represents the structured output from the LLM.
type ExtractionResponse struct {
	Events       []Event         `json:"events"`
	Entities     []Entity        `json:"entities"`
	Relationships []Relationship `json:"relationships"`
}

// Event represents an extracted event.
type Event struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	DateText    string  `json:"date_text"`
	Confidence  float64 `json:"confidence"`
	Excerpt     string  `json:"excerpt"`
}

// Entity represents an extracted entity.
type Entity struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Aliases    []string `json:"aliases"`
	Confidence float64  `json:"confidence"`
	Excerpt    string   `json:"excerpt"`
}

// Relationship represents an extracted relationship between entities.
type Relationship struct {
	EntityA     string  `json:"entity_a"`
	EntityB     string  `json:"entity_b"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// ConfidenceClassification represents confidence classification for an item.
type ConfidenceClassification struct {
	DatePrecision     string  `json:"date_precision"`
	EntityCertainty   string  `json:"entity_certainty"`
	ConfidenceScore   float64 `json:"confidence_score"`
	Reasoning         string  `json:"reasoning"`
}

// ChronologicalOrder represents the timeline ordering of events.
type ChronologicalOrder struct {
	ChronologicalOrder []TimelinePosition `json:"chronological_order"`
	Anomalies          []Anomaly          `json:"anomalies"`
}

// TimelinePosition represents an event's position in the timeline.
type TimelinePosition struct {
	EventID              string `json:"event_id"`
	ChronologicalPosition int    `json:"chronological_position"`
	Reasoning            string `json:"reasoning"`
}

// Anomaly represents a detected temporal impossibility.
type Anomaly struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
}

// ParsedDate represents a normalized date range.
type ParsedDate struct {
	DateStart     *time.Time `json:"date_start"`
	DateEnd       *time.Time `json:"date_end"`
	DatePrecision string     `json:"date_precision"`
}
