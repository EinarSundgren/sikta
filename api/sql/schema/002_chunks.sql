CREATE TABLE IF NOT EXISTS chunks (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id        UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index        INTEGER     NOT NULL,
    content            TEXT        NOT NULL,
    chapter_title      TEXT,
    chapter_number     INTEGER,
    page_start         INTEGER,
    page_end           INTEGER,
    narrative_position INTEGER     NOT NULL,
    word_count         INTEGER,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (document_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS chunks_document_id_idx ON chunks(document_id);
