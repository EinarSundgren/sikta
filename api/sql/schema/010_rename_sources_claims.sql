-- Migration: rename documents→sources, events→claims throughout the schema.
-- Rationale: "sources" is more accurate (any ingested material, not just documents).
-- "claims" captures that extractions are assertions, not ground truth.
-- Adds: claim_type discriminator, source_trust (Level 1 confidence).

-- 1. Rename tables
ALTER TABLE documents RENAME TO sources;
ALTER TABLE events RENAME TO claims;
ALTER TABLE event_entities RENAME TO claim_entities;

-- 2. Rename document_id → source_id in all tables
ALTER TABLE chunks RENAME COLUMN document_id TO source_id;
ALTER TABLE claims RENAME COLUMN document_id TO source_id;
ALTER TABLE entities RENAME COLUMN document_id TO source_id;
ALTER TABLE relationships RENAME COLUMN document_id TO source_id;
ALTER TABLE inconsistencies RENAME COLUMN document_id TO source_id;

-- 3. Rename event_id → claim_id in all tables
ALTER TABLE source_references RENAME COLUMN event_id TO claim_id;
ALTER TABLE claim_entities RENAME COLUMN event_id TO claim_id;
ALTER TABLE inconsistency_items RENAME COLUMN event_id TO claim_id;

-- 4. Rename start/end event references in relationships
ALTER TABLE relationships RENAME COLUMN start_event_id TO start_claim_id;
ALTER TABLE relationships RENAME COLUMN end_event_id TO end_claim_id;

-- 5. Add claim_type to claims (discriminator: 'event', 'attribute', 'relation')
ALTER TABLE claims ADD COLUMN IF NOT EXISTS claim_type TEXT NOT NULL DEFAULT 'event';

-- 6. Drop the overly restrictive event_type CHECK constraint (will be enforced at application level)
ALTER TABLE claims DROP CONSTRAINT IF EXISTS events_event_type_check;
-- Make event_type nullable: only required when claim_type = 'event'
ALTER TABLE claims ALTER COLUMN event_type DROP NOT NULL;

-- 7. Add source trust columns to sources (Level 1 confidence)
ALTER TABLE sources ADD COLUMN IF NOT EXISTS source_trust REAL;
ALTER TABLE sources ADD COLUMN IF NOT EXISTS trust_reason TEXT;

-- 8. Rename indexes for consistency
ALTER INDEX IF EXISTS idx_events_document_id RENAME TO idx_claims_source_id;
ALTER INDEX IF EXISTS idx_events_chronological RENAME TO idx_claims_chronological;
ALTER INDEX IF EXISTS idx_events_narrative RENAME TO idx_claims_narrative;
ALTER INDEX IF EXISTS idx_events_confidence RENAME TO idx_claims_confidence;
ALTER INDEX IF EXISTS idx_events_type RENAME TO idx_claims_type;
ALTER INDEX IF EXISTS idx_entities_document_id RENAME TO idx_entities_source_id;
ALTER INDEX IF EXISTS idx_relationships_document_id RENAME TO idx_relationships_source_id;
ALTER INDEX IF EXISTS idx_inconsistencies_document_id RENAME TO idx_inconsistencies_source_id;
ALTER INDEX IF EXISTS idx_source_refs_event RENAME TO idx_source_refs_claim;
ALTER INDEX IF EXISTS idx_event_entities_event RENAME TO idx_claim_entities_claim;
ALTER INDEX IF EXISTS idx_event_entities_entity RENAME TO idx_claim_entities_entity;
ALTER INDEX IF EXISTS idx_event_entities_role RENAME TO idx_claim_entities_role;
ALTER INDEX IF EXISTS chunks_document_id_idx RENAME TO chunks_source_id_idx;
