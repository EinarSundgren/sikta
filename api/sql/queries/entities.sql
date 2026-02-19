-- name: CreateEntity :one
INSERT INTO entities (
    source_id, name, entity_type, aliases, description,
    first_appearance_chunk, last_appearance_chunk, confidence, metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetEntity :one
SELECT * FROM entities WHERE id = $1;

-- name: ListEntitiesBySource :many
SELECT * FROM entities WHERE source_id = $1 ORDER BY first_appearance_chunk;

-- name: UpdateEntityAliases :one
UPDATE entities
SET aliases = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateEntityReviewStatus :one
UPDATE entities
SET review_status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountEntitiesBySource :one
SELECT COUNT(*) FROM entities WHERE source_id = $1;

-- name: CountEntitiesByStatusForSource :many
SELECT review_status, COUNT(*) AS count
FROM entities
WHERE source_id = $1
GROUP BY review_status;
