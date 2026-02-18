-- +migrate Up
-- inconsistency_items junction table links inconsistencies to events and entities

CREATE TABLE inconsistency_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inconsistency_id UUID NOT NULL REFERENCES inconsistencies(id) ON DELETE CASCADE,
    event_id UUID REFERENCES events(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    side TEXT CHECK (side IN ('a', 'b', 'neutral')),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (event_id IS NOT NULL)::integer +
        (entity_id IS NOT NULL)::integer >= 1
    )
);

CREATE INDEX idx_inconsistency_items_inconsistency ON inconsistency_items(inconsistency_id);
CREATE INDEX idx_inconsistency_items_event ON inconsistency_items(event_id);
CREATE INDEX idx_inconsistency_items_entity ON inconsistency_items(entity_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_inconsistency_items_entity;
DROP INDEX IF EXISTS idx_inconsistency_items_event;
DROP INDEX IF EXISTS idx_inconsistency_items_inconsistency;
DROP TABLE IF EXISTS inconsistency_items;
