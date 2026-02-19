-- name: CreateClaimEntity :one
INSERT INTO claim_entities (claim_id, entity_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (claim_id, entity_id) DO UPDATE SET role = EXCLUDED.role
RETURNING *;

-- name: ListClaimEntities :many
SELECT e.*, ce.role
FROM claim_entities ce
JOIN entities e ON ce.entity_id = e.id
WHERE ce.claim_id = $1;

-- name: ListEntityClaims :many
SELECT c.*, ce.role
FROM claim_entities ce
JOIN claims c ON ce.claim_id = c.id
WHERE ce.entity_id = $1
ORDER BY c.narrative_position;

-- name: ListClaimEntitiesBySource :many
SELECT ce.claim_id, e.id AS entity_id, e.name, e.entity_type, ce.role
FROM claim_entities ce
JOIN entities e ON ce.entity_id = e.id
JOIN claims c ON ce.claim_id = c.id
WHERE c.source_id = $1;
