-- name: CreateChunk :one
INSERT INTO chunks (
    source_id, chunk_index, content, chapter_title,
    chapter_number, page_start, page_end, narrative_position, word_count
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, chunk_index, chapter_title, chapter_number;

-- name: CountChunksBySource :one
SELECT COUNT(*) FROM chunks WHERE source_id = $1;

-- name: ListChunksBySource :many
SELECT * FROM chunks WHERE source_id = $1 ORDER BY chunk_index;

-- name: GetChunk :one
SELECT * FROM chunks WHERE id = $1;
