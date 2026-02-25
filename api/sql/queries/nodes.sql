-- name: CreateNode :one
INSERT INTO nodes (node_type, label, properties)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetNode :one
SELECT * FROM nodes
WHERE id = $1;

-- name: ListNodes :many
SELECT * FROM nodes
ORDER BY created_at DESC
LIMIT $1;

-- name: ListNodesByType :many
SELECT * FROM nodes
WHERE node_type = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: UpdateNode :one
UPDATE nodes
SET label = $2,
    properties = $3
WHERE id = $1
RETURNING *;

-- name: DeleteNode :exec
DELETE FROM nodes
WHERE id = $1;

-- name: ListNodesBySource :many
-- List all nodes that have provenance from a specific source
-- First find the document node for this source, then find all nodes with provenance from that document
SELECT DISTINCT n.*
FROM nodes n
INNER JOIN provenance p ON p.target_type = 'node' AND p.target_id = n.id
WHERE p.source_id IN (
    SELECT id FROM nodes
    WHERE node_type = 'document'
    AND properties->>'source_id' = $1::text
)
ORDER BY n.created_at DESC;

-- name: CountNodesByType :one
SELECT node_type, COUNT(*) as count
FROM nodes
GROUP BY node_type
ORDER BY count DESC;

-- name: GetDocumentNodeByLegacySourceID :one
SELECT * FROM nodes
WHERE node_type = 'document'
  AND properties->>'source_id' = $1::text
LIMIT 1;
