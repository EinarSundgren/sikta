package graph

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
)

// Views provides query-time resolution strategies for graph data
type Views struct {
	db     *database.Queries
	logger *slog.Logger
}

// NewViews creates a new views service
func NewViews(db *database.Queries, logger *slog.Logger) *Views {
	return &Views{
		db:     db,
		logger: logger,
	}
}

// findDocumentNode looks up the document node for a legacy source ID.
// Returns an error if no document node exists for this source.
func (v *Views) findDocumentNode(ctx context.Context, sourceID uuid.UUID) (*database.Node, error) {
	node, err := v.db.GetDocumentNodeByLegacySourceID(ctx, sourceID.String())
	if err != nil {
		return nil, fmt.Errorf("document node not found for source_id %s: %w", sourceID, err)
	}
	return node, nil
}

// GetEventsForTimeline returns event nodes formatted for the timeline view
// using the specified view strategy to resolve conflicts
func (v *Views) GetEventsForTimeline(ctx context.Context, sourceID uuid.UUID, strategy ViewStrategy) ([]TimelineEvent, error) {
	docNode, err := v.findDocumentNode(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	// Get all nodes that have provenance from this document node
	nodes, err := v.db.ListNodesBySource(ctx, docNode.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var events []TimelineEvent
	for _, node := range nodes {
		if node.NodeType != database.NodeTypeEvent {
			continue
		}

		// Get provenance for this node
		provenance, err := v.db.ListProvenanceByTarget(ctx, database.ListProvenanceByTargetParams{
			TargetType: "node",
			TargetID:   node.ID,
		})
		if err != nil {
			v.logger.Warn("failed to get provenance for node", "node_id", node.ID, "error", err)
			continue
		}

		// Apply view strategy only to assertion provenance, not ordering records
		selectedProv := v.selectProvenance(assertionOnly(provenance), strategy)
		confidence := float32(0)
		reviewStatus := string(database.StatusPending)
		if selectedProv != nil {
			confidence = selectedProv.Confidence
			reviewStatus = selectedProv.Status
		}

		// Extract temporal info from all provenance records
		var dateStart, dateEnd *time.Time
		var dateText string
		for _, prov := range provenance {
			if prov.ClaimedTimeStart.Valid {
				t := prov.ClaimedTimeStart.Time
				dateStart = &t
			}
			if prov.ClaimedTimeEnd.Valid {
				t := prov.ClaimedTimeEnd.Time
				dateEnd = &t
			}
			if prov.ClaimedTimeText.Valid && prov.ClaimedTimeText.String != "" {
				dateText = prov.ClaimedTimeText.String
			}
		}

		// Get position from ordering provenance records
		narrativePos, chronoPos := extractOrderingFromProvenance(provenance)

		event := TimelineEvent{
			ID:                    database.UUIDStr(node.ID),
			Title:                 node.Label,
			Description:           node.GetProperty("description", "").(string),
			Type:                  node.GetProperty("event_type", "action").(string),
			DateText:              dateText,
			NarrativePosition:     narrativePos,
			ChronologicalPosition: chronoPos,
			Confidence:            confidence,
			ReviewStatus:          reviewStatus,
			DateStart:             dateStart,
			DateEnd:               dateEnd,
		}

		events = append(events, event)
	}

	return events, nil
}

// GetEntitiesForGraph returns entity nodes for the graph view
func (v *Views) GetEntitiesForGraph(ctx context.Context, sourceID uuid.UUID) ([]GraphEntity, error) {
	docNode, err := v.findDocumentNode(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	nodes, err := v.db.ListNodesBySource(ctx, docNode.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var entities []GraphEntity
	for _, node := range nodes {
		if node.NodeType != database.NodeTypePerson &&
			node.NodeType != database.NodeTypePlace &&
			node.NodeType != database.NodeTypeOrganization &&
			node.NodeType != database.NodeTypeObject {
			continue
		}

		// Get provenance
		provenance, err := v.db.ListProvenanceByTarget(ctx, database.ListProvenanceByTargetParams{
			TargetType: "node",
			TargetID:   node.ID,
		})
		if err != nil {
			continue
		}

		selectedProv := v.selectProvenance(assertionOnly(provenance), ViewStrategyTrustWeighted)
		confidence := float32(0)
		reviewStatus := string(database.StatusPending)
		if selectedProv != nil {
			confidence = selectedProv.Confidence
			reviewStatus = selectedProv.Status
		}

		// Extract aliases — JSON unmarshal produces []interface{}, not []string
		var aliases []string
		if raw := node.GetProperty("aliases", nil); raw != nil {
			if arr, ok := raw.([]interface{}); ok {
				for _, item := range arr {
					if s, ok := item.(string); ok {
						aliases = append(aliases, s)
					}
				}
			}
		}

		entity := GraphEntity{
			ID:           database.UUIDStr(node.ID),
			Name:         node.Label,
			Type:         node.NodeType,
			Aliases:      aliases,
			Description:  node.GetProperty("description", "").(string),
			Confidence:   confidence,
			ReviewStatus: reviewStatus,
		}

		// Get same_as edges for identity claims
		edges, err := v.db.ListEdgesBySource(ctx, node.ID)
		if err == nil {
			for _, edge := range edges {
				if edge.EdgeType == database.EdgeTypeSameAs {
					entity.SameAsEdges = append(entity.SameAsEdges, SameAsEdge{
						TargetEntityID: database.UUIDStr(edge.TargetNode),
						Confidence:     1.0, // Would come from edge properties
						Reasoning:      "",
					})
				}
			}
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

// GetRelationshipsForGraph returns relationship edges for the graph view
func (v *Views) GetRelationshipsForGraph(ctx context.Context, sourceID uuid.UUID) ([]GraphRelationship, error) {
	docNode, err := v.findDocumentNode(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	edges, err := v.db.ListEdgesBySourceDocument(ctx, docNode.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}

	var relationships []GraphRelationship
	for _, edge := range edges {
		// Skip structural/internal edges — only surface semantic relationship edges
		if edge.EdgeType == database.EdgeTypeInvolvedIn ||
			edge.EdgeType == database.EdgeTypeSameAs {
			continue
		}

		// Get provenance for this edge
		provenance, err := v.db.ListProvenanceByTarget(ctx, database.ListProvenanceByTargetParams{
			TargetType: "edge",
			TargetID:   edge.ID,
		})
		if err != nil {
			v.logger.Warn("failed to get provenance for edge", "edge_id", edge.ID, "error", err)
		}

		selectedProv := v.selectProvenance(provenance, ViewStrategyTrustWeighted)
		confidence := float32(0)
		reviewStatus := string(database.StatusPending)
		if selectedProv != nil {
			confidence = selectedProv.Confidence
			reviewStatus = selectedProv.Status
		}

		rel := GraphRelationship{
			ID:               database.UUIDStr(edge.ID),
			EntityAID:        database.UUIDStr(edge.SourceNode),
			EntityBID:        database.UUIDStr(edge.TargetNode),
			RelationshipType: edge.EdgeType,
			Description:      edge.GetProperty("description", "").(string),
			Confidence:       confidence,
			ReviewStatus:     reviewStatus,
		}
		relationships = append(relationships, rel)
	}

	return relationships, nil
}

// ResolveIdentity returns the "canonical" entity ID based on same_as edges and view strategy
// For human_decided strategy, returns the approved identity
// For trust_weighted, returns the highest confidence identity
func (v *Views) ResolveIdentity(ctx context.Context, entityID uuid.UUID, strategy ViewStrategy) (uuid.UUID, error) {
	// Get same_as edges for this entity
	edges, err := v.db.ListEdgesBySource(ctx, database.PgUUID(entityID))
	if err != nil {
		return entityID, nil
	}

	var candidateIDs []uuid.UUID
	for _, edge := range edges {
		if edge.EdgeType == database.EdgeTypeSameAs {
			targetID, _ := uuid.FromBytes(edge.TargetNode.Bytes[:16])
			candidateIDs = append(candidateIDs, targetID)
		}
	}

	if len(candidateIDs) == 0 {
		return entityID, nil
	}

	// For now, return the first candidate
	// In production, this would check provenance, confidence, approval status
	return candidateIDs[0], nil
}

// assertionOnly filters out ordering provenance records (modality="inferred",
// PositionType set), returning only the substantive assertion records.
// This prevents ordering records from polluting confidence/status selection.
func assertionOnly(provenance []*database.Provenance) []*database.Provenance {
	var result []*database.Provenance
	for _, p := range provenance {
		if p.GetLocation().PositionType == "" {
			result = append(result, p)
		}
	}
	return result
}

// extractOrderingFromProvenance reads narrative and chronological positions from
// ordering provenance records (those with location.position_type set).
func extractOrderingFromProvenance(provenance []*database.Provenance) (narrativePos, chronoPos int32) {
	for _, prov := range provenance {
		loc := prov.GetLocation()
		switch loc.PositionType {
		case "narrative":
			narrativePos = int32(loc.Position)
		case "chronological":
			chronoPos = int32(loc.Position)
		}
	}
	return
}

// selectProvenance applies the view strategy to select the best provenance record
func (v *Views) selectProvenance(provenance []*database.Provenance, strategy ViewStrategy) *database.Provenance {
	if len(provenance) == 0 {
		return nil
	}

	// ViewStrategyHumanDecided: prefer approved provenance, fall through to trust weighted
	if strategy == ViewStrategyHumanDecided {
		for _, p := range provenance {
			if p.Status == string(database.StatusApproved) {
				return p
			}
		}
		// No approved provenance, use trust weighted
		strategy = ViewStrategyTrustWeighted
	}

	switch strategy {
	case ViewStrategyTrustWeighted:
		// Select highest effective confidence (trust * assertion confidence)
		var best *database.Provenance
		bestScore := float32(0)
		for _, p := range provenance {
			score := p.Trust * p.Confidence
			if score > bestScore {
				best = p
				bestScore = score
			}
		}
		if best == nil && len(provenance) > 0 {
			best = provenance[0]
		}
		return best

	case ViewStrategySingleSource:
		// Return the first (most recent) provenance
		return provenance[0]

	case ViewStrategyConflict:
		// Return the provenance with lowest confidence (to show disagreement)
		worst := provenance[0]
		worstScore := worst.Trust * worst.Confidence
		for _, p := range provenance[1:] {
			score := p.Trust * p.Confidence
			if score < worstScore {
				worst = p
				worstScore = score
			}
		}
		return worst

	case ViewStrategyMajority:
		// For timeline, this would aggregate multiple sources
		// For now, return trust-weighted
		return v.selectProvenance(provenance, ViewStrategyTrustWeighted)

	default:
		return provenance[0]
	}
}
