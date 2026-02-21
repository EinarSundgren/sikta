package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
	"github.com/einarsundgren/sikta/internal/graph"
	"github.com/google/uuid"
)

// GraphService handles extraction to the graph model
type GraphService struct {
	db     *database.Queries
	claude *claude.Client
	graph  *graph.Service
	logger *slog.Logger
	model  string
}

// NewGraphService creates a new graph extraction service
func NewGraphService(db *database.Queries, claude *claude.Client, graph *graph.Service, logger *slog.Logger, model string) *GraphService {
	return &GraphService{
		db:     db,
		claude: claude,
		graph:  graph,
		logger: logger,
		model:  model,
	}
}

// ExtractionProgress tracks extraction progress for graph extraction
type GraphExtractionProgress struct {
	DocumentID             string
	TotalChunks            int
	ProcessedChunks        int
	NodesExtracted        int
	EdgesExtracted        int
	CurrentChunk           int
	Status                 string
	Error                  string
}

// ProgressCallback is called with progress updates
type ProgressCallback func(progress GraphExtractionProgress)

// ExtractDocumentToGraph extracts nodes and edges from a document
func (s *GraphService) ExtractDocumentToGraph(ctx context.Context, sourceID string, progressCb ProgressCallback) error {
	s.logger.Info("starting graph extraction", "source_id", sourceID)

	// Get the document node
	docNodeID, err := s.getDocumentNode(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get document node: %w", err)
	}

	// Get chunks
	chunks, err := s.db.ListChunksBySource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return fmt.Errorf("failed to get chunks: %w", err)
	}

	totalChunks := len(chunks)
	s.logger.Info("processing chunks for graph extraction", "total", totalChunks)

	// Track entity labels for edge creation
	entityLabelToID := make(map[string]uuid.UUID)

	for i, chunk := range chunks {
		s.logger.Info("processing chunk for graph extraction", "index", i, "chapter", chunk.ChapterTitle.String)

		if progressCb != nil {
			progressCb(GraphExtractionProgress{
				DocumentID:      sourceID,
				TotalChunks:     totalChunks,
				ProcessedChunks: i,
				CurrentChunk:    i,
				Status:          "processing",
			})
		}

		// Extract nodes and edges from this chunk
		nodes, edges, err := s.extractFromChunk(ctx, chunk, docNodeID)
		if err != nil {
			s.logger.Error("failed to extract from chunk", "index", i, "error", err)
			continue
		}

		// Store nodes and track entity labels
		for _, node := range nodes {
			nodeID, err := s.storeExtractedNode(ctx, node, chunk, docNodeID)
			if err != nil {
				s.logger.Error("failed to store node", "label", node.Label, "error", err)
				continue
			}

			// Track entity labels for edge creation
			if node.NodeType == "person" || node.NodeType == "place" ||
			   node.NodeType == "organization" || node.NodeType == "object" {
				entityLabelToID[node.Label] = nodeID
			}
		}

		// Store edges (linking entity labels to node IDs)
		for _, edge := range edges {
			_, err := s.storeExtractedEdge(ctx, edge, chunk, docNodeID, entityLabelToID)
			if err != nil {
				s.logger.Error("failed to store edge", "type", edge.EdgeType, "error", err)
				continue
			}
		}

		if progressCb != nil {
			progressCb(GraphExtractionProgress{
				DocumentID:      sourceID,
				TotalChunks:     totalChunks,
				ProcessedChunks: i + 1,
				NodesExtracted:  len(entityLabelToID),
				EdgesExtracted:  len(edges),
				CurrentChunk:   i,
				Status:         "processing",
			})
		}

		time.Sleep(500 * time.Millisecond)
	}

	s.logger.Info("graph extraction complete", "source_id", sourceID)

	if progressCb != nil {
		progressCb(GraphExtractionProgress{
			DocumentID: sourceID,
			Status:     "complete",
		})
	}

	return nil
}

// getDocumentNode gets or creates the document node for a source
func (s *GraphService) getDocumentNode(ctx context.Context, sourceID string) (uuid.UUID, error) {
	// Try to find existing document node
	nodesPtrs, err := s.db.ListNodesByType(ctx, database.ListNodesByTypeParams{
		NodeType: "document",
		Limit:    1,
	})
	if err == nil && len(nodesPtrs) > 0 {
		// Check if this node belongs to our source
		for _, node := range nodesPtrs {
			provenance, _ := s.db.ListProvenanceByTarget(ctx, database.ListProvenanceByTargetParams{
				TargetType: "node",
				TargetID:   node.ID,
			})
			for _, prov := range provenance {
				if database.UUIDStr(prov.SourceID) == sourceID {
					id, _ := uuid.FromBytes(node.ID.Bytes[:16])
					return id, nil
				}
			}
		}
	}

	// Get source info to create document node
	source, err := s.db.GetSource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return uuid.Nil, err
	}

	// Create document node
	return s.graph.CreateNode(ctx, graph.CreateNodeParams{
		NodeType: database.NodeTypeDocument,
		Label:    source.Title,
		Properties: map[string]interface{}{
			"filename":  source.Filename,
			"file_type": source.FileType,
		},
	})
}

