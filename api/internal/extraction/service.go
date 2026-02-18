package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/einarsundgren/sikta/internal/extraction/claude"
)

// Service handles the extraction pipeline.
type Service struct {
	db      *database.Queries
	 claude  *claude.Client
	logger  *slog.Logger
}

// NewService creates a new extraction service.
func NewService(db *database.Queries, claude *claude.Client, logger *slog.Logger) *Service {
	return &Service{
		db:     db,
		claude: claude,
		logger: logger,
	}
}

// ExtractionProgress tracks extraction progress.
type ExtractionProgress struct {
	DocumentID      string
	TotalChunks     int
	ProcessedChunks int
	EventsExtracted int
	EntitiesExtracted int
	RelationshipsExtracted int
	CurrentChunk    int
	Status          string
	Error           string
}

// ProgressCallback is called with progress updates.
type ProgressCallback func(ExtractionProgress)

// ExtractDocument extracts events, entities, and relationships from a document.
func (s *Service) ExtractDocument(ctx context.Context, documentID string, progressCb ProgressCallback) error {
	s.logger.Info("starting extraction", "document_id", documentID)

	// Get all chunks for the document
	chunks, err := s.db.ListChunksByDocument(ctx, database.UUID(parseUUID(documentID)))
	if err != nil {
		return fmt.Errorf("failed to get chunks: %w", err)
	}

	totalChunks := len(chunks)
	s.logger.Info("processing chunks", "total", totalChunks)

	// Process each chunk
	for i, chunk := range chunks {
		s.logger.Info("processing chunk", "index", i, "chapter", chunk.ChapterTitle)

		// Report progress
		if progressCb != nil {
			progressCb(ExtractionProgress{
				DocumentID:      documentID,
				TotalChunks:     totalChunks,
				ProcessedChunks: i,
				CurrentChunk:    i,
				Status:          "processing",
			})
		}

		// Extract from this chunk
		resp, err := s.extractFromChunk(ctx, chunk)
		if err != nil {
			s.logger.Error("failed to extract from chunk", "index", i, "error", err)
			// Continue with next chunk instead of failing entirely
			continue
		}

		// Store extractions
		eventIDs, entityIDs, relationshipIDs, err := s.storeExtractions(ctx, chunk, resp)
		if err != nil {
			s.logger.Error("failed to store extractions", "index", i, "error", err)
			continue
		}

		// Report updated progress
		if progressCb != nil {
			progressCb(ExtractionProgress{
				DocumentID:           documentID,
				TotalChunks:          totalChunks,
				ProcessedChunks:      i + 1,
				EventsExtracted:      len(eventIDs),
				EntitiesExtracted:    len(entityIDs),
				RelationshipsExtracted: len(relationshipIDs),
				CurrentChunk:         i,
				Status:               "processing",
			})
		}

		// Small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	s.logger.Info("extraction complete", "document_id", documentID)

	if progressCb != nil {
		progressCb(ExtractionProgress{
			DocumentID: documentID,
			Status:     "complete",
		})
	}

	return nil
}

// extractFromChunk extracts data from a single chunk.
func (s *Service) extractFromChunk(ctx context.Context, chunk database.Chunk) (*ExtractionResponse, error) {
	// Build prompt with few-shot example
	userMessage := fmt.Sprintf("%s\n\n%s", FewShotExample1, chunk.Content)

	// Call Claude API
	apiResp, err := s.claude.SendSystemPrompt(ctx, ExtractionSystemPrompt, userMessage, "claude-sonnet-4-20250514")
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	// Parse JSON response
	var resp ExtractionResponse
	if err := json.Unmarshal([]byte(apiResp.Content[0].Text), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	s.logger.Debug("extraction successful",
		"events", len(resp.Events),
		"entities", len(resp.Entities),
		"relationships", len(resp.Relationships))

	return &resp, nil
}

// storeExtractions stores extracted data in the database.
func (s *Service) storeExtractions(ctx context.Context, chunk database.Chunk, resp *ExtractionResponse) ([]string, []string, []string, error) {
	documentID := chunk.DocumentID
	var eventIDs []string
	var entityIDs []string
	var relationshipIDs []string

	// Create entity name -> ID mapping for relationships
	entityNameIDMap := make(map[string]string)

	// Store entities first (relationships depend on them)
	for _, entity := range resp.Entities {
		entityID, err := s.storeEntity(ctx, documentID, chunk, entity)
		if err != nil {
			s.logger.Error("failed to store entity", "name", entity.Name, "error", err)
			continue
		}
		entityIDs = append(entityIDs, entityID)
		entityNameIDMap[entity.Name] = entityID
	}

	// Store events
	for _, event := range resp.Events {
		eventID, err := s.storeEvent(ctx, documentID, chunk, event)
		if err != nil {
			s.logger.Error("failed to store event", "title", event.Title, "error", err)
			continue
		}
		eventIDs = append(eventIDs, eventID)
	}

	// Store relationships
	for _, relationship := range resp.Relationships {
		relID, err := s.storeRelationship(ctx, documentID, chunk, relationship, entityNameIDMap)
		if err != nil {
			s.logger.Error("failed to store relationship", "type", relationship.Type, "error", err)
			continue
		}
		relationshipIDs = append(relationshipIDs, relID)
	}

	return eventIDs, entityIDs, relationshipIDs, nil
}

// storeEntity stores a single entity.
func (s *Service) storeEntity(ctx context.Context, documentID database.UUID, chunk database.Chunk, entity Entity) (string, error) {
	// Check if entity already exists (by name and document)
	entities, err := s.db.ListEntitiesByDocument(ctx, documentID)
	if err == nil {
		for _, e := range entities {
			if strings.EqualFold(e.Name, entity.Name) {
				// Entity exists, update it
				return e.ID.String(), nil
			}
		}
	}

	// Create new entity
	params := database.CreateEntityParams{
		DocumentID:          documentID,
		Name:                entity.Name,
		EntityType:          entity.Type,
		Aliases:             entity.Aliases,
		Description:         &entity.Description,
		FirstAppearanceChunk: &chunk.ChunkIndex,
		LastAppearanceChunk:  &chunk.ChunkIndex,
		Confidence:          entity.Confidence,
	}

	created, err := s.db.CreateEntity(ctx, params)
	if err != nil {
		return "", err
	}

	// Create source reference
	if entity.Excerpt != "" {
		_, _ = s.db.CreateSourceReference(ctx, database.CreateSourceReferenceParams{
			ChunkID:  chunk.ID,
			EntityID: &created.ID,
			Excerpt:  entity.Excerpt,
		})
	}

	return created.ID.String(), nil
}

// storeEvent stores a single event.
func (s *Service) storeEvent(ctx context.Context, documentID database.UUID, chunk database.Chunk, event Event) (string, error) {
	params := database.CreateEventParams{
		DocumentID:         documentID,
		Title:              event.Title,
		Description:        &event.Description,
		EventType:          event.Type,
		DateText:           &event.DateText,
		NarrativePosition:  chunk.NarrativePosition,
		Confidence:         event.Confidence,
	}

	created, err := s.db.CreateEvent(ctx, params)
	if err != nil {
		return "", err
	}

	// Create source reference
	if event.Excerpt != "" {
		_, _ = s.db.CreateSourceReference(ctx, database.CreateSourceReferenceParams{
			ChunkID:  chunk.ID,
			EventID:  &created.ID,
			Excerpt:  event.Excerpt,
		})
	}

	return created.ID.String(), nil
}

// storeRelationship stores a single relationship.
func (s *Service) storeRelationship(ctx context.Context, documentID database.UUID, chunk database.Chunk, relationship Relationship, entityMap map[string]string) (string, error) {
	// Look up entity IDs
	entityAID, ok := entityMap[relationship.EntityA]
	if !ok {
		return "", fmt.Errorf("entity A not found: %s", relationship.EntityA)
	}

	entityBID, ok := entityMap[relationship.EntityB]
	if !ok {
		return "", fmt.Errorf("entity B not found: %s", relationship.EntityB)
	}

	params := database.CreateRelationshipParams{
		DocumentID:        documentID,
		EntityAID:         parseUUID(entityAID),
		EntityBID:         parseUUID(entityBID),
		RelationshipType:  relationship.Type,
		Description:       &relationship.Description,
		Confidence:        relationship.Confidence,
	}

	created, err := s.db.CreateRelationship(ctx, params)
	if err != nil {
		return "", err
	}

	return created.ID.String(), nil
}

// parseUUID converts a string to database.UUID.
func parseUUID(s string) database.UUID {
	return database.UUID(s)
}
