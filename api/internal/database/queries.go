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

// ========== Events ==========

// CreateEvent creates a new event.
func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) (Event, error) {
	sql := `INSERT INTO events (
		document_id, title, description, event_type, date_text, date_start, date_end,
		date_precision, chronological_position, narrative_position, confidence,
		confidence_reason, metadata
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	RETURNING *`
	row := q.db.QueryRow(ctx, sql,
		arg.DocumentID, arg.Title, arg.Description, arg.EventType,
		arg.DateText, arg.DateStart, arg.DateEnd, arg.DatePrecision,
		arg.ChronologicalPosition, arg.NarrativePosition, arg.Confidence,
		arg.ConfidenceReason, arg.Metadata,
	)
	var event Event
	err := row.Scan(
		&event.ID, &event.DocumentID, &event.Title, &event.Description,
		&event.EventType, &event.DateText, &event.DateStart, &event.DateEnd,
		&event.DatePrecision, &event.ChronologicalPosition, &event.NarrativePosition,
		&event.Confidence, &event.ConfidenceReason, &event.ReviewStatus,
		&event.Metadata, &event.CreatedAt, &event.UpdatedAt,
	)
	return event, err
}

// GetEvent retrieves an event by ID.
func (q *Queries) GetEvent(ctx context.Context, id UUID) (Event, error) {
	sql := `SELECT * FROM events WHERE id = $1`
	row := q.db.QueryRow(ctx, sql, id)
	var event Event
	err := row.Scan(
		&event.ID, &event.DocumentID, &event.Title, &event.Description,
		&event.EventType, &event.DateText, &event.DateStart, &event.DateEnd,
		&event.DatePrecision, &event.ChronologicalPosition, &event.NarrativePosition,
		&event.Confidence, &event.ConfidenceReason, &event.ReviewStatus,
		&event.Metadata, &event.CreatedAt, &event.UpdatedAt,
	)
	return event, err
}

