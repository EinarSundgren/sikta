package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// PostProcessor handles cross-document post-processing
type PostProcessor struct {
	db         *database.Queries
	graph      *Service
	claude     *claude.Client
	logger     *slog.Logger
	model      string
	promptDir  string
}

// NewPostProcessor creates a new post-processor
func NewPostProcessor(db *database.Queries, graph *Service, claude *claude.Client, logger *slog.Logger, model, promptDir string) *PostProcessor {
	return &PostProcessor{
		db:        db,
		graph:     graph,
		claude:    claude,
		logger:    logger,
		model:     model,
		promptDir: promptDir,
	}
}

// DeduplicationResult represents the LLM response for entity deduplication
type DeduplicationResult struct {
	Matches []EntityMatch `json:"matches"`
}

// EntityMatch represents a matched entity group
type EntityMatch struct {
	Canonical  string   `json:"canonical"`
	Aliases    []string `json:"aliases"`
	EntityType string   `json:"entity_type"`
	Confidence float32  `json:"confidence"`
}

// RunDeduplication runs entity deduplication across all documents in a project
func (p *PostProcessor) RunDeduplication(ctx context.Context, projectID uuid.UUID) (*DeduplicationResult, error) {
	p.logger.Info("starting entity deduplication", "project_id", projectID)

	// Get all sources for the project
	sources, err := p.db.GetProjectSources(ctx, database.PgUUID(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get project sources: %w", err)
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources found for project")
	}

	// Collect all entity nodes from all sources
	var allEntities []EntityInfo
	entityMap := make(map[string]uuid.UUID) // label -> node ID

	for _, source := range sources {
		nodes, err := p.db.ListNodesBySource(ctx, source.ID)
		if err != nil {
			p.logger.Warn("failed to get nodes for source", "source_id", source.ID, "error", err)
			continue
		}

		for _, node := range nodes {
			// Only include entity types (person, organization, place)
			if node.NodeType == "person" || node.NodeType == "organization" || node.NodeType == "place" {
				allEntities = append(allEntities, EntityInfo{
					ID:         pgUUIDToStr(node.ID),
					Label:      node.Label,
					EntityType: node.NodeType,
					SourceID:   pgUUIDToStr(source.ID),
				})
				entityMap[node.Label] = uuid.MustParse(pgUUIDToStr(node.ID))
			}
		}
	}

	p.logger.Info("collected entities for deduplication", "count", len(allEntities))

	if len(allEntities) < 2 {
		p.logger.Info("not enough entities for deduplication")
		return &DeduplicationResult{Matches: []EntityMatch{}}, nil
	}

	// Load deduplication prompt
	prompt, err := p.loadPrompt("dedup")
	if err != nil {
		return nil, fmt.Errorf("failed to load dedup prompt: %w", err)
	}

	// Build entities JSON for the prompt
	entitiesJSON, err := json.MarshalIndent(allEntities, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entities: %w", err)
	}

	userMessage := fmt.Sprintf("Here are the entities extracted from the project documents:\n\n%s", string(entitiesJSON))

	// Call LLM
	apiResp, err := p.claude.SendSystemPrompt(ctx, prompt, userMessage, p.model)
	if err != nil {
		return nil, fmt.Errorf("deduplication API call failed: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from deduplication")
	}

	responseText := apiResp.Content[0].Text

	// Parse response
	var result DeduplicationResult
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		// Try stripping markdown
		responseText = stripMarkdownCodeBlocks(responseText)
		if err := json.Unmarshal([]byte(responseText), &result); err != nil {
			return nil, fmt.Errorf("failed to parse deduplication response: %w", err)
		}
	}

	// Create same_as edges for matched entities
	for _, match := range result.Matches {
		if err := p.createSameAsEdges(ctx, match, entityMap); err != nil {
			p.logger.Warn("failed to create same_as edges", "canonical", match.Canonical, "error", err)
		}
	}

	p.logger.Info("deduplication complete", "matches", len(result.Matches))
	return &result, nil
}

