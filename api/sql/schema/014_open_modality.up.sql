-- Drop the CHECK constraint on provenance.modality.
-- Modality is an open text field — any value is valid.
-- The spec requires open strings, not a closed enum.
ALTER TABLE provenance DROP CONSTRAINT IF EXISTS provenance_modality_check;
