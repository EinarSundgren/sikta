package graph

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
)

// Migrator handles migration from legacy tables to graph model
type Migrator struct {
	db     *database.Queries
	graph  *Service
	logger *slog.Logger
}

// NewMigrator creates a new migrator
func NewMigrator(db *database.Queries, graph *Service, logger *slog.Logger) *Migrator {
	return &Migrator{
		db:     db,
		graph:  graph,
		logger: logger,
	}
}

// MigrateSourceToNode converts a source (document) to a document node
func (m *Migrator) MigrateSourceToNode(ctx context.Context, source *database.Source) (uuid.UUID, error) {
	sourceID, err := uuid.FromBytes(source.ID.Bytes[:16])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse source ID: %w", err)
	}

	// Create properties map
	properties := map[string]interface{}{
		"filename":   source.Filename,
		"file_type":  source.FileType,
		"is_demo":    source.IsDemo,
		"source_id":  sourceID.String(), // Store original source ID for lookup
	}

	if source.TotalPages.Valid {
		properties["total_pages"] = source.TotalPages.Int32
	}

	// Get trust value
	trust := float32(1.0)
	if source.SourceTrust.Valid {
		trust = source.SourceTrust.Float32
	}

	// Create document node
	nodeID, err := m.graph.CreateNode(ctx, CreateNodeParams{
		NodeType:   database.NodeTypeDocument,
		Label:      source.Title,
		Properties: properties,
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Create self-referencing provenance
	_, err = m.graph.CreateProvenance(ctx, CreateProvenanceParams{
		TargetType: "node",
		TargetID:   nodeID,
		SourceID:   nodeID,
		Excerpt:    fmt.Sprintf("Document: %s", source.Title),
		Confidence: 1.0,
		Trust:      trust,
		Modality:   database.ModalityAsserted,
		Status:     database.StatusApproved,
	})

	m.logger.Info("source migrated to node", "source_id", sourceID, "node_id", nodeID)
	return nodeID, nil
}

// MigrateChunkToNode converts a chunk to a chunk node with provenance
func (m *Migrator) MigrateChunkToNode(ctx context.Context, chunk *database.Chunk, docNodeID uuid.UUID) (uuid.UUID, error) {
	chunkID, err := uuid.FromBytes(chunk.ID.Bytes[:16])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse chunk ID: %w", err)
	}

	properties := map[string]interface{}{
		"chunk_index": chunk.ChunkIndex,
	}

	if chunk.ChapterTitle.Valid {
		properties["chapter_title"] = chunk.ChapterTitle.String
	}
	if chunk.ChapterNumber.Valid {
		properties["chapter_number"] = chunk.ChapterNumber.Int32
	}

	// Create chunk node
	nodeID, err := m.graph.CreateNode(ctx, CreateNodeParams{
		NodeType:   database.NodeTypeChunk,
		Label:      fmt.Sprintf("Chunk %d", chunk.ChunkIndex),
		Properties: properties,
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Create provenance
	location := database.Location{
		Chapter: chunk.ChapterTitle.String,
	}
	if chunk.PageStart.Valid {
		location.Page = int(chunk.PageStart.Int32)
	}

	_, err = m.graph.CreateProvenance(ctx, CreateProvenanceParams{
		TargetType: "node",
		TargetID:   nodeID,
		SourceID:   docNodeID,
		Excerpt:    chunk.Content,
		Location:   location,
		Confidence: 1.0,
		Trust:      1.0,
		Modality:   database.ModalityAsserted,
		Status:     database.StatusApproved,
	})

	m.logger.Info("chunk migrated to node", "chunk_id", chunkID, "node_id", nodeID)
	return nodeID, nil
}

// MigrateClaimToNode converts a claim (event/attribute/relation) to a node with provenance
func (m *Migrator) MigrateClaimToNode(ctx context.Context, claim *database.Claim, docNodeID uuid.UUID) (uuid.UUID, error) {
	claimID, err := uuid.FromBytes(claim.ID.Bytes[:16])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse claim ID: %w", err)
	}

	// Determine node type from claim type
	var nodeType string
	switch claim.ClaimType {
	case "event":
		nodeType = database.NodeTypeEvent
	case "attribute":
		nodeType = database.NodeTypeAttribute
	case "relation":
		nodeType = database.NodeTypeRelation
	default:
		nodeType = database.NodeTypeEvent
	}

	// Build properties — no positions here; ordering is expressed as provenance records
	properties := map[string]interface{}{
		"event_type":  claim.EventType.String,
		"description": claim.Description.String,
	}

	// Create claim node
	nodeID, err := m.graph.CreateNode(ctx, CreateNodeParams{
		NodeType:   nodeType,
		Label:      claim.Title,
		Properties: properties,
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Create main provenance with temporal info
	var claimedTimeText string
	if claim.DateText.Valid {
		claimedTimeText = claim.DateText.String
	}

	_, err = m.graph.CreateProvenance(ctx, CreateProvenanceParams{
		TargetType:      "node",
		TargetID:        nodeID,
		SourceID:        docNodeID,
		Excerpt:         claim.Description.String,
		Location:        database.Location{},
		Confidence:      claim.Confidence,
		Trust:           1.0,
		Modality:        database.ModalityAsserted,
		Status:          database.ReviewStatus(claim.ReviewStatus),
		ClaimedTimeText: claimedTimeText,
	})
	if err != nil {
		m.logger.Warn("failed to create provenance for claim", "claim_id", claimID, "error", err)
	}

	// Create narrative ordering provenance — position in story sequence is an inferred claim
	_, err = m.graph.CreateProvenance(ctx, CreateProvenanceParams{
		TargetType: "node",
		TargetID:   nodeID,
		SourceID:   docNodeID,
		Excerpt:    fmt.Sprintf("Narrative position: %d", claim.NarrativePosition),
		Location:   database.Location{PositionType: "narrative", Position: int(claim.NarrativePosition)},
		Confidence: 1.0,
		Trust:      1.0,
		Modality:   database.ModalityInferred,
		Status:     database.StatusApproved,
	})
	if err != nil {
		m.logger.Warn("failed to create narrative ordering provenance", "claim_id", claimID, "error", err)
	}

	// Create chronological ordering provenance (only if set)
	if claim.ChronologicalPosition.Valid {
		_, err = m.graph.CreateProvenance(ctx, CreateProvenanceParams{
			TargetType: "node",
			TargetID:   nodeID,
			SourceID:   docNodeID,
			Excerpt:    fmt.Sprintf("Chronological position: %d", claim.ChronologicalPosition.Int32),
			Location:   database.Location{PositionType: "chronological", Position: int(claim.ChronologicalPosition.Int32)},
			Confidence: 1.0,
			Trust:      1.0,
			Modality:   database.ModalityInferred,
			Status:     database.StatusApproved,
		})
		if err != nil {
			m.logger.Warn("failed to create chronological ordering provenance", "claim_id", claimID, "error", err)
		}
	}

	m.logger.Info("claim migrated to node", "claim_id", claimID, "node_id", nodeID)
	return nodeID, nil
}

// MigrateEntityToNode converts an entity to an entity node with provenance
func (m *Migrator) MigrateEntityToNode(ctx context.Context, entity *database.Entity, docNodeID uuid.UUID) (uuid.UUID, error) {
	entityID, err := uuid.FromBytes(entity.ID.Bytes[:16])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse entity ID: %w", err)
	}

	// Build properties — appearance positions are resolved values, omit per spec
	properties := map[string]interface{}{
		"entity_type": entity.EntityType,
		"aliases":     entity.Aliases,
	}

	description := ""
	if entity.Description.Valid {
		description = entity.Description.String
	}
	properties["description"] = description

	// Create entity node
	nodeID, err := m.graph.CreateNode(ctx, CreateNodeParams{
		NodeType:   entity.EntityType,
		Label:      entity.Name,
		Properties: properties,
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Create provenance
	_, err = m.graph.CreateProvenance(ctx, CreateProvenanceParams{
		TargetType: "node",
		TargetID:   nodeID,
		SourceID:   docNodeID,
		Excerpt:    description,
		Confidence: entity.Confidence,
		Trust:      1.0,
		Modality:   database.ModalityAsserted,
		Status:     database.ReviewStatus(entity.ReviewStatus),
	})
	if err != nil {
		m.logger.Warn("failed to create provenance for entity", "entity_id", entityID, "error", err)
	}

	m.logger.Info("entity migrated to node", "entity_id", entityID, "node_id", nodeID)
	return nodeID, nil
}

// MigrateRelationshipToEdge converts a relationship to an edge with provenance
func (m *Migrator) MigrateRelationshipToEdge(ctx context.Context, relationship *database.Relationship, docNodeID uuid.UUID, entityIDMap map[string]uuid.UUID) (uuid.UUID, error) {
	relID, err := uuid.FromBytes(relationship.ID.Bytes[:16])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse relationship ID: %w", err)
	}

	// Get entity IDs from map
	entityAID, ok := entityIDMap[database.UUIDStr(relationship.EntityAID)]
	if !ok {
		return uuid.Nil, fmt.Errorf("entity A not found in map: %s", database.UUIDStr(relationship.EntityAID))
	}
	entityBID, ok := entityIDMap[database.UUIDStr(relationship.EntityBID)]
	if !ok {
		return uuid.Nil, fmt.Errorf("entity B not found in map: %s", database.UUIDStr(relationship.EntityBID))
	}

	// Create edge
	edgeID, err := m.graph.CreateEdge(ctx, CreateEdgeParams{
		EdgeType:   relationship.RelationshipType,
		SourceNode: entityAID,
		TargetNode: entityBID,
		Properties: map[string]interface{}{
			"description": relationship.Description.String,
		},
		IsNegated: false,
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Create provenance
	description := ""
	if relationship.Description.Valid {
		description = relationship.Description.String
	}

	_, err = m.graph.CreateProvenance(ctx, CreateProvenanceParams{
		TargetType: "edge",
		TargetID:   edgeID,
		SourceID:   docNodeID,
		Excerpt:    description,
		Confidence: relationship.Confidence,
		Trust:      1.0,
		Modality:   database.ModalityAsserted,
		Status:     database.ReviewStatus(relationship.ReviewStatus),
	})

	m.logger.Info("relationship migrated to edge", "rel_id", relID, "edge_id", edgeID)
	return edgeID, nil
}

// MigrateClaimEntityToEdge converts a claim_entity (participation) to an edge
func (m *Migrator) MigrateClaimEntityToEdge(ctx context.Context, claimEntity *database.ClaimEntity, claimNodeID, entityNodeID uuid.UUID) error {
	// Create participates_in edge
	_, err := m.graph.CreateEdge(ctx, CreateEdgeParams{
		EdgeType:   database.EdgeTypeInvolvedIn,
		SourceNode: entityNodeID,
		TargetNode: claimNodeID,
		Properties: map[string]interface{}{
			"role": claimEntity.Role.String,
		},
		IsNegated: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create participation edge: %w", err)
	}

	return nil
}

// MigrateDocument migrates an entire document (source + all related data) to graph
func (m *Migrator) MigrateDocument(ctx context.Context, sourceID uuid.UUID) error {
	// Get source
	source, err := m.db.GetSource(ctx, database.PgUUID(sourceID))
	if err != nil {
		return fmt.Errorf("failed to get source: %w", err)
	}

	// Migrate source to document node
	docNodeID, err := m.MigrateSourceToNode(ctx, source)
	if err != nil {
		return fmt.Errorf("failed to migrate source: %w", err)
	}

	// Get and migrate chunks
	chunks, err := m.db.ListChunksBySource(ctx, database.PgUUID(sourceID))
	if err != nil {
		return fmt.Errorf("failed to get chunks: %w", err)
	}

	chunkNodeIDs := make(map[uuid.UUID]uuid.UUID)
	for _, chunk := range chunks {
		chunkID, _ := uuid.FromBytes(chunk.ID.Bytes[:16])
		nodeID, err := m.MigrateChunkToNode(ctx, chunk, docNodeID)
		if err != nil {
			m.logger.Warn("failed to migrate chunk", "chunk_id", chunkID, "error", err)
			continue
		}
		chunkNodeIDs[chunkID] = nodeID
	}

	// Get and migrate entities
	entities, err := m.db.ListEntitiesBySource(ctx, database.PgUUID(sourceID))
	if err != nil {
		return fmt.Errorf("failed to get entities: %w", err)
	}

	entityNodeIDs := make(map[string]uuid.UUID)
	for _, entity := range entities {
		entityID, _ := uuid.FromBytes(entity.ID.Bytes[:16])
		nodeID, err := m.MigrateEntityToNode(ctx, entity, docNodeID)
		if err != nil {
			m.logger.Warn("failed to migrate entity", "entity_id", entityID, "error", err)
			continue
		}
		entityNodeIDs[database.UUIDStr(entity.ID)] = nodeID
	}

	// Get and migrate claims
	claims, err := m.db.ListClaimsBySource(ctx, database.PgUUID(sourceID))
	if err != nil {
		return fmt.Errorf("failed to get claims: %w", err)
	}

	claimNodeIDs := make(map[uuid.UUID]uuid.UUID)
	for _, claim := range claims {
		claimID, _ := uuid.FromBytes(claim.ID.Bytes[:16])
		nodeID, err := m.MigrateClaimToNode(ctx, claim, docNodeID)
		if err != nil {
			m.logger.Warn("failed to migrate claim", "claim_id", claimID, "error", err)
			continue
		}
		claimNodeIDs[claimID] = nodeID
	}

	// Get and migrate relationships
	relationships, err := m.db.ListRelationshipsBySource(ctx, database.PgUUID(sourceID))
	if err != nil {
		m.logger.Warn("failed to get relationships", "error", err)
	} else {
		for _, relationship := range relationships {
			_, err := m.MigrateRelationshipToEdge(ctx, relationship, docNodeID, entityNodeIDs)
			if err != nil {
				m.logger.Warn("failed to migrate relationship", "rel_id", database.UUIDStr(relationship.ID), "error", err)
				continue
			}
		}
	}

	// Migrate claim_entities (participation)
	for _, claim := range claims {
		claimID, _ := uuid.FromBytes(claim.ID.Bytes[:16])
		claimNodeID, ok := claimNodeIDs[claimID]
		if !ok {
			continue
		}

		// For now, skip claim_entities migration as the query method may not exist
		// This would be added in a later iteration
		_ = claimNodeID
	}

	m.logger.Info("document migrated to graph", "source_id", sourceID, "doc_node_id", docNodeID,
		"chunks", len(chunks), "entities", len(entities), "claims", len(claims))

	return nil
}