// createSameAsEdges creates same_as edges between matched entities
func (p *PostProcessor) createSameAsEdges(ctx context.Context, match EntityMatch, entityMap map[string]uuid.UUID) error {
	// Find canonical node ID
	canonicalID, ok := entityMap[match.Canonical]
	if !ok {
		return fmt.Errorf("canonical entity not found: %s", match.Canonical)
	}

	// Create edges from canonical to each alias
	for _, alias := range match.Aliases {
		if alias == match.Canonical {
			continue
		}

		aliasID, ok := entityMap[alias]
		if !ok {
			p.logger.Debug("alias entity not found, skipping", "alias", alias)
			continue
		}

		// Create same_as edge
		properties := map[string]interface{}{
			"canonical":  match.Canonical,
			"alias":      alias,
			"confidence": match.Confidence,
		}

		_, err := p.graph.CreateEdge(ctx, CreateEdgeParams{
			EdgeType:   "same_as",
			SourceNode: canonicalID,
			TargetNode: aliasID,
			Properties: properties,
		})

		if err != nil {
			p.logger.Warn("failed to create same_as edge", "canonical", match.Canonical, "alias", alias, "error", err)
		} else {
			p.logger.Info("created same_as edge", "canonical", match.Canonical, "alias", alias)
		}
	}

	return nil
}

// EntityInfo represents an entity for deduplication
type EntityInfo struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	EntityType string `json:"entity_type"`
	SourceID   string `json:"source_id"`
}

// InconsistencyResult represents the LLM response for inconsistency detection
type InconsistencyResult struct {
	Inconsistencies []DetectedInconsistency `json:"inconsistencies"`
}

// DetectedInconsistency represents a detected inconsistency
type DetectedInconsistency struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Severity         string                 `json:"severity"`
	Description      string                 `json:"description"`
	Documents        []string               `json:"documents"`
	EntitiesInvolved []string               `json:"entities_involved"`
	Evidence         InconsistencyEvidence  `json:"evidence"`
	Confidence       float32                `json:"confidence"`
}

// InconsistencyEvidence contains evidence from both sides
type InconsistencyEvidence struct {
	SideA InconsistencySide `json:"side_a"`
	SideB InconsistencySide `json:"side_b"`
}

// InconsistencySide represents one side of an inconsistency
type InconsistencySide struct {
	Doc     string `json:"doc"`
	Section string `json:"section"`
	Claim   string `json:"claim"`
}

