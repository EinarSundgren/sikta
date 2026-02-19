package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for sources and chunks.
type Repository struct {
	queries *Queries
	ctx     context.Context
}

// NewRepository creates a new repository.
func NewRepository(db *pgxpool.Pool, ctx context.Context) *Repository {
	return &Repository{
		queries: New(db),
		ctx:     ctx,
	}
}

// CreateSource creates a new source record.
func (r *Repository) CreateSource(title, filename, filePath, fileType, uploadStatus string, isDemo bool) (*Source, error) {
	return r.queries.CreateSource(r.ctx, CreateSourceParams{
		Title:        title,
		Filename:     filename,
		FilePath:     filePath,
		FileType:     fileType,
		UploadStatus: uploadStatus,
		IsDemo:       isDemo,
	})
}

// UpdateSourceStatus updates a source's upload status.
func (r *Repository) UpdateSourceStatus(id uuid.UUID, status string, errorMessage *string) (*Source, error) {
	var errMsg pgtype.Text
	if errorMessage != nil {
		errMsg = pgtype.Text{String: *errorMessage, Valid: true}
	}
	return r.queries.UpdateSourceStatus(r.ctx, UpdateSourceStatusParams{
		ID:           PgUUID(id),
		UploadStatus: status,
		ErrorMessage: errMsg,
	})
}

// UpdateSourceTotalPages updates a source's total page count.
func (r *Repository) UpdateSourceTotalPages(id uuid.UUID, totalPages int32) error {
	return r.queries.UpdateSourceTotalPages(r.ctx, UpdateSourceTotalPagesParams{
		ID:         PgUUID(id),
		TotalPages: pgtype.Int4{Int32: totalPages, Valid: true},
	})
}

// CreateChunk creates a new chunk record.
func (r *Repository) CreateChunk(sourceID uuid.UUID, chunkIndex int32, content string, chapterTitle *string, chapterNumber *int32, pageStart, pageEnd *int32, narrativePosition int32, wordCount *int32) (*CreateChunkRow, error) {
	params := CreateChunkParams{
		SourceID:          PgUUID(sourceID),
		ChunkIndex:        chunkIndex,
		Content:           content,
		ChapterTitle:      PgTextPtr(chapterTitle),
		NarrativePosition: narrativePosition,
	}
	if chapterNumber != nil {
		params.ChapterNumber = pgtype.Int4{Int32: *chapterNumber, Valid: true}
	}
	if pageStart != nil {
		params.PageStart = pgtype.Int4{Int32: *pageStart, Valid: true}
	}
	if pageEnd != nil {
		params.PageEnd = pgtype.Int4{Int32: *pageEnd, Valid: true}
	}
	if wordCount != nil {
		params.WordCount = pgtype.Int4{Int32: *wordCount, Valid: true}
	}
	return r.queries.CreateChunk(r.ctx, params)
}

// GetSource retrieves a source by ID.
func (r *Repository) GetSource(id uuid.UUID) (*Source, error) {
	return r.queries.GetSource(r.ctx, PgUUID(id))
}

// ListSources retrieves all sources.
func (r *Repository) ListSources() ([]*Source, error) {
	return r.queries.ListSources(r.ctx)
}

// GetChunks retrieves all chunks for a source.
func (r *Repository) GetChunks(sourceID uuid.UUID) ([]*Chunk, error) {
	return r.queries.ListChunksBySource(r.ctx, PgUUID(sourceID))
}

// GetChunkCount returns the number of chunks for a source.
func (r *Repository) GetChunkCount(sourceID uuid.UUID) (int64, error) {
	return r.queries.CountChunksBySource(r.ctx, PgUUID(sourceID))
}
