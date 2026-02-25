-- Restore the CHECK constraint on provenance.modality.
ALTER TABLE provenance ADD CONSTRAINT provenance_modality_check
CHECK (modality IN (
    'asserted',
    'hypothetical',
    'denied',
    'conditional',
    'inferred',
    'obligatory',
    'permitted'
));
