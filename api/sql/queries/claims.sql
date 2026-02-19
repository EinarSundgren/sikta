-- name: CreateClaim :one
INSERT INTO claims (
    source_id, claim_type, title, description, event_type, date_text, date_start, date_end,
    date_precision, chronological_position, narrative_position, confidence,
    confidence_reason, metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: GetClaim :one
SELECT * FROM claims WHERE id = $1;

-- name: ListClaimsBySource :many
SELECT * FROM claims WHERE source_id = $1 ORDER BY narrative_position;

-- name: UpdateClaimChronologicalPosition :one
UPDATE claims
SET chronological_position = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateClaimConfidence :one
UPDATE claims
SET confidence = $2, confidence_reason = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateClaimReviewStatus :one
UPDATE claims
SET review_status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountClaimsBySource :one
SELECT COUNT(*) FROM claims WHERE source_id = $1;
