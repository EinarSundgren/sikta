package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for documents and chunks.
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

// CreateDocument creates a new document record.
func (r *Repository) CreateDocument(title, filename, filePath, fileType, uploadStatus string, isDemo bool) (Document, error) {
	id := uuid.New()
	return r.queries.CreateDocument(r.ctx, CreateDocumentParams{
		ID:           id,
		Title:        title,
		Filename:     filename,
		FilePath:     filePath,
		FileType:     fileType,
		UploadStatus: uploadStatus,
		IsDemo:       isDemo,
	})
}

// UpdateDocumentStatus updates a document's status.
func (r *Repository) UpdateDocumentStatus(id uuid.UUID, status string, errorMessage *string) (Document, error) {
	return r.queries.UpdateDocumentStatus(r.ctx, UpdateDocumentStatusParams{
		ID:           id,
		UploadStatus: status,
		ErrorMessage: errorMessage,
	})
}

// UpdateDocumentTotalPages updates a document's total page count.
func (r *Repository) UpdateDocumentTotalPages(id uuid.UUID, totalPages int32) error {
	return r.queries.UpdateDocumentTotalPages(r.ctx, UpdateDocumentTotalPagesParams{
		ID:         id,
		TotalPages: totalPages,
	})
}

// CreateChunk creates a new chunk record.
func (r *Repository) CreateChunk(documentID uuid.UUID, chunkIndex int32, content string, chapterTitle *string, chapterNumber *int32, pageStart, pageEnd *int32, narrativePosition int32, wordCount *int32) (Chunk, error) {
	return r.queries.CreateChunk(r.ctx, CreateChunkParams{
		DocumentID:        documentID,
		ChunkIndex:        chunkIndex,
		Content:           content,
		ChapterTitle:      chapterTitle,
		ChapterNumber:     chapterNumber,
		PageStart:         pageStart,
		PageEnd:           pageEnd,
		NarrativePosition: narrativePosition,
		WordCount:         wordCount,
	})
}

// GetDocument retrieves a document by ID.
func (r *Repository) GetDocument(id uuid.UUID) (Document, error) {
	return r.queries.GetDocument(r.ctx, id)
}

// ListDocuments retrieves all documents.
func (r *Repository) ListDocuments() ([]Document, error) {
	return r.queries.ListDocuments(r.ctx)
}

// GetChunks retrieves all chunks for a document.
func (r *Repository) GetChunks(documentID uuid.UUID) ([]Chunk, error) {
	return r.queries.ListChunksByDocument(r.ctx, documentID)
}

// GetChunkCount returns the number of chunks for a document.
func (r *Repository) GetChunkCount(documentID uuid.UUID) (int64, error) {
	return r.queries.CountChunksByDocument(r.ctx, documentID)
}
