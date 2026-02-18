-- Junction table linking events to participating entities
CREATE TABLE IF NOT EXISTS event_entities (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    role TEXT CHECK (role IN (
        'actor', 'subject', 'witness', 'mentioned', 'location'
    )),
    PRIMARY KEY (event_id, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_event_entities_event ON event_entities(event_id);
CREATE INDEX IF NOT EXISTS idx_event_entities_entity ON event_entities(entity_id);
CREATE INDEX IF NOT EXISTS idx_event_entities_role ON event_entities(role);
