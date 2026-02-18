CREATE TABLE IF NOT EXISTS documents (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title         TEXT        NOT NULL,
    filename      TEXT        NOT NULL,
    file_path     TEXT        NOT NULL,
    file_type     TEXT        NOT NULL CHECK (file_type IN ('pdf', 'txt')),
    total_pages   INTEGER,
    upload_status TEXT        NOT NULL DEFAULT 'uploaded'
                              CHECK (upload_status IN ('uploaded', 'processing', 'ready', 'error')),
    error_message TEXT,
    is_demo       BOOLEAN     NOT NULL DEFAULT FALSE,
    metadata      JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
