-- name: CreateDocument :one
INSERT INTO documents (title, filename, file_path, file_type, upload_status, is_demo)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetDocument :one
SELECT * FROM documents WHERE id = $1;

-- name: ListDocuments :many
SELECT * FROM documents ORDER BY created_at DESC;

-- name: UpdateDocumentStatus :one
UPDATE documents
SET upload_status = $2,
    error_message = $3,
    updated_at    = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDocumentTotalPages :exec
UPDATE documents
SET total_pages = $2,
    updated_at  = NOW()
WHERE id = $1;
