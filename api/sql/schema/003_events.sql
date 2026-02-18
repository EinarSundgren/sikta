-- Extracted events from document chunks
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    event_type TEXT NOT NULL CHECK (event_type IN (
        'action', 'decision', 'encounter', 'announcement', 'death',
        'birth', 'marriage', 'travel', 'correspondence',
        'social_gathering', 'conflict', 'revelation', 'other'
    )),
    date_text TEXT,
    date_start DATE,
    date_end DATE,
    date_precision TEXT CHECK (date_precision IN (
        'exact', 'month', 'season', 'year', 'approximate', 'relative', 'inferred', 'unknown'
    )),
    chronological_position INTEGER,
    narrative_position INTEGER NOT NULL,
    confidence REAL NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    confidence_reason TEXT,
    review_status TEXT NOT NULL DEFAULT 'pending' CHECK (review_status IN (
        'pending', 'approved', 'rejected', 'edited'
    )),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_document_id ON events(document_id);
CREATE INDEX IF NOT EXISTS idx_events_chronological ON events(chronological_position);
CREATE INDEX IF NOT EXISTS idx_events_narrative ON events(narrative_position);
CREATE INDEX IF NOT EXISTS idx_events_confidence ON events(confidence);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
