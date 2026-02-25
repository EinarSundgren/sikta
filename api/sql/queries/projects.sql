-- name: CreateProject :one
INSERT INTO projects (title, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE id = $1;

-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET title = $2,
    description = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;

-- name: GetProjectSources :many
SELECT * FROM sources WHERE project_id = $1 ORDER BY created_at DESC;

-- name: GetProjectStats :one
SELECT
    COUNT(DISTINCT s.id) as doc_count,
    COALESCE(SUM(CASE WHEN n.id IS NOT NULL THEN 1 ELSE 0 END), 0)::bigint as node_count,
    COALESCE(SUM(CASE WHEN e.id IS NOT NULL THEN 1 ELSE 0 END), 0)::bigint as edge_count
FROM sources s
LEFT JOIN nodes n ON n.source_id = s.id
LEFT JOIN edges e ON e.source_id = s.id
WHERE s.project_id = $1;

-- name: SetSourceProject :one
UPDATE sources
SET project_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
