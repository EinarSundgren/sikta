package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service handles graph operations
type Service struct {
	db     *database.Queries
	logger *slog.Logger
}

// NewService creates a new graph service
func NewService(db *database.Queries, logger *slog.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// CreateNode creates a new node and returns its ID
func (s *Service) CreateNode(ctx context.Context, params CreateNodeParams) (uuid.UUID, error) {
	node, err := s.db.CreateNode(ctx, database.CreateNodeParams{
		NodeType:   params.NodeType,
		Label:      params.Label,
		Properties: database.JSONB(params.Properties),
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create node: %w", err)
	}

	id, err := uuid.FromBytes(node.ID.Bytes[:16])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse node ID: %w", err)
	}

	s.logger.Info("node created", "id", id, "type", params.NodeType, "label", params.Label)
	return id, nil
}

// GetNode retrieves a node by ID
func (s *Service) GetNode(ctx context.Context, id uuid.UUID) (*database.Node, error) {
	node, err := s.db.GetNode(ctx, database.PgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	return node, nil
}

// CreateEdge creates a new edge and returns its ID
func (s *Service) CreateEdge(ctx context.Context, params CreateEdgeParams) (uuid.UUID, error) {
	edge, err := s.db.CreateEdge(ctx, database.CreateEdgeParams{
		EdgeType:   params.EdgeType,
		SourceNode: database.PgUUID(params.SourceNode),
		TargetNode: database.PgUUID(params.TargetNode),
		Properties: database.JSONB(params.Properties),
		IsNegated:  params.IsNegated,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create edge: %w", err)
	}

	id, err := uuid.FromBytes(edge.ID.Bytes[:16])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse edge ID: %w", err)
	}

	s.logger.Info("edge created", "id", id, "type", params.EdgeType, "source", params.SourceNode, "target", params.TargetNode)
	return id, nil
}

// GetEdge retrieves an edge by ID
func (s *Service) GetEdge(ctx context.Context, id uuid.UUID) (*database.Edge, error) {
	edge, err := s.db.GetEdge(ctx, database.PgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get edge: %w", err)
	}
	return edge, nil
}

// CreateProvenance creates a new provenance record and returns its ID
func (s *Service) CreateProvenance(ctx context.Context, params CreateProvenanceParams) (uuid.UUID, error) {
	// Convert location to JSONB
	locationJSON, err := json.Marshal(params.Location)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal location: %w", err)
	}

	var claimedByUUID pgtype.UUID
	if params.ClaimedBy != nil {
		claimedByUUID = database.PgUUID(*params.ClaimedBy)
	}

	prov, err := s.db.CreateProvenance(ctx, database.CreateProvenanceParams{
		TargetType:       params.TargetType,
		TargetID:         database.PgUUID(params.TargetID),
		SourceID:         database.PgUUID(params.SourceID),
		Excerpt:          params.Excerpt,
		Location:         locationJSON,
		Confidence:       params.Confidence,
		Trust:            params.Trust,
		Status:           string(params.Status),
		Modality:         params.Modality,
		ClaimedTimeStart: database.PgTimePtr(params.ClaimedTimeStart),
		ClaimedTimeEnd:   database.PgTimePtr(params.ClaimedTimeEnd),
		ClaimedTimeText:  database.PgText(params.ClaimedTimeText),
		ClaimedGeoRegion: database.PgText(params.ClaimedGeoRegion),
		ClaimedGeoText:   database.PgText(params.ClaimedGeoText),
		ClaimedBy:        claimedByUUID,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create provenance: %w", err)
	}

	id, err := uuid.FromBytes(prov.ID.Bytes[:16])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse provenance ID: %w", err)
	}

	s.logger.Info("provenance created", "id", id, "target_type", params.TargetType, "target_id", params.TargetID)
	return id, nil
}

// GetNodeWithProvenance retrieves a node with all its provenance records
func (s *Service) GetNodeWithProvenance(ctx context.Context, id uuid.UUID) (*database.NodeWithProvenance, error) {
	node, err := s.db.GetNode(ctx, database.PgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	provenance, err := s.db.ListProvenanceByTarget(ctx, database.ListProvenanceByTargetParams{
		TargetType: "node",
		TargetID:   database.PgUUID(id),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get provenance: %w", err)
	}

	return &database.NodeWithProvenance{
		Node:       *node,
		Provenance: provenance,
	}, nil
}

// GetEdgeWithProvenance retrieves an edge with all its provenance records
func (s *Service) GetEdgeWithProvenance(ctx context.Context, id uuid.UUID) (*database.EdgeWithProvenance, error) {
	edge, err := s.db.GetEdge(ctx, database.PgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get edge: %w", err)
	}

	provenance, err := s.db.ListProvenanceByTarget(ctx, database.ListProvenanceByTargetParams{
		TargetType: "edge",
		TargetID:   database.PgUUID(id),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get provenance: %w", err)
	}

	return &database.EdgeWithProvenance{
		Edge:       *edge,
		Provenance: provenance,
	}, nil
}

// ListNodesByType lists all nodes of a given type
func (s *Service) ListNodesByType(ctx context.Context, nodeType string, limit int32) ([]database.Node, error) {
	nodesPtrs, err := s.db.ListNodesByType(ctx, database.ListNodesByTypeParams{
		NodeType: nodeType,
		Limit:    limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	// Convert pointer slice to value slice
	nodes := make([]database.Node, len(nodesPtrs))
	for i, n := range nodesPtrs {
		if n != nil {
			nodes[i] = *n
		}
	}
	return nodes, nil
}

// ListEdgesByType lists all edges of a given type
func (s *Service) ListEdgesByType(ctx context.Context, edgeType string, limit int32) ([]database.Edge, error) {
	edgesPtrs, err := s.db.ListEdgesByType(ctx, database.ListEdgesByTypeParams{
		EdgeType: edgeType,
		Limit:    limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list edges: %w", err)
	}
	// Convert pointer slice to value slice
	edges := make([]database.Edge, len(edgesPtrs))
	for i, e := range edgesPtrs {
		if e != nil {
			edges[i] = *e
		}
	}
	return edges, nil
}

// ListNodesBySource lists all nodes that have provenance from a specific source
func (s *Service) ListNodesBySource(ctx context.Context, sourceID uuid.UUID) ([]database.Node, error) {
	nodesPtrs, err := s.db.ListNodesBySource(ctx, database.PgUUID(sourceID))
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes by source: %w", err)
	}
	// Convert pointer slice to value slice
	nodes := make([]database.Node, len(nodesPtrs))
	for i, n := range nodesPtrs {
		if n != nil {
			nodes[i] = *n
		}
	}
	return nodes, nil
}

// UpdateProvenanceStatus updates the status of a provenance record
func (s *Service) UpdateProvenanceStatus(ctx context.Context, provenanceID uuid.UUID, status database.ReviewStatus) error {
	_, err := s.db.UpdateProvenanceStatus(ctx, database.UpdateProvenanceStatusParams{
		ID:     database.PgUUID(provenanceID),
		Status: string(status),
	})
	if err != nil {
		return fmt.Errorf("failed to update provenance status: %w", err)
	}

	s.logger.Info("provenance status updated", "id", provenanceID, "status", status)
	return nil
}

// DeleteNode deletes a node (cascades to edges and provenance)
func (s *Service) DeleteNode(ctx context.Context, id uuid.UUID) error {
	err := s.db.DeleteNode(ctx, database.PgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	s.logger.Info("node deleted", "id", id)
	return nil
}

// DeleteEdge deletes an edge (cascades to provenance)
func (s *Service) DeleteEdge(ctx context.Context, id uuid.UUID) error {
	err := s.db.DeleteEdge(ctx, database.PgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete edge: %w", err)
	}

	s.logger.Info("edge deleted", "id", id)
	return nil
}

// DeleteProvenance deletes a provenance record
func (s *Service) DeleteProvenance(ctx context.Context, id uuid.UUID) error {
	err := s.db.DeleteProvenance(ctx, database.PgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete provenance: %w", err)
	}

	s.logger.Info("provenance deleted", "id", id)
	return nil
}
