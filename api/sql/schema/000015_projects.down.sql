-- Remove project support
DROP INDEX IF EXISTS idx_sources_project;
ALTER TABLE sources DROP COLUMN IF EXISTS project_id;
DROP TABLE IF EXISTS projects;
