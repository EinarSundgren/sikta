-- Relationships between entities
CREATE TABLE IF NOT EXISTS relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    entity_a_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    entity_b_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL CHECK (relationship_type IN (
        'family', 'romantic', 'social', 'professional', 'adversarial', 'other'
    )),
    description TEXT,
    start_event_id UUID REFERENCES events(id),
    end_event_id UUID REFERENCES events(id),
    confidence REAL NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    review_status TEXT NOT NULL DEFAULT 'pending' CHECK (review_status IN (
        'pending', 'approved', 'rejected', 'edited'
    )),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (document_id, entity_a_id, entity_b_id, relationship_type)
);

CREATE INDEX IF NOT EXISTS idx_relationships_document_id ON relationships(document_id);
CREATE INDEX IF NOT EXISTS idx_relationships_entity_a ON relationships(entity_a_id);
CREATE INDEX IF NOT EXISTS idx_relationships_entity_b ON relationships(entity_b_id);
CREATE INDEX IF NOT EXISTS idx_relationships_type ON relationships(relationship_type);
