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
	"github.com/jackc/pgx/v5/pgtype"
)

// Service handles the extraction pipeline.
type Service struct {
	db     *database.Queries
	claude *claude.Client
	logger *slog.Logger
	model  string
}

// NewService creates a new extraction service.
func NewService(db *database.Queries, claude *claude.Client, logger *slog.Logger, model string) *Service {
	return &Service{
		db:     db,
		claude: claude,
		logger: logger,
		model:  model,
	}
}

// ExtractionProgress tracks extraction progress.
type ExtractionProgress struct {
	DocumentID             string
	TotalChunks            int
	ProcessedChunks        int
	EventsExtracted        int
	EntitiesExtracted      int
	RelationshipsExtracted int
	CurrentChunk           int
	Status                 string
	Error                  string
}

// ProgressCallback is called with progress updates.
type ProgressCallback func(ExtractionProgress)

// GetChunkCount returns the number of chunks for a document.
func (s *Service) GetChunkCount(ctx context.Context, sourceID string) (int, error) {
	chunks, err := s.db.ListChunksBySource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return 0, fmt.Errorf("failed to get chunks: %w", err)
	}
	return len(chunks), nil
}

// ExtractDocument extracts events, entities, and relationships from a document.
func (s *Service) ExtractDocument(ctx context.Context, sourceID string, progressCb ProgressCallback) error {
	s.logger.Info("starting extraction", "source_id", sourceID)

	chunks, err := s.db.ListChunksBySource(ctx, database.PgUUID(parseUUID(sourceID)))
	if err != nil {
		return fmt.Errorf("failed to get chunks: %w", err)
	}

	totalChunks := len(chunks)
	s.logger.Info("processing chunks", "total", totalChunks)

	for i, chunk := range chunks {
		s.logger.Info("processing chunk", "index", i, "chapter", chunk.ChapterTitle.String)

		if progressCb != nil {
			progressCb(ExtractionProgress{
				DocumentID:      sourceID,
				TotalChunks:     totalChunks,
				ProcessedChunks: i,
				CurrentChunk:    i,
				Status:          "processing",
			})
		}

		resp, err := s.extractFromChunk(ctx, chunk)
		if err != nil {
			s.logger.Error("failed to extract from chunk", "index", i, "error", err)
			continue
		}

		eventIDs, entityIDs, relationshipIDs, err := s.storeExtractions(ctx, chunk, resp)
		if err != nil {
			s.logger.Error("failed to store extractions", "index", i, "error", err)
			continue
		}

		if progressCb != nil {
			progressCb(ExtractionProgress{
				DocumentID:             sourceID,
				TotalChunks:            totalChunks,
				ProcessedChunks:        i + 1,
				EventsExtracted:        len(eventIDs),
				EntitiesExtracted:      len(entityIDs),
				RelationshipsExtracted: len(relationshipIDs),
				CurrentChunk:           i,
				Status:                 "processing",
			})
		}

		time.Sleep(500 * time.Millisecond)
	}

	s.logger.Info("extraction complete", "source_id", sourceID)

	if progressCb != nil {
		progressCb(ExtractionProgress{
			DocumentID: sourceID,
			Status:     "complete",
		})
	}

	return nil
}

