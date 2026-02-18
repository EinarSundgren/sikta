-- Phase 0: minimal schema for sqlc verification.
-- Real schema (documents, chunks, events, etc.) added in Phase 1.

CREATE TABLE IF NOT EXISTS schema_versions (
    id         SERIAL PRIMARY KEY,
    version    TEXT        NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
