package graph

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

// GetEventsForTimeline returns event nodes formatted for the timeline view
// using the specified view strategy to resolve conflicts
func (v *Views) GetEventsForTimeline(ctx context.Context, sourceID uuid.UUID, strategy ViewStrategy) ([]TimelineEvent, error) {
	// Get all event nodes for this source
	nodes, err := v.db.ListNodesBySource(ctx, database.PgUUID(sourceID))
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var events []TimelineEvent
	for _, node := range nodes {
		if node.NodeType != string(database.NodeTypeEvent) {
			continue
		}

		// Get provenance for this node
		provenance, err := v.db.ListProvenanceByTarget(ctx, "node", database.PgUUID(node.ID))
		if err != nil {
			v.logger.Warn("failed to get provenance for node", "node_id", node.ID, "error", err)
			continue
		}

		// Apply view strategy to select which provenance to use
		selectedProv := v.selectProvenance(provenance, strategy)

		// Extract temporal info
		var dateStart, dateEnd *int64
		var dateText string
		for _, prov := range provenance {
			if prov.ClaimedTimeStart.Valid {
				ts := prov.ClaimedTimeStart.Time.Unix()
				dateStart = &ts
			}
			if prov.ClaimedTimeEnd.Valid {
				ts := prov.ClaimedTimeEnd.Time.Unix()
				dateEnd = &ts
			}
			if prov.ClaimedTimeText.Valid && prov.ClaimedTimeText.String != "" {
				dateText = prov.ClaimedTimeText.String
			}
		}

		// Get position from properties
		narrativePos := int32(node.GetProperty("narrative_position", int32(0)).(int32))
		chronoPos := int32(node.GetProperty("chronological_position", int32(0)).(int32))

		event := TimelineEvent{
			ID:                    database.UUIDStr(node.ID),
			Title:                 node.Label,
			Description:           node.GetProperty("description", "").(string),
			Type:                  node.GetProperty("event_type", "action").(string),
			DateText:              dateText,
			NarrativePosition:     narrativePos,
			ChronologicalPosition: chronoPos,
			Confidence:            selectedProv.Confidence,
			ReviewStatus:          selectedProv.Status,
		}

		if dateStart != nil {
			t := pgtype.Time{Time: *selectedProv.ClaimedTimeStart.Time, Valid: true}
			event.DateStart = &t.Time
		}
		if dateEnd != nil {
			t := pgtype.Time{Time: *selectedProv.ClaimedTimeEnd.Time, Valid: true}
			event.DateEnd = &t.Time
		}

		events = append(events, event)
	}

	return events, nil
}

// GetEntitiesForGraph returns entity nodes for the graph view
func (v *Views) GetEntitiesForGraph(ctx context.Context, sourceID uuid.UUID) ([]GraphEntity, error) {
	nodes, err := v.db.ListNodesBySource(ctx, database.PgUUID(sourceID))
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var entities []GraphEntity
	for _, node := range nodes {
		nodeType := database.NodeType(node.NodeType)
		if nodeType != database.NodeTypePerson &&
			nodeType != database.NodeTypePlace &&
			nodeType != database.NodeTypeOrganization &&
			nodeType != database.NodeTypeObject {
			continue
		}

		// Get provenance
		provenance, err := v.db.ListProvenanceByTarget(ctx, "node", database.PgUUID(node.ID))
		if err != nil {
			continue
		}
		selectedProv := v.selectProvenance(provenance, ViewStrategyTrustWeighted)

		// Extract aliases
		var aliases []string
		if aliasVal := node.GetProperty("aliases", []string{}); aliasVal != nil {
			aliases = aliasVal.([]string)
		}

		entity := GraphEntity{
			ID:           database.UUIDStr(node.ID),
			Name:         node.Label,
			Type:         node.NodeType,
			Aliases:      aliases,
			Description:  node.GetProperty("description", "").(string),
			Confidence:   selectedProv.Confidence,
			ReviewStatus: selectedProv.Status,
		}

		// Get same_as edges for identity claims
		edges, err := v.db.ListEdgesBySource(ctx, database.PgUUID(node.ID))
		if err == nil {
			for _, edge := range edges {
				if edge.EdgeType == string(database.EdgeTypeSameAs) {
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
	// This would query edges between entity nodes
	// For now, return empty slice
	return []GraphRelationship{}, nil
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
		if edge.EdgeType == string(database.EdgeTypeSameAs) {
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

// selectProvenance applies the view strategy to select the best provenance record
func (v *Views) selectProvenance(provenance []database.Provenance, strategy ViewStrategy) database.Provenance {
	if len(provenance) == 0 {
		return database.Provenance{}
	}

	switch strategy {
	case ViewStrategyHumanDecided:
		// Prefer approved provenance
		for _, p := range provenance {
			if p.Status == string(database.StatusApproved) {
				return p
			}
		}
		// Fall through to trust weighted

	case ViewStrategyTrustWeighted:
		// Select highest effective confidence (trust * assertion confidence)
		var best database.Provenance
		bestScore := float32(0)
		for _, p := range provenance {
			score := p.EffectiveConfidence()
			if score > bestScore {
				best = p
				bestScore = score
			}
		}
		return best

	case ViewStrategySingleSource:
		// Return the first (most recent) provenance
		return provenance[0]

	case ViewStrategyConflict:
		// Return the provenance with lowest confidence (to show disagreement)
		worst := provenance[0]
		worstScore := worst.EffectiveConfidence()
		for _, p := range provenance[1:] {
			score := p.EffectiveConfidence()
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
