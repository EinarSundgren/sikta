-- name: CreateEdge :one
INSERT INTO edges (edge_type, source_node, target_node, properties, is_negated)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetEdge :one
SELECT * FROM edges
WHERE id = $1;

-- name: ListEdges :many
SELECT * FROM edges
ORDER BY created_at DESC
LIMIT $1;

-- name: ListEdgesByType :many
SELECT * FROM edges
WHERE edge_type = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListEdgesBySource :many
SELECT * FROM edges
WHERE source_node = $1
ORDER BY created_at DESC;

-- name: ListEdgesByTarget :many
SELECT * FROM edges
WHERE target_node = $1
ORDER BY created_at DESC;

-- name: ListEdgesByNodes :many
-- Get all edges between two nodes (both directions)
SELECT * FROM edges
WHERE (source_node = $1 AND target_node = $2)
   OR (source_node = $2 AND target_node = $1)
ORDER BY created_at DESC;

-- name: UpdateEdge :one
UPDATE edges
SET edge_type = $2,
    properties = $3,
    is_negated = $4
WHERE id = $1
RETURNING *;

-- name: DeleteEdge :exec
DELETE FROM edges
WHERE id = $1;

-- name: CountEdgesByType :one
SELECT edge_type, COUNT(*) as count
FROM edges
GROUP BY edge_type
ORDER BY count DESC;

-- name: ListEdgesBySourceDocument :many
-- List all edges that have provenance from a specific source
-- First find the document node for this source, then find all edges with provenance from that document
SELECT DISTINCT e.*
FROM edges e
INNER JOIN provenance p ON p.target_type = 'edge' AND p.target_id = e.id
WHERE p.source_id IN (
    SELECT id FROM nodes
    WHERE node_type = 'document'
    AND properties->>'source_id' = $1::text
)
ORDER BY e.created_at DESC;
