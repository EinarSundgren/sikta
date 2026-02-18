-- +migrate Up
-- inconsistencies table stores detected contradictions, temporal impossibilities, and mismatches

CREATE TABLE inconsistencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    inconsistency_type TEXT NOT NULL CHECK (inconsistency_type IN (
        'narrative_chronological_mismatch',
        'temporal_impossibility',
        'contradiction',
        'cross_reference',
        'duplicate_entity',
        'data_mismatch'
    )),
    severity TEXT NOT NULL CHECK (severity IN ('info', 'warning', 'conflict')),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    resolution_status TEXT NOT NULL DEFAULT 'unresolved' CHECK (resolution_status IN (
        'unresolved',
        'resolved',
        'noted',
        'dismissed'
    )),
    resolution_note TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inconsistencies_document_id ON inconsistencies(document_id);
CREATE INDEX idx_inconsistencies_type ON inconsistencies(inconsistency_type);
CREATE INDEX idx_inconsistencies_severity ON inconsistencies(severity);
CREATE INDEX idx_inconsistencies_status ON inconsistencies(resolution_status);

-- +migrate Down
DROP INDEX IF EXISTS idx_inconsistencies_status;
DROP INDEX IF EXISTS idx_inconsistencies_severity;
DROP INDEX IF EXISTS idx_inconsistencies_type;
DROP INDEX IF EXISTS idx_inconsistencies_document_id;
DROP TABLE IF EXISTS inconsistencies;
