-- name: CreateEvent :one
INSERT INTO events (
    document_id, title, description, event_type, date_text, date_start, date_end,
    date_precision, chronological_position, narrative_position, confidence,
    confidence_reason, metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM events WHERE id = $1;

-- name: ListEventsByDocument :many
SELECT * FROM events WHERE document_id = $1 ORDER BY narrative_position;

-- name: UpdateEventChronologicalPosition :one
UPDATE events
SET chronological_position = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateEventConfidence :one
UPDATE events
SET confidence = $2, confidence_reason = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateEventReviewStatus :one
UPDATE events
SET review_status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountEventsByDocument :one
SELECT COUNT(*) FROM events WHERE document_id = $1;
