-- name: CreateProvenance :one
INSERT INTO provenance (
    target_type, target_id, source_id, excerpt, location,
    confidence, trust, status, modality,
    claimed_time_start, claimed_time_end, claimed_time_text,
    claimed_geo_region, claimed_geo_text, claimed_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: GetProvenance :one
SELECT * FROM provenance
WHERE id = $1;

-- name: ListProvenanceByTarget :many
SELECT * FROM provenance
WHERE target_type = $1 AND target_id = $2
ORDER BY created_at DESC;

-- name: ListProvenanceBySource :many
SELECT * FROM provenance
WHERE source_id = $1
ORDER BY created_at DESC;

-- name: ListProvenanceByStatus :many
SELECT * FROM provenance
WHERE status = $1
ORDER BY created_at DESC;

-- name: UpdateProvenanceStatus :one
UPDATE provenance
SET status = $2
WHERE id = $1
RETURNING *;

-- name: UpdateProvenance :one
UPDATE provenance
SET excerpt = $2,
    location = $3,
    confidence = $4,
    trust = $5,
    status = $6,
    modality = $7,
    claimed_time_start = $8,
    claimed_time_end = $9,
    claimed_time_text = $10,
    claimed_geo_region = $11,
    claimed_geo_text = $12
WHERE id = $1
RETURNING *;

-- name: DeleteProvenance :exec
DELETE FROM provenance
WHERE id = $1;

-- name: CountProvenanceByStatus :one
SELECT status, COUNT(*) as count
FROM provenance
GROUP BY status
ORDER BY count DESC;

-- name: CountProvenanceByModality :one
SELECT modality, COUNT(*) as count
FROM provenance
GROUP BY modality
ORDER BY count DESC;

-- name: CountClaimProvenanceByStatusForSource :many
-- Count claim/event nodes by their asserted provenance status for a document node
SELECT p.status, COUNT(*) as count
FROM provenance p
INNER JOIN nodes n ON n.id = p.target_id AND p.target_type = 'node'
WHERE p.source_id = $1
  AND n.node_type IN ('event', 'attribute', 'relation')
  AND p.modality = 'asserted'
GROUP BY p.status;

-- name: CountEntityProvenanceByStatusForSource :many
-- Count entity nodes by their asserted provenance status for a document node
SELECT p.status, COUNT(*) as count
FROM provenance p
INNER JOIN nodes n ON n.id = p.target_id AND p.target_type = 'node'
WHERE p.source_id = $1
  AND n.node_type IN ('person', 'place', 'organization', 'object')
  AND p.modality = 'asserted'
GROUP BY p.status;

-- name: UpdateProvenanceStatusByTarget :exec
-- Update status on asserted provenance records for a given target node or edge
UPDATE provenance
SET status = $3
WHERE target_type = $1
  AND target_id = $2
  AND modality = 'asserted';
