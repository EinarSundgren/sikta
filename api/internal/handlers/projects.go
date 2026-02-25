package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/einarsundgren/sikta/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	db     *database.Queries
	logger *slog.Logger
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(db *database.Queries, logger *slog.Logger) *ProjectHandler {
	return &ProjectHandler{
		db:     db,
		logger: logger,
	}
}

// CreateProjectRequest is the request body for creating a project
type CreateProjectRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ProjectResponse is the response for a single project
type ProjectResponse struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Stats       *ProjectStatsDTO  `json:"stats,omitempty"`
}

// ProjectStatsDTO contains statistics about a project
type ProjectStatsDTO struct {
	DocCount  int64 `json:"doc_count"`
	NodeCount int64 `json:"node_count"`
	EdgeCount int64 `json:"edge_count"`
}

// ListProjects handles GET /api/projects
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.db.ListProjects(r.Context())
	if err != nil {
		h.logger.Error("failed to list projects", "error", err)
		http.Error(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}

	response := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		response[i] = ProjectResponse{
			ID:          pgUUIDToStr(p.ID),
			Title:       p.Title,
			Description: p.Description.String,
			CreatedAt:   p.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   p.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProject handles GET /api/projects/{id}
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.db.GetProject(r.Context(), strToPgUUID(id.String()))
	if err != nil {
		h.logger.Error("failed to get project", "error", err, "id", idStr)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Get stats
	stats, err := h.db.GetProjectStats(r.Context(), strToPgUUID(id.String()))
	if err != nil {
		h.logger.Warn("failed to get project stats", "error", err)
		// Don't fail, just don't include stats
	}

	response := ProjectResponse{
		ID:          pgUUIDToStr(project.ID),
		Title:       project.Title,
		Description: project.Description.String,
		CreatedAt:   project.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   project.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	if stats != nil {
		response.Stats = &ProjectStatsDTO{
			DocCount:  stats.DocCount,
			NodeCount: stats.NodeCount,
			EdgeCount: stats.EdgeCount,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateProject handles POST /api/projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	var desc pgtype.Text
	if req.Description != "" {
		desc = pgtype.Text{String: req.Description, Valid: true}
	}

	project, err := h.db.CreateProject(r.Context(), req.Title, desc)
	if err != nil {
		h.logger.Error("failed to create project", "error", err)
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	response := ProjectResponse{
		ID:          pgUUIDToStr(project.ID),
		Title:       project.Title,
		Description: project.Description.String,
		CreatedAt:   project.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   project.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateProjectRequest is the request body for updating a project
type UpdateProjectRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateProject handles PUT /api/projects/{id}
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var desc pgtype.Text
	if req.Description != "" {
		desc = pgtype.Text{String: req.Description, Valid: true}
	}

	project, err := h.db.UpdateProject(r.Context(), strToPgUUID(id.String()), req.Title, desc)
	if err != nil {
		h.logger.Error("failed to update project", "error", err, "id", idStr)
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	response := ProjectResponse{
		ID:          pgUUIDToStr(project.ID),
		Title:       project.Title,
		Description: project.Description.String,
		CreatedAt:   project.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   project.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteProject handles DELETE /api/projects/{id}
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	if err := h.db.DeleteProject(r.Context(), strToPgUUID(id.String())); err != nil {
		h.logger.Error("failed to delete project", "error", err, "id", idStr)
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetProjectDocuments handles GET /api/projects/{id}/documents
func (h *ProjectHandler) GetProjectDocuments(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	sources, err := h.db.GetProjectSources(r.Context(), strToPgUUID(id.String()))
	if err != nil {
		h.logger.Error("failed to get project documents", "error", err, "id", idStr)
		http.Error(w, "Failed to get project documents", http.StatusInternalServerError)
		return
	}

	response := make([]DocumentResponse, len(sources))
	for i, s := range sources {
		response[i] = DocumentResponse{
			ID:           pgUUIDToStr(s.ID),
			Title:        s.Title,
			Filename:     s.Filename,
			FileType:     s.FileType,
			TotalPages:   s.TotalPages.Int32,
			UploadStatus: s.UploadStatus,
			IsDemo:       s.IsDemo,
			CreatedAt:    s.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    s.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddDocumentRequest is the request body for adding a document to a project
type AddDocumentRequest struct {
	DocumentID string `json:"document_id"`
}

// AddDocumentToProject handles POST /api/projects/{id}/documents
func (h *ProjectHandler) AddDocumentToProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.PathValue("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req AddDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	docID, err := uuid.Parse(req.DocumentID)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	_, err = h.db.SetSourceProject(r.Context(), strToPgUUID(docID.String()), strToPgUUID(projectID.String()))
	if err != nil {
		h.logger.Error("failed to add document to project", "error", err, "project", projectIDStr, "document", docID)
		http.Error(w, "Failed to add document to project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GraphResponse represents a merged graph for a project
type GraphResponse struct {
	Nodes []NodeResponse `json:"nodes"`
	Edges []EdgeResponse `json:"edges"`
}

// NodeResponse represents a node in the graph
type NodeResponse struct {
	ID         string                 `json:"id"`
	NodeType   string                 `json:"node_type"`
	Label      string                 `json:"label"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	CreatedAt  string                 `json:"created_at"`
}

// EdgeResponse represents an edge in the graph
type EdgeResponse struct {
	ID         string                 `json:"id"`
	EdgeType   string                 `json:"edge_type"`
	SourceNode string                 `json:"source_node"`
	TargetNode string                 `json:"target_node"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	CreatedAt  string                 `json:"created_at"`
}

// GetProjectGraph handles GET /api/projects/{id}/graph
// Returns merged nodes and edges from all documents in the project
func (h *ProjectHandler) GetProjectGraph(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Get all sources for the project
	sources, err := h.db.GetProjectSources(r.Context(), strToPgUUID(projectID.String()))
	if err != nil {
		h.logger.Error("failed to get project sources", "error", err, "id", idStr)
		http.Error(w, "Failed to get project graph", http.StatusInternalServerError)
		return
	}

	// Collect nodes and edges from all sources
	var allNodes []NodeResponse
	var allEdges []EdgeResponse

	for _, source := range sources {
		// Get nodes for this source
		nodes, err := h.db.ListNodesBySource(r.Context(), strToPgUUID(pgUUIDToStr(source.ID)))
		if err == nil {
			for _, n := range nodes {
				props := make(map[string]interface{})
				if len(n.Properties) > 0 {
					_ = json.Unmarshal(n.Properties, &props)
				}
				allNodes = append(allNodes, NodeResponse{
					ID:         pgUUIDToStr(n.ID),
					NodeType:   n.NodeType,
					Label:      n.Label,
					Properties: props,
					CreatedAt:  n.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				})
			}
		}

		// Get edges for this source
		edges, err := h.db.ListEdgesBySource(r.Context(), strToPgUUID(pgUUIDToStr(source.ID)))
		if err == nil {
			for _, e := range edges {
				props := make(map[string]interface{})
				if len(e.Properties) > 0 {
					_ = json.Unmarshal(e.Properties, &props)
				}
				allEdges = append(allEdges, EdgeResponse{
					ID:         pgUUIDToStr(e.ID),
					EdgeType:   e.EdgeType,
					SourceNode: pgUUIDToStr(e.SourceNode),
					TargetNode: pgUUIDToStr(e.TargetNode),
					Properties: props,
					CreatedAt:  e.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				})
			}
		}
	}

	response := GraphResponse{
		Nodes: allNodes,
		Edges: allEdges,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DocumentResponse is the response for a document
type DocumentResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Filename     string `json:"filename"`
	FileType     string `json:"file_type"`
	TotalPages   int32  `json:"total_pages"`
	UploadStatus string `json:"upload_status"`
	IsDemo       bool   `json:"is_demo"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Helper functions
func pgUUIDToStr(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	uuid, _ := uuid.FromBytes(id.Bytes[:])
	return uuid.String()
}

func strToPgUUID(s string) pgtype.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{
		Bytes: [16]byte(id),
		Valid: true,
	}
}
