-- Migration: Indexes for graph primitives
-- Performance indexes for common query patterns

-- Node indexes
CREATE INDEX idx_nodes_type ON nodes(node_type);
CREATE INDEX idx_nodes_created_at ON nodes(created_at DESC);

-- Edge indexes
CREATE INDEX idx_edges_type ON edges(edge_type);
CREATE INDEX idx_edges_source ON edges(source_node);
CREATE INDEX idx_edges_target ON edges(target_node);
CREATE INDEX idx_edges_source_target ON edges(source_node, target_node);

-- Provenance indexes
CREATE INDEX idx_provenance_target ON provenance(target_type, target_id);
CREATE INDEX idx_provenance_source ON provenance(source_id);
CREATE INDEX idx_provenance_status ON provenance(status);
CREATE INDEX idx_provenance_modality ON provenance(modality);
CREATE INDEX idx_provenance_confidence ON provenance(confidence DESC);
CREATE INDEX idx_provenance_trust ON provenance(trust DESC);

-- Composite indexes for common queries
CREATE INDEX idx_provenance_target_status ON provenance(target_type, target_id, status);
CREATE INDEX idx_provenance_source_confidence ON provenance(source_id, confidence DESC);

-- Temporal indexes
CREATE INDEX idx_provenance_time_start ON provenance(claimed_time_start) WHERE claimed_time_start IS NOT NULL;
CREATE INDEX idx_provenance_time_end ON provenance(claimed_time_end) WHERE claimed_time_end IS NOT NULL;

-- Spatial indexes (for future PostGIS integration)
-- CREATE INDEX idx_provenance_geo ON provenance USING GIST(claimed_geo_point) WHERE claimed_geo_point IS NOT NULL;