// ListEventsByDocument retrieves all events for a document.
func (q *Queries) ListEventsByDocument(ctx context.Context, documentID UUID) ([]Event, error) {
	sql := `SELECT * FROM events WHERE document_id = $1 ORDER BY narrative_position`
	rows, _ := q.db.Query(ctx, sql, documentID)
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID, &event.DocumentID, &event.Title, &event.Description,
			&event.EventType, &event.DateText, &event.DateStart, &event.DateEnd,
			&event.DatePrecision, &event.ChronologicalPosition, &event.NarrativePosition,
			&event.Confidence, &event.ConfidenceReason, &event.ReviewStatus,
			&event.Metadata, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

// UpdateEventChronologicalPosition updates an event's chronological position.
func (q *Queries) UpdateEventChronologicalPosition(ctx context.Context, arg UpdateEventChronologicalPositionParams) (Event, error) {
	sql := `UPDATE events
		SET chronological_position = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING *`
	row := q.db.QueryRow(ctx, sql, arg.ID, arg.ChronologicalPosition)
	var event Event
	err := row.Scan(
		&event.ID, &event.DocumentID, &event.Title, &event.Description,
		&event.EventType, &event.DateText, &event.DateStart, &event.DateEnd,
		&event.DatePrecision, &event.ChronologicalPosition, &event.NarrativePosition,
		&event.Confidence, &event.ConfidenceReason, &event.ReviewStatus,
		&event.Metadata, &event.CreatedAt, &event.UpdatedAt,
	)
	return event, err
}

// CountEventsByDocument returns the number of events for a document.
func (q *Queries) CountEventsByDocument(ctx context.Context, documentID UUID) (int64, error) {
	sql := `SELECT COUNT(*) FROM events WHERE document_id = $1`
	var count int64
	err := q.db.QueryRow(ctx, sql, documentID).Scan(&count)
	return count, err
}

// ========== Entities ==========

// CreateEntity creates a new entity.
func (q *Queries) CreateEntity(ctx context.Context, arg CreateEntityParams) (Entity, error) {
	sql := `INSERT INTO entities (
		document_id, name, entity_type, aliases, description,
		first_appearance_chunk, last_appearance_chunk, confidence, metadata
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING *`
	row := q.db.QueryRow(ctx, sql,
		arg.DocumentID, arg.Name, arg.EntityType, arg.Aliases, arg.Description,
		arg.FirstAppearanceChunk, arg.LastAppearanceChunk, arg.Confidence, arg.Metadata,
	)
	var entity Entity
	err := row.Scan(
		&entity.ID, &entity.DocumentID, &entity.Name, &entity.EntityType,
		&entity.Aliases, &entity.Description, &entity.FirstAppearanceChunk,
		&entity.LastAppearanceChunk, &entity.Confidence, &entity.ReviewStatus,
		&entity.Metadata, &entity.CreatedAt, &entity.UpdatedAt,
	)
	return entity, err
}

// GetEntity retrieves an entity by ID.
func (q *Queries) GetEntity(ctx context.Context, id UUID) (Entity, error) {
	sql := `SELECT * FROM entities WHERE id = $1`
	row := q.db.QueryRow(ctx, sql, id)
	var entity Entity
	err := row.Scan(
		&entity.ID, &entity.DocumentID, &entity.Name, &entity.EntityType,
		&entity.Aliases, &entity.Description, &entity.FirstAppearanceChunk,
		&entity.LastAppearanceChunk, &entity.Confidence, &entity.ReviewStatus,
		&entity.Metadata, &entity.CreatedAt, &entity.UpdatedAt,
	)
	return entity, err
}

// ListEntitiesByDocument retrieves all entities for a document.
func (q *Queries) ListEntitiesByDocument(ctx context.Context, documentID UUID) ([]Entity, error) {
	sql := `SELECT * FROM entities WHERE document_id = $1 ORDER BY first_appearance_chunk`
	rows, _ := q.db.Query(ctx, sql, documentID)
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var entity Entity
		err := rows.Scan(
			&entity.ID, &entity.DocumentID, &entity.Name, &entity.EntityType,
			&entity.Aliases, &entity.Description, &entity.FirstAppearanceChunk,
			&entity.LastAppearanceChunk, &entity.Confidence, &entity.ReviewStatus,
			&entity.Metadata, &entity.CreatedAt, &entity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, rows.Err()
}

// UpdateEntityAliases updates an entity's aliases.
func (q *Queries) UpdateEntityAliases(ctx context.Context, arg UpdateEntityAliasesParams) (Entity, error) {
	sql := `UPDATE entities
		SET aliases = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING *`
	row := q.db.QueryRow(ctx, sql, arg.ID, arg.Aliases)
	var entity Entity
	err := row.Scan(
		&entity.ID, &entity.DocumentID, &entity.Name, &entity.EntityType,
		&entity.Aliases, &entity.Description, &entity.FirstAppearanceChunk,
		&entity.LastAppearanceChunk, &entity.Confidence, &entity.ReviewStatus,
		&entity.Metadata, &entity.CreatedAt, &entity.UpdatedAt,
	)
	return entity, err
}

// CountEntitiesByDocument returns the number of entities for a document.
func (q *Queries) CountEntitiesByDocument(ctx context.Context, documentID UUID) (int64, error) {
	sql := `SELECT COUNT(*) FROM entities WHERE document_id = $1`
	var count int64
	err := q.db.QueryRow(ctx, sql, documentID).Scan(&count)
	return count, err
}

// ========== Relationships ==========

// CreateRelationship creates a new relationship.
func (q *Queries) CreateRelationship(ctx context.Context, arg CreateRelationshipParams) (Relationship, error) {
	sql := `INSERT INTO relationships (
		document_id, entity_a_id, entity_b_id, relationship_type, description,
		start_event_id, end_event_id, confidence, metadata
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING *`
	row := q.db.QueryRow(ctx, sql,
		arg.DocumentID, arg.EntityAID, arg.EntityBID, arg.RelationshipType,
		arg.Description, arg.StartEventID, arg.EndEventID, arg.Confidence, arg.Metadata,
	)
	var rel Relationship
	err := row.Scan(
		&rel.ID, &rel.DocumentID, &rel.EntityAID, &rel.EntityBID,
		&rel.RelationshipType, &rel.Description, &rel.StartEventID,
		&rel.EndEventID, &rel.Confidence, &rel.ReviewStatus,
		&rel.Metadata, &rel.CreatedAt, &rel.UpdatedAt,
	)
	return rel, err
}

// GetRelationship retrieves a relationship by ID.
func (q *Queries) GetRelationship(ctx context.Context, id UUID) (Relationship, error) {
	sql := `SELECT * FROM relationships WHERE id = $1`
	row := q.db.QueryRow(ctx, sql, id)
	var rel Relationship
	err := row.Scan(
		&rel.ID, &rel.DocumentID, &rel.EntityAID, &rel.EntityBID,
		&rel.RelationshipType, &rel.Description, &rel.StartEventID,
		&rel.EndEventID, &rel.Confidence, &rel.ReviewStatus,
		&rel.Metadata, &rel.CreatedAt, &rel.UpdatedAt,
	)
	return rel, err
}

// ListRelationshipsByDocument retrieves all relationships for a document.
func (q *Queries) ListRelationshipsByDocument(ctx context.Context, documentID UUID) ([]Relationship, error) {
	sql := `SELECT * FROM relationships WHERE document_id = $1`
	rows, _ := q.db.Query(ctx, sql, documentID)
	defer rows.Close()

	var relationships []Relationship
	for rows.Next() {
		var rel Relationship
		err := rows.Scan(
			&rel.ID, &rel.DocumentID, &rel.EntityAID, &rel.EntityBID,
			&rel.RelationshipType, &rel.Description, &rel.StartEventID,
			&rel.EndEventID, &rel.Confidence, &rel.ReviewStatus,
			&rel.Metadata, &rel.CreatedAt, &rel.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		relationships = append(relationships, rel)
	}
	return relationships, rows.Err()
}

// CountRelationshipsByDocument returns the number of relationships for a document.
func (q *Queries) CountRelationshipsByDocument(ctx context.Context, documentID UUID) (int64, error) {
	sql := `SELECT COUNT(*) FROM relationships WHERE document_id = $1`
	var count int64
	err := q.db.QueryRow(ctx, sql, documentID).Scan(&count)
	return count, err
}

// ========== Source References ==========

// CreateSourceReference creates a new source reference.
func (q *Queries) CreateSourceReference(ctx context.Context, arg CreateSourceReferenceParams) (SourceReference, error) {
	sql := `INSERT INTO source_references (
		chunk_id, event_id, entity_id, relationship_id, excerpt, char_start, char_end
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING *`
	row := q.db.QueryRow(ctx, sql,
		arg.ChunkID, arg.EventID, arg.EntityID, arg.RelationshipID,
		arg.Excerpt, arg.CharStart, arg.CharEnd,
	)
	var ref SourceReference
	err := row.Scan(
		&ref.ID, &ref.ChunkID, &ref.EventID, &ref.EntityID,
		&ref.RelationshipID, &ref.Excerpt, &ref.CharStart, &ref.CharEnd,
		&ref.CreatedAt,
	)
	return ref, err
}

// ========== Parameter Types ==========

// CreateEventParams defines parameters for creating an event.
type CreateEventParams struct {
	DocumentID              UUID
	Title                   string
	Description             *string
	EventType               string
	DateText                *string
	DateStart               interface{}
	DateEnd                 interface{}
	DatePrecision           *string
	ChronologicalPosition   *int32
	NarrativePosition        int32
	Confidence              float64
	ConfidenceReason        *string
	Metadata                interface{}
}

// UpdateEventChronologicalPositionParams defines parameters for updating event position.
type UpdateEventChronologicalPositionParams struct {
	ID                    UUID
	ChronologicalPosition int32
}

// CreateEntityParams defines parameters for creating an entity.
type CreateEntityParams struct {
	DocumentID           UUID
	Name                 string
	EntityType           string
	Aliases              []string
	Description          *string
	FirstAppearanceChunk *int32
	LastAppearanceChunk  *int32
	Confidence           float64
	Metadata             interface{}
}

// UpdateEntityAliasesParams defines parameters for updating entity aliases.
type UpdateEntityAliasesParams struct {
	ID      UUID
	Aliases []string
}

// CreateRelationshipParams defines parameters for creating a relationship.
type CreateRelationshipParams struct {
	DocumentID     UUID
	EntityAID      UUID
	EntityBID      UUID
	RelationshipType string
	Description    *string
	StartEventID   *UUID
	EndEventID     *UUID
	Confidence     float64
	Metadata       interface{}
}

// CreateSourceReferenceParams defines parameters for creating a source reference.
type CreateSourceReferenceParams struct {
	ChunkID        UUID
	EventID        *UUID
	EntityID       *UUID
	RelationshipID *UUID
	Excerpt        string
	CharStart      *int32
	CharEnd        *int32
}

// Inconsistency-related types and queries

// Inconsistency represents a detected inconsistency in the document.
type Inconsistency struct {
	ID               UUID
	DocumentID       UUID
	InconsistencyType string
	Severity         string
	Title            string
	Description      string
	ResolutionStatus string
	ResolutionNote   *string
	Metadata         interface{}
	CreatedAt        interface{}
	UpdatedAt        interface{}
}

// InconsistencyItem links an inconsistency to events or entities.
type InconsistencyItem struct {
	ID              UUID
	InconsistencyID UUID
	EventID         *UUID
	EntityID        *UUID
	Side            *string
	Description     string
	CreatedAt       interface{}
}

// CreateInconsistencyParams defines parameters for creating an inconsistency.
type CreateInconsistencyParams struct {
	ID               UUID
	DocumentID       UUID
	InconsistencyType string
	Severity         string
	Title            string
	Description      string
	ResolutionStatus string
	ResolutionNote   interface{}
	Metadata         interface{}
}

// CreateInconsistencyItemParams defines parameters for creating an inconsistency item.
type CreateInconsistencyItemParams struct {
	ID              UUID
	InconsistencyID UUID
	EventID         *UUID
	EntityID        *UUID
	Side            *string
	Description     string
}

// UpdateInconsistencyResolutionParams defines parameters for updating inconsistency resolution.
type UpdateInconsistencyResolutionParams struct {
	ID               UUID
	ResolutionStatus string
	ResolutionNote   interface{}
}

// CreateInconsistency creates a new inconsistency.
func (q *Queries) CreateInconsistency(ctx context.Context, arg CreateInconsistencyParams) (Inconsistency, error) {
	sql := `INSERT INTO inconsistencies (id, document_id, inconsistency_type, severity, title, description, resolution_status, resolution_note, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, document_id, inconsistency_type, severity, title, description, resolution_status, resolution_note, metadata, created_at, updated_at`
	row := q.db.QueryRow(ctx, sql,
		arg.ID,
		arg.DocumentID,
		arg.InconsistencyType,
		arg.Severity,
		arg.Title,
		arg.Description,
		arg.ResolutionStatus,
		arg.ResolutionNote,
		arg.Metadata,
	)
	var inc Inconsistency
	err := row.Scan(
		&inc.ID,
		&inc.DocumentID,
		&inc.InconsistencyType,
		&inc.Severity,
		&inc.Title,
		&inc.Description,
		&inc.ResolutionStatus,
		&inc.ResolutionNote,
		&inc.Metadata,
		&inc.CreatedAt,
		&inc.UpdatedAt,
	)
	return inc, err
}

// ListInconsistenciesByDocument retrieves all inconsistencies for a document.
func (q *Queries) ListInconsistenciesByDocument(ctx context.Context, documentID UUID) ([]Inconsistency, error) {
	sql := `SELECT id, document_id, inconsistency_type, severity, title, description, resolution_status, resolution_note, metadata, created_at, updated_at
		FROM inconsistencies
		WHERE document_id = $1
		ORDER BY
			CASE severity
				WHEN 'conflict' THEN 1
				WHEN 'warning' THEN 2
				WHEN 'info' THEN 3
			END,
			created_at DESC`
	rows, err := q.db.Query(ctx, sql, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inconsistencies []Inconsistency
	for rows.Next() {
		var inc Inconsistency
		if err := rows.Scan(
			&inc.ID,
			&inc.DocumentID,
			&inc.InconsistencyType,
			&inc.Severity,
			&inc.Title,
			&inc.Description,
			&inc.ResolutionStatus,
			&inc.ResolutionNote,
			&inc.Metadata,
			&inc.CreatedAt,
			&inc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		inconsistencies = append(inconsistencies, inc)
	}
	return inconsistencies, rows.Err()
}

// GetInconsistency retrieves a single inconsistency by ID.
func (q *Queries) GetInconsistency(ctx context.Context, id UUID) (Inconsistency, error) {
	sql := `SELECT id, document_id, inconsistency_type, severity, title, description, resolution_status, resolution_note, metadata, created_at, updated_at
		FROM inconsistencies
		WHERE id = $1`
	row := q.db.QueryRow(ctx, sql, id)
	var inc Inconsistency
	err := row.Scan(
		&inc.ID,
		&inc.DocumentID,
		&inc.InconsistencyType,
		&inc.Severity,
		&inc.Title,
		&inc.Description,
		&inc.ResolutionStatus,
		&inc.ResolutionNote,
		&inc.Metadata,
		&inc.CreatedAt,
		&inc.UpdatedAt,
	)
	return inc, err
}

// ListInconsistencyItems retrieves all items for an inconsistency with event/entity details.
func (q *Queries) ListInconsistencyItems(ctx context.Context, inconsistencyID UUID) ([]InconsistencyItem, error) {
	sql := `
		SELECT 
			ii.id,
			ii.inconsistency_id,
			ii.event_id,
			ii.entity_id,
			ii.side,
			ii.description,
			e.title as event_title,
			e.event_type,
			en.name as entity_name,
			en.entity_type
		FROM inconsistency_items ii
		LEFT JOIN events e ON ii.event_id = e.id
		LEFT JOIN entities en ON ii.entity_id = en.id
		WHERE ii.inconsistency_id = $1
		ORDER BY ii.side, ii.id
	`
	rows, err := q.db.Query(ctx, sql, inconsistencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InconsistencyItem
	for rows.Next() {
		var item InconsistencyItem
		var eventTitle, eventType *string
		var entityName, entityType *string
		if err := rows.Scan(
			&item.ID,
			&item.InconsistencyID,
			&item.EventID,
			&item.EntityID,
			&item.Side,
			&item.Description,
			&eventTitle,
			&eventType,
			&entityName,
			&entityType,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// CreateInconsistencyItem links an inconsistency to an event or entity.
func (q *Queries) CreateInconsistencyItem(ctx context.Context, arg CreateInconsistencyItemParams) (InconsistencyItem, error) {
	sql := `INSERT INTO inconsistency_items (id, inconsistency_id, event_id, entity_id, side, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, inconsistency_id, event_id, entity_id, side, description, created_at`
	row := q.db.QueryRow(ctx, sql,
		arg.ID,
		arg.InconsistencyID,
		arg.EventID,
		arg.EntityID,
		arg.Side,
		arg.Description,
	)
	var item InconsistencyItem
	err := row.Scan(
		&item.ID,
		&item.InconsistencyID,
		&item.EventID,
		&item.EntityID,
		&item.Side,
		&item.Description,
		&item.CreatedAt,
	)
	return item, err
}

// UpdateInconsistencyResolution updates the resolution status and note.
func (q *Queries) UpdateInconsistencyResolution(ctx context.Context, arg UpdateInconsistencyResolutionParams) (Inconsistency, error) {
	sql := `UPDATE inconsistencies
		SET resolution_status = $2,
		    resolution_note = $3,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, document_id, inconsistency_type, severity, title, description, resolution_status, resolution_note, metadata, created_at, updated_at`
	row := q.db.QueryRow(ctx, sql, arg.ID, arg.ResolutionStatus, arg.ResolutionNote)
	var inc Inconsistency
	err := row.Scan(
		&inc.ID,
		&inc.DocumentID,
		&inc.InconsistencyType,
		&inc.Severity,
		&inc.Title,
		&inc.Description,
		&inc.ResolutionStatus,
		&inc.ResolutionNote,
		&inc.Metadata,
		&inc.CreatedAt,
		&inc.UpdatedAt,
	)
	return inc, err
}

// CountInconsistenciesByDocument counts inconsistencies by severity for a document.
func (q *Queries) CountInconsistenciesByDocument(ctx context.Context, documentID UUID) (CountInconsistenciesByDocumentRow, error) {
	sql := `SELECT
			COUNT(*) as total,
			SUM(CASE WHEN severity = 'conflict' THEN 1 ELSE 0 END) as conflicts,
			SUM(CASE WHEN severity = 'warning' THEN 1 ELSE 0 END) as warnings,
			SUM(CASE WHEN severity = 'info' THEN 1 ELSE 0 END) as info
		FROM inconsistencies
		WHERE document_id = $1`
	queryRow := q.db.QueryRow(ctx, sql, documentID)
	var result CountInconsistenciesByDocumentRow
	err := queryRow.Scan(&result.Total, &result.Conflicts, &result.Warnings, &result.Info)
	return result, err
}

// CountInconsistenciesByDocumentRow represents the result of counting inconsistencies.
type CountInconsistenciesByDocumentRow struct {
	Total    int64
	Conflicts int64
	Warnings  int64
	Info      int64
}
