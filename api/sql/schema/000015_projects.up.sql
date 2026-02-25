-- Projects for grouping multiple documents
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add project_id to sources
ALTER TABLE sources ADD COLUMN project_id UUID REFERENCES projects(id);
CREATE INDEX idx_sources_project ON sources(project_id);
