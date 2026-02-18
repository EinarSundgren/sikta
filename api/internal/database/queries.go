package database

import (
	"context"

	"github.com/google/uuid"
)

// UUID wraps google uuid for use in database types.
type UUID = uuid.UUID

// CreateDocument creates a new document.
func (q *Queries) CreateDocument(ctx context.Context, arg CreateDocumentParams) (Document, error) {
	sql := `INSERT INTO documents (id, title, filename, file_path, file_type, upload_status, is_demo)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, filename, file_path, file_type, total_pages, upload_status, error_message, is_demo, metadata, created_at, updated_at`
	row := q.db.QueryRow(ctx, sql,
		arg.ID,
		arg.Title,
		arg.Filename,
		arg.FilePath,
		arg.FileType,
		arg.UploadStatus,
		arg.IsDemo,
	)
	var doc Document
	err := row.Scan(
		&doc.ID,
		&doc.Title,
		&doc.Filename,
		&doc.FilePath,
		&doc.FileType,
		&doc.TotalPages,
		&doc.UploadStatus,
		&doc.ErrorMessage,
		&doc.IsDemo,
		&doc.Metadata,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	return doc, err
}

// GetDocument retrieves a document by ID.
func (q *Queries) GetDocument(ctx context.Context, id UUID) (Document, error) {
	sql := `SELECT id, title, filename, file_path, file_type, total_pages, upload_status, error_message, is_demo, metadata, created_at, updated_at
		FROM documents WHERE id = $1`
	row := q.db.QueryRow(ctx, sql, id)
	var doc Document
	err := row.Scan(
		&doc.ID,
		&doc.Title,
		&doc.Filename,
		&doc.FilePath,
		&doc.FileType,
		&doc.TotalPages,
		&doc.UploadStatus,
		&doc.ErrorMessage,
		&doc.IsDemo,
		&doc.Metadata,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	return doc, err
}

// ListDocuments retrieves all documents.
func (q *Queries) ListDocuments(ctx context.Context) ([]Document, error) {
	sql := `SELECT id, title, filename, file_path, file_type, total_pages, upload_status, error_message, is_demo, metadata, created_at, updated_at
		FROM documents ORDER BY created_at DESC`
	rows, _ := q.db.Query(ctx, sql)
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(
			&doc.ID,
			&doc.Title,
			&doc.Filename,
			&doc.FilePath,
			&doc.FileType,
			&doc.TotalPages,
			&doc.UploadStatus,
			&doc.ErrorMessage,
			&doc.IsDemo,
			&doc.Metadata,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, rows.Err()
}

// UpdateDocumentStatus updates a document's status.
func (q *Queries) UpdateDocumentStatus(ctx context.Context, arg UpdateDocumentStatusParams) (Document, error) {
	sql := `UPDATE documents
		SET upload_status = $2, error_message = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, filename, file_path, file_type, total_pages, upload_status, error_message, is_demo, metadata, created_at, updated_at`
	row := q.db.QueryRow(ctx, sql, arg.ID, arg.UploadStatus, arg.ErrorMessage)
	var doc Document
	err := row.Scan(
		&doc.ID,
		&doc.Title,
		&doc.Filename,
		&doc.FilePath,
		&doc.FileType,
		&doc.TotalPages,
		&doc.UploadStatus,
		&doc.ErrorMessage,
		&doc.IsDemo,
		&doc.Metadata,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	return doc, err
}

// UpdateDocumentTotalPages updates a document's total page count.
func (q *Queries) UpdateDocumentTotalPages(ctx context.Context, arg UpdateDocumentTotalPagesParams) error {
	sql := `UPDATE documents SET total_pages = $2, updated_at = NOW() WHERE id = $1`
	_, err := q.db.Exec(ctx, sql, arg.ID, arg.TotalPages)
	return err
}

// CreateChunk creates a new chunk.
func (q *Queries) CreateChunk(ctx context.Context, arg CreateChunkParams) (Chunk, error) {
	sql := `INSERT INTO chunks (document_id, chunk_index, content, chapter_title, chapter_number, page_start, page_end, narrative_position, word_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, chunk_index, chapter_title, chapter_number`
	row := q.db.QueryRow(ctx, sql,
		arg.DocumentID,
		arg.ChunkIndex,
		arg.Content,
		arg.ChapterTitle,
		arg.ChapterNumber,
		arg.PageStart,
		arg.PageEnd,
		arg.NarrativePosition,
		arg.WordCount,
	)
	var chunk Chunk
	err := row.Scan(&chunk.ID, &chunk.ChunkIndex, &chunk.ChapterTitle, &chunk.ChapterNumber)
	return chunk, err
}

// CountChunksByDocument returns the number of chunks for a document.
func (q *Queries) CountChunksByDocument(ctx context.Context, documentID UUID) (int64, error) {
	sql := `SELECT COUNT(*) FROM chunks WHERE document_id = $1`
	var count int64
	err := q.db.QueryRow(ctx, sql, documentID).Scan(&count)
	return count, err
}

// ListChunksByDocument retrieves all chunks for a document.
func (q *Queries) ListChunksByDocument(ctx context.Context, documentID UUID) ([]Chunk, error) {
	sql := `SELECT id, document_id, chunk_index, content, chapter_title, chapter_number, page_start, page_end, narrative_position, word_count, created_at
		FROM chunks WHERE document_id = $1 ORDER BY chunk_index`
	rows, _ := q.db.Query(ctx, sql, documentID)
	defer rows.Close()

	var chunks []Chunk
	for rows.Next() {
		var chunk Chunk
		err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.ChunkIndex,
			&chunk.Content,
			&chunk.ChapterTitle,
			&chunk.ChapterNumber,
			&chunk.PageStart,
			&chunk.PageEnd,
			&chunk.NarrativePosition,
			&chunk.WordCount,
			&chunk.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}
	return chunks, rows.Err()
}

// GetChunk retrieves a single chunk by ID.
func (q *Queries) GetChunk(ctx context.Context, id UUID) (Chunk, error) {
	sql := `SELECT id, document_id, chunk_index, content, chapter_title, chapter_number, page_start, page_end, narrative_position, word_count, created_at
		FROM chunks WHERE id = $1`
	row := q.db.QueryRow(ctx, sql, id)
	var chunk Chunk
	err := row.Scan(
		&chunk.ID,
		&chunk.DocumentID,
		&chunk.ChunkIndex,
		&chunk.Content,
		&chunk.ChapterTitle,
		&chunk.ChapterNumber,
		&chunk.PageStart,
		&chunk.PageEnd,
		&chunk.NarrativePosition,
		&chunk.WordCount,
		&chunk.CreatedAt,
	)
	return chunk, err
}

// CreateDocumentParams defines parameters for creating a document.
type CreateDocumentParams struct {
	ID           UUID
	Title        string
	Filename     string
	FilePath     string
	FileType     string
	UploadStatus string
	IsDemo       bool
}

// UpdateDocumentStatusParams defines parameters for updating document status.
type UpdateDocumentStatusParams struct {
	ID           UUID
	UploadStatus string
	ErrorMessage *string
}

// UpdateDocumentTotalPagesParams defines parameters for updating total pages.
type UpdateDocumentTotalPagesParams struct {
	ID         UUID
	TotalPages int32
}

// CreateChunkParams defines parameters for creating a chunk.
type CreateChunkParams struct {
	DocumentID        UUID
	ChunkIndex        int32
	Content           string
	ChapterTitle      *string
	ChapterNumber     *int32
	PageStart         *int32
	PageEnd           *int32
	NarrativePosition int32
	WordCount         *int32
}
