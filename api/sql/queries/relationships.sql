-- name: CreateRelationship :one
INSERT INTO relationships (
    document_id, entity_a_id, entity_b_id, relationship_type, description,
    start_event_id, end_event_id, confidence, metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetRelationship :one
SELECT * FROM relationships WHERE id = $1;

-- name: ListRelationshipsByDocument :many
SELECT * FROM relationships WHERE document_id = $1;

-- name: ListRelationshipsByEntity :many
SELECT * FROM relationships
WHERE entity_a_id = $1 OR entity_b_id = $1;

-- name: UpdateRelationshipReviewStatus :one
UPDATE relationships
SET review_status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountRelationshipsByDocument :one
SELECT COUNT(*) FROM relationships WHERE document_id = $1;
