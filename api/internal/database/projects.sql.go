package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

// ProjectQueries contains hand-written project queries until sqlc can handle migrations
// These will be replaced by sqlc-generated code once the migration issue is resolved

// CreateProject creates a new project
func (q *Queries) CreateProject(ctx context.Context, title string, description pgtype.Text) (*Project, error) {
	query := `INSERT INTO projects (title, description) VALUES ($1, $2) RETURNING id, title, description, created_at, updated_at`
	var p Project
	err := q.db.QueryRow(ctx, query, title, description).Scan(
		&p.ID, &p.Title, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetProject retrieves a project by ID
func (q *Queries) GetProject(ctx context.Context, id pgtype.UUID) (*Project, error) {
	query := `SELECT id, title, description, created_at, updated_at FROM projects WHERE id = $1`
	var p Project
	err := q.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Title, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ListProjects lists all projects
func (q *Queries) ListProjects(ctx context.Context) ([]*Project, error) {
	query := `SELECT id, title, description, created_at, updated_at FROM projects ORDER BY created_at DESC`
	rows, err := q.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}
	return projects, rows.Err()
}

// UpdateProject updates a project
func (q *Queries) UpdateProject(ctx context.Context, id pgtype.UUID, title string, description pgtype.Text) (*Project, error) {
	query := `UPDATE projects SET title = $2, description = $3, updated_at = NOW() WHERE id = $1 RETURNING id, title, description, created_at, updated_at`
	var p Project
	err := q.db.QueryRow(ctx, query, id, title, description).Scan(
		&p.ID, &p.Title, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteProject deletes a project
func (q *Queries) DeleteProject(ctx context.Context, id pgtype.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := q.db.Exec(ctx, query, id)
	return err
}

// GetProjectSources retrieves all sources for a project
func (q *Queries) GetProjectSources(ctx context.Context, projectID pgtype.UUID) ([]*Source, error) {
	query := `SELECT id, title, filename, file_path, file_type, total_pages, upload_status, error_message, is_demo, metadata, created_at, updated_at, source_trust, trust_reason, project_id
	          FROM sources WHERE project_id = $1 ORDER BY created_at DESC`
	rows, err := q.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []*Source
	for rows.Next() {
		var s Source
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Filename, &s.FilePath, &s.FileType,
			&s.TotalPages, &s.UploadStatus, &s.ErrorMessage, &s.IsDemo,
			&s.Metadata, &s.CreatedAt, &s.UpdatedAt, &s.SourceTrust,
			&s.TrustReason, &s.ProjectID,
		); err != nil {
			return nil, err
		}
		sources = append(sources, &s)
	}
	return sources, rows.Err()
}

// ProjectStats contains statistics about a project
type ProjectStats struct {
	DocCount  int64 `json:"doc_count"`
	NodeCount int64 `json:"node_count"`
	EdgeCount int64 `json:"edge_count"`
}

// GetProjectStats retrieves statistics for a project
func (q *Queries) GetProjectStats(ctx context.Context, projectID pgtype.UUID) (*ProjectStats, error) {
	query := `SELECT
		COUNT(DISTINCT s.id) as doc_count,
		COALESCE(SUM(CASE WHEN n.id IS NOT NULL THEN 1 ELSE 0 END), 0)::bigint as node_count,
		COALESCE(SUM(CASE WHEN e.id IS NOT NULL THEN 1 ELSE 0 END), 0)::bigint as edge_count
		FROM sources s
		LEFT JOIN nodes n ON n.source_id = s.id
		LEFT JOIN edges e ON e.source_id = s.id
		WHERE s.project_id = $1`
	var stats ProjectStats
	err := q.db.QueryRow(ctx, query, projectID).Scan(&stats.DocCount, &stats.NodeCount, &stats.EdgeCount)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// SetSourceProject sets the project for a source
func (q *Queries) SetSourceProject(ctx context.Context, sourceID, projectID pgtype.UUID) (*Source, error) {
	query := `UPDATE sources SET project_id = $2, updated_at = NOW() WHERE id = $1
	          RETURNING id, title, filename, file_path, file_type, total_pages, upload_status, error_message, is_demo, metadata, created_at, updated_at, source_trust, trust_reason, project_id`
	var s Source
	err := q.db.QueryRow(ctx, query, sourceID, projectID).Scan(
		&s.ID, &s.Title, &s.Filename, &s.FilePath, &s.FileType,
		&s.TotalPages, &s.UploadStatus, &s.ErrorMessage, &s.IsDemo,
		&s.Metadata, &s.CreatedAt, &s.UpdatedAt, &s.SourceTrust,
		&s.TrustReason, &s.ProjectID,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
