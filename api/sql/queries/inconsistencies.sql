-- name: CreateInconsistency
INSERT INTO inconsistencies (id, document_id, inconsistency_type, severity, title, description, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListInconsistenciesByDocument
SELECT * FROM inconsistencies
WHERE document_id = $1
ORDER BY
    CASE severity
        WHEN 'conflict' THEN 1
        WHEN 'warning' THEN 2
        WHEN 'info' THEN 3
    END,
    created_at DESC;

-- name: GetInconsistency
SELECT * FROM inconsistencies
WHERE id = $1;

-- name: ListInconsistencyItems
SELECT
    ii.*,
    e.title as event_title,
    e.event_type,
    en.name as entity_name,
    en.entity_type
FROM inconsistency_items ii
LEFT JOIN events e ON ii.event_id = e.id
LEFT JOIN entities en ON ii.entity_id = en.id
WHERE ii.inconsistency_id = $1
ORDER BY ii.side, ii.id;

-- name: CreateInconsistencyItem
INSERT INTO inconsistency_items (id, inconsistency_id, event_id, entity_id, side, description)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateInconsistencyResolution
UPDATE inconsistencies
SET resolution_status = $2,
    resolution_note = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountInconsistenciesByDocument
SELECT
    COUNT(*) as total,
    SUM(CASE WHEN severity = 'conflict' THEN 1 ELSE 0 END) as conflicts,
    SUM(CASE WHEN severity = 'warning' THEN 1 ELSE 0 END) as warnings,
    SUM(CASE WHEN severity = 'info' THEN 1 ELSE 0 END) as info
FROM inconsistencies
WHERE document_id = $1;