// extractFromChunk extracts nodes and edges from a single chunk
func (s *GraphService) extractFromChunk(ctx context.Context, chunk *database.Chunk, docNodeID uuid.UUID) ([]ExtractedNode, []ExtractedEdge, error) {
	userMessage := fmt.Sprintf("%s\n\n%s", GraphFewShotExample, chunk.Content)

	apiResp, err := s.claude.SendSystemPrompt(ctx, GraphExtractionSystemPrompt, userMessage, s.model)
	if err != nil {
		return nil, nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, nil, fmt.Errorf("empty response from Claude")
	}

	responseText := stripMarkdownCodeBlocks(apiResp.Content[0].Text)

	// Fallback: try to extract JSON from markdown
	var resp GraphExtractionResponse
	if err := json.Unmarshal([]byte(responseText), &resp); err != nil {
		if jsonStr := extractJSONFromMarkdown(responseText); jsonStr != "" {
			responseText = jsonStr
			if err := json.Unmarshal([]byte(responseText), &resp); err != nil {
				return nil, nil, fmt.Errorf("failed to parse JSON response: %w", err)
			}
		} else {
			return nil, nil, fmt.Errorf("failed to parse JSON response: %w", err)
		}
	}

	s.logger.Info("extracted from chunk",
		"nodes", len(resp.Nodes),
		"edges", len(resp.Edges))

	return resp.Nodes, resp.Edges, nil
}

// storeExtractedNode stores an extracted node with provenance
func (s *GraphService) storeExtractedNode(ctx context.Context, node ExtractedNode, chunk *database.Chunk, docNodeID uuid.UUID) (uuid.UUID, error) {
	// Create the node
	nodeID, err := s.graph.CreateNode(ctx, graph.CreateNodeParams{
		NodeType:   node.NodeType,
		Label:      node.Label,
		Properties: node.Properties,
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Build location
	location := database.Location{
		Chapter: chunk.ChapterTitle.String,
	}
	if chunk.PageStart.Valid {
		location.Page = int(chunk.PageStart.Int32)
	}

	// Determine modality
	modality := database.ModalityAsserted
	if node.Modality != "" {
		modality = node.Modality
	}

	// Create provenance
	_, err = s.graph.CreateProvenance(ctx, graph.CreateProvenanceParams{
		TargetType:       "node",
		TargetID:         nodeID,
		SourceID:         docNodeID,
		Excerpt:          node.Excerpt,
		Location:         location,
		Confidence:       float32(node.Confidence),
		Trust:            1.0, // Would come from source trust
		Modality:         modality,
		Status:           database.StatusPending,
		ClaimedTimeStart: node.ClaimedTimeStart,
		ClaimedTimeEnd:   node.ClaimedTimeEnd,
		ClaimedTimeText:  node.ClaimedTimeText,
		ClaimedGeoRegion: node.ClaimedGeoRegion,
		ClaimedGeoText:   node.ClaimedGeoText,
	})

	return nodeID, nil
}

// storeExtractedEdge stores an extracted edge with provenance
func (s *GraphService) storeExtractedEdge(ctx context.Context, edge ExtractedEdge, chunk *database.Chunk, docNodeID uuid.UUID, entityLabelToID map[string]uuid.UUID) (uuid.UUID, error) {
	// Look up source node ID by label
	sourceID, ok := entityLabelToID[edge.SourceNode]
	if !ok {
		return uuid.Nil, fmt.Errorf("source node not found: %s", edge.SourceNode)
	}

	// Look up target node ID by label
	targetID, ok := entityLabelToID[edge.TargetNode]
	if !ok {
		return uuid.Nil, fmt.Errorf("target node not found: %s", edge.TargetNode)
	}

	// Determine modality
	modality := database.ModalityAsserted
	if edge.Modality != "" {
		modality = edge.Modality
	}

	// Create the edge
	edgeID, err := s.graph.CreateEdge(ctx, graph.CreateEdgeParams{
		EdgeType:   edge.EdgeType,
		SourceNode: sourceID,
		TargetNode: targetID,
		Properties: edge.Properties,
		IsNegated:  edge.IsNegated,
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Build location
	location := database.Location{
		Chapter: chunk.ChapterTitle.String,
	}

	// Create provenance for the edge
	_, err = s.graph.CreateProvenance(ctx, graph.CreateProvenanceParams{
		TargetType: "edge",
		TargetID:   edgeID,
		SourceID:   docNodeID,
		Excerpt:    edge.Excerpt,
		Location:   location,
		Confidence: float32(edge.Confidence),
		Trust:      1.0,
		Modality:   modality,
		Status:     database.StatusPending,
	})

	return edgeID, nil
}
