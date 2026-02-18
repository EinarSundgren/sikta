-- name: CreateSourceReference :one
INSERT INTO source_references (
    chunk_id, event_id, entity_id, relationship_id, excerpt, char_start, char_end
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListSourceReferencesByEvent :many
SELECT sr.*, c.chapter_title, c.chapter_number
FROM source_references sr
JOIN chunks c ON sr.chunk_id = c.id
WHERE sr.event_id = $1;

-- name: ListSourceReferencesByEntity :many
SELECT sr.*, c.chapter_title, c.chapter_number
FROM source_references sr
JOIN chunks c ON sr.chunk_id = c.id
WHERE sr.entity_id = $1;

-- name: ListSourceReferencesByRelationship :many
SELECT sr.*, c.chapter_title, c.chapter_number
FROM source_references sr
JOIN chunks c ON sr.chunk_id = c.id
WHERE sr.relationship_id = $1;