// extractFromChunk extracts data from a single chunk.
func (s *Service) extractFromChunk(ctx context.Context, chunk *database.Chunk) (*ExtractionResponse, error) {
	userMessage := fmt.Sprintf("%s\n\n%s", FewShotExample1, chunk.Content)

	apiResp, err := s.claude.SendSystemPrompt(ctx, ExtractionSystemPrompt, userMessage, s.model)
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	responseText := stripMarkdownCodeBlocks(apiResp.Content[0].Text)

	var resp ExtractionResponse
	if err := json.Unmarshal([]byte(responseText), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	s.logger.Info("extracted from chunk",
		"events", len(resp.Events),
		"entities", len(resp.Entities),
		"relationships", len(resp.Relationships))

	return &resp, nil
}

// storeExtractions stores extracted data in the database.
func (s *Service) storeExtractions(ctx context.Context, chunk *database.Chunk, resp *ExtractionResponse) ([]string, []string, []string, error) {
	sourceID := chunk.SourceID
	var eventIDs []string
	var entityIDs []string
	var relationshipIDs []string

	entityNameIDMap := make(map[string]string)

	for _, entity := range resp.Entities {
		entityID, err := s.storeEntity(ctx, sourceID, chunk, entity)
		if err != nil {
			s.logger.Error("failed to store entity", "name", entity.Name, "error", err)
			continue
		}
		entityIDs = append(entityIDs, entityID)
		entityNameIDMap[entity.Name] = entityID
	}

	// Assign sequential narrative positions to events within this chunk
	baseNarrativePos := int(chunk.NarrativePosition)
	eventsPerChunk := 1000 // Allow up to 1000 events per chunk with sequential positions

	for i, event := range resp.Events {
		// Calculate narrative position: chunk position * eventsPerChunk + event index within chunk
		narrativePos := baseNarrativePos*eventsPerChunk + i
		eventID, err := s.storeEvent(ctx, sourceID, chunk, event, narrativePos)
		if err != nil {
			s.logger.Error("failed to store event", "title", event.Title, "error", err)
			continue
		}
		eventIDs = append(eventIDs, eventID)
	}

	for _, relationship := range resp.Relationships {
		relID, err := s.storeRelationship(ctx, sourceID, chunk, relationship, entityNameIDMap)
		if err != nil {
			s.logger.Error("failed to store relationship", "type", relationship.Type, "error", err)
			continue
		}
		relationshipIDs = append(relationshipIDs, relID)
	}

	return eventIDs, entityIDs, relationshipIDs, nil
}

// storeEntity stores a single entity.
func (s *Service) storeEntity(ctx context.Context, sourceID pgtype.UUID, chunk *database.Chunk, entity Entity) (string, error) {
	entities, err := s.db.ListEntitiesBySource(ctx, sourceID)
	if err == nil {
		for _, e := range entities {
			if strings.EqualFold(e.Name, entity.Name) {
				return database.UUIDStr(e.ID), nil
			}
		}
	}

	params := database.CreateEntityParams{
		SourceID:             sourceID,
		Name:                 entity.Name,
		EntityType:           entity.Type,
		Aliases:              entity.Aliases,
		Description:          pgtype.Text{},
		FirstAppearanceChunk: pgtype.Int4{Int32: chunk.ChunkIndex, Valid: true},
		LastAppearanceChunk:  pgtype.Int4{Int32: chunk.ChunkIndex, Valid: true},
		Confidence:           float32(entity.Confidence),
		Metadata:             []byte("{}"),
	}

	created, err := s.db.CreateEntity(ctx, params)
	if err != nil {
		return "", err
	}

	if entity.Excerpt != "" {
		_, _ = s.db.CreateSourceReference(ctx, database.CreateSourceReferenceParams{
			ChunkID:        chunk.ID,
			ClaimID:        pgtype.UUID{},
			EntityID:       created.ID,
			RelationshipID: pgtype.UUID{},
			Excerpt:        entity.Excerpt,
		})
	}

	return database.UUIDStr(created.ID), nil
}

// storeEvent stores a single event as a claim.
func (s *Service) storeEvent(ctx context.Context, sourceID pgtype.UUID, chunk *database.Chunk, event Event, narrativePos int) (string, error) {
	params := database.CreateClaimParams{
		SourceID:           sourceID,
		ClaimType:          "event",
		Title:              event.Title,
		Description:        database.PgText(event.Description),
		EventType:          database.PgText(event.Type),
		DateText:           database.PgText(event.DateText),
		NarrativePosition:  int32(narrativePos),
		ChronologicalPosition: pgtype.Int4{Int32: int32(narrativePos), Valid: true},
		Confidence:         float32(event.Confidence),
		Metadata:           []byte("{}"),
	}

	created, err := s.db.CreateClaim(ctx, params)
	if err != nil {
		return "", err
	}

	if event.Excerpt != "" {
		_, _ = s.db.CreateSourceReference(ctx, database.CreateSourceReferenceParams{
			ChunkID:        chunk.ID,
			ClaimID:        created.ID,
			EntityID:       pgtype.UUID{},
			RelationshipID: pgtype.UUID{},
			Excerpt:        event.Excerpt,
		})
	}

	return database.UUIDStr(created.ID), nil
}

// storeRelationship stores a single relationship.
func (s *Service) storeRelationship(ctx context.Context, sourceID pgtype.UUID, chunk *database.Chunk, relationship Relationship, entityMap map[string]string) (string, error) {
	entityAID, ok := entityMap[relationship.EntityA]
	if !ok {
		return "", fmt.Errorf("entity A not found: %s", relationship.EntityA)
	}

	entityBID, ok := entityMap[relationship.EntityB]
	if !ok {
		return "", fmt.Errorf("entity B not found: %s", relationship.EntityB)
	}

	params := database.CreateRelationshipParams{
		SourceID:         sourceID,
		EntityAID:        database.PgUUID(parseUUID(entityAID)),
		EntityBID:        database.PgUUID(parseUUID(entityBID)),
		RelationshipType: relationship.Type,
		Description:      database.PgText(relationship.Description),
		Confidence:       float32(relationship.Confidence),
		Metadata:         []byte("{}"),
	}

	created, err := s.db.CreateRelationship(ctx, params)
	if err != nil {
		return "", err
	}

	return database.UUIDStr(created.ID), nil
}

// stripMarkdownCodeBlocks removes markdown code block formatting from text.
func stripMarkdownCodeBlocks(text string) string {
	trimmed := strings.TrimSpace(text)

	if strings.HasPrefix(trimmed, "```") {
		firstNewline := strings.Index(trimmed, "\n")
		if firstNewline != -1 {
			closingIndex := strings.LastIndex(trimmed, "```")
			if closingIndex > firstNewline {
				trimmed = strings.TrimSpace(trimmed[firstNewline+1 : closingIndex])
			} else {
				trimmed = strings.TrimSpace(trimmed[firstNewline+1:])
			}
		}
	}

	return trimmed
}
