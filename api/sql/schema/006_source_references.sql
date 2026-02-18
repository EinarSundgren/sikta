-- Source references linking extractions to text chunks
CREATE TABLE IF NOT EXISTS source_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chunk_id UUID NOT NULL REFERENCES chunks(id) ON DELETE CASCADE,
    event_id UUID REFERENCES events(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    relationship_id UUID REFERENCES relationships(id) ON DELETE CASCADE,
    excerpt TEXT NOT NULL,
    char_start INTEGER,
    char_end INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (event_id IS NOT NULL)::integer +
        (entity_id IS NOT NULL)::integer +
        (relationship_id IS NOT NULL)::integer = 1
    )
);

CREATE INDEX IF NOT EXISTS idx_source_refs_chunk ON source_references(chunk_id);
CREATE INDEX IF NOT EXISTS idx_source_refs_event ON source_references(event_id);
CREATE INDEX IF NOT EXISTS idx_source_refs_entity ON source_references(entity_id);
CREATE INDEX IF NOT EXISTS idx_source_refs_relationship ON source_references(relationship_id);
