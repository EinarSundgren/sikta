-- name: CreateSource :one
INSERT INTO sources (title, filename, file_path, file_type, upload_status, is_demo)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSource :one
SELECT * FROM sources WHERE id = $1;

-- name: ListSources :many
SELECT * FROM sources ORDER BY created_at DESC;

-- name: UpdateSourceStatus :one
UPDATE sources
SET upload_status = $2,
    error_message = $3,
    updated_at    = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateSourceTotalPages :exec
UPDATE sources
SET total_pages = $2,
    updated_at  = NOW()
WHERE id = $1;

-- name: DeleteSource :exec
DELETE FROM sources WHERE id = $1;
