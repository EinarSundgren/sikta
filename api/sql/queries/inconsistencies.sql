-- name: CreateInconsistency :one
INSERT INTO inconsistencies (id, source_id, inconsistency_type, severity, title, description, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListInconsistenciesBySource :many
SELECT * FROM inconsistencies
WHERE source_id = $1
ORDER BY
    CASE severity
        WHEN 'conflict' THEN 1
        WHEN 'warning' THEN 2
        WHEN 'info' THEN 3
    END,
    created_at DESC;

-- name: GetInconsistency :one
SELECT * FROM inconsistencies
WHERE id = $1;

-- name: ListInconsistencyItems :many
SELECT
    ii.*,
    c.title as claim_title,
    c.event_type,
    en.name as entity_name,
    en.entity_type
FROM inconsistency_items ii
LEFT JOIN claims c ON ii.claim_id = c.id
LEFT JOIN entities en ON ii.entity_id = en.id
WHERE ii.inconsistency_id = $1
ORDER BY ii.side, ii.id;

-- name: CreateInconsistencyItem :one
INSERT INTO inconsistency_items (id, inconsistency_id, claim_id, entity_id, side, description)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateInconsistencyResolution :one
UPDATE inconsistencies
SET resolution_status = $2,
    resolution_note = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountInconsistenciesBySource :one
SELECT
    COUNT(*) as total,
    SUM(CASE WHEN severity = 'conflict' THEN 1 ELSE 0 END) as conflicts,
    SUM(CASE WHEN severity = 'warning' THEN 1 ELSE 0 END) as warnings,
    SUM(CASE WHEN severity = 'info' THEN 1 ELSE 0 END) as info
FROM inconsistencies
WHERE source_id = $1;

-- name: ListInconsistenciesByClaim :many
SELECT i.*
FROM inconsistencies i
INNER JOIN inconsistency_items ii ON i.id = ii.inconsistency_id
WHERE ii.claim_id = $1
ORDER BY
    CASE i.severity
        WHEN 'conflict' THEN 1
        WHEN 'warning' THEN 2
        WHEN 'info' THEN 3
    END,
    i.created_at DESC;