// RunInconsistencyDetection runs inconsistency detection across all documents in a project
func (p *PostProcessor) RunInconsistencyDetection(ctx context.Context, projectID uuid.UUID) (*InconsistencyResult, error) {
	p.logger.Info("starting inconsistency detection", "project_id", projectID)

	// Get all sources for the project
	sources, err := p.db.GetProjectSources(ctx, database.PgUUID(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get project sources: %w", err)
	}

	if len(sources) < 2 {
		p.logger.Info("not enough documents for inconsistency detection")
		return &InconsistencyResult{Inconsistencies: []DetectedInconsistency{}}, nil
	}

	// Collect all event nodes from all sources
	var allEvents []EventInfo
	eventMap := make(map[string]uuid.UUID) // label -> node ID

	for _, source := range sources {
		nodes, err := p.db.ListNodesBySource(ctx, source.ID)
		if err != nil {
			p.logger.Warn("failed to get nodes for source", "source_id", source.ID, "error", err)
			continue
		}

		for _, node := range nodes {
			if node.NodeType == "event" {
				allEvents = append(allEvents, EventInfo{
					ID:         pgUUIDToStr(node.ID),
					Label:      node.Label,
					SourceType: node.NodeType,
					SourceID:   pgUUIDToStr(source.ID),
				})
				eventMap[node.Label] = uuid.MustParse(pgUUIDToStr(node.ID))
			}
		}
	}

	p.logger.Info("collected events for inconsistency detection", "count", len(allEvents))

	if len(allEvents) < 2 {
		p.logger.Info("not enough events for inconsistency detection")
		return &InconsistencyResult{Inconsistencies: []DetectedInconsistency{}}, nil
	}

	// Load inconsistency prompt
	prompt, err := p.loadPrompt("inconsistency")
	if err != nil {
		return nil, fmt.Errorf("failed to load inconsistency prompt: %w", err)
	}

	// Build events JSON for the prompt
	eventsJSON, err := json.MarshalIndent(allEvents, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	userMessage := fmt.Sprintf("Here are the extracted events from all documents:\n\n%s", string(eventsJSON))

	// Call LLM
	apiResp, err := p.claude.SendSystemPrompt(ctx, prompt, userMessage, p.model)
	if err != nil {
		return nil, fmt.Errorf("inconsistency detection API call failed: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from inconsistency detection")
	}

	responseText := apiResp.Content[0].Text

	// Parse response
	var result InconsistencyResult
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		responseText = stripMarkdownCodeBlocks(responseText)
		if err := json.Unmarshal([]byte(responseText), &result); err != nil {
			return nil, fmt.Errorf("failed to parse inconsistency response: %w", err)
		}
	}

	// Create contradicts edges for detected inconsistencies
	for _, inc := range result.Inconsistencies {
		if err := p.createContradictsEdge(ctx, inc, eventMap); err != nil {
			p.logger.Warn("failed to create contradicts edge", "id", inc.ID, "error", err)
		}
	}

	p.logger.Info("inconsistency detection complete", "inconsistencies", len(result.Inconsistencies))
	return &result, nil
}

// createContradictsEdge creates a contradicts edge for an inconsistency
func (p *PostProcessor) createContradictsEdge(ctx context.Context, inc DetectedInconsistency, eventMap map[string]uuid.UUID) error {
	// Find node IDs for the claims involved
	// This is a simplified approach - in reality we'd need to match claims more precisely
	var sourceID, targetID uuid.UUID
	var found bool

	// Try to find matching events by description
	for label, id := range eventMap {
		if !found && (containsIgnoreCase(label, inc.Evidence.SideA.Claim) || containsIgnoreCase(label, inc.Evidence.SideB.Claim)) {
			if sourceID == uuid.Nil {
				sourceID = id
			} else {
				targetID = id
				found = true
				break
			}
		}
	}

	if !found {
		p.logger.Debug("could not find matching events for inconsistency", "id", inc.ID)
		return nil
	}

	properties := map[string]interface{}{
		"type":        inc.Type,
		"severity":    inc.Severity,
		"description": inc.Description,
		"documents":   inc.Documents,
		"confidence":  inc.Confidence,
		"evidence":    inc.Evidence,
	}

	_, err := p.graph.CreateEdge(ctx, CreateEdgeParams{
		EdgeType:   "contradicts",
		SourceNode: sourceID,
		TargetNode: targetID,
		Properties: properties,
	})

	if err != nil {
		return fmt.Errorf("failed to create contradicts edge: %w", err)
	}

	p.logger.Info("created contradicts edge", "id", inc.ID, "type", inc.Type)
	return nil
}

// EventInfo represents an event for inconsistency detection
type EventInfo struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	SourceType string `json:"source_type"`
	SourceID   string `json:"source_id"`
}

// loadPrompt loads a post-processing prompt by name
func (p *PostProcessor) loadPrompt(name string) (string, error) {
	if p.promptDir == "" {
		return "", fmt.Errorf("prompt directory not configured")
	}

	path := p.promptDir + "/postprocess/" + name + ".txt"
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt %s: %w", path, err)
	}

	return string(content), nil
}

// Helper functions

func pgUUIDToStr(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	uuid, _ := uuid.FromBytes(id.Bytes[:])
	return uuid.String()
}

func stripMarkdownCodeBlocks(s string) string {
	// Remove ```json and ``` markers
	result := s
	if len(result) > 7 && result[:7] == "```json" {
		result = result[7:]
	} else if len(result) > 3 && result[:3] == "```" {
		result = result[3:]
	}
	if len(result) > 3 && result[len(result)-4:] == "```\n" {
		result = result[:len(result)-4]
	} else if len(result) > 3 && result[len(result)-3:] == "```" {
		result = result[:len(result)-3]
	}
	return result
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsIgnoreCase(s[1:], substr) ||
		(len(s) > 0 && len(substr) > 0 && (s[0]|32) == (substr[0]|32) && containsIgnoreCase(s[1:], substr[1:])))
}
