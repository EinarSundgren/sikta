-- Migration: Graph primitives (Node, Edge, Provenance)
-- This creates the universal three-primitive model alongside existing tables.
-- Tables coexist with legacy tables during migration period.

-- Nodes: anything in the evidence graph
CREATE TABLE IF NOT EXISTS nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_type TEXT NOT NULL,
    label TEXT NOT NULL,
    properties JSONB DEFAULT '{}',

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Edges: directed connections between nodes
CREATE TABLE IF NOT EXISTS edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    edge_type TEXT NOT NULL,
    source_node UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    target_node UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    properties JSONB DEFAULT '{}',
    is_negated BOOLEAN NOT NULL DEFAULT FALSE,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Provenance: where any node/edge came from
CREATE TABLE IF NOT EXISTS provenance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type TEXT NOT NULL CHECK (target_type IN ('node', 'edge')),
    target_id UUID NOT NULL,
    source_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    excerpt TEXT NOT NULL,
    location JSONB DEFAULT '{}',

    -- Confidence and trust
    confidence REAL NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    trust REAL NOT NULL CHECK (trust >= 0 AND trust <= 1),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'edited')),

    -- Modality: what kind of claim is this?
    modality TEXT NOT NULL DEFAULT 'asserted' CHECK (modality IN (
        'asserted',      -- straightforward assertion
        'hypothetical',  -- conditional, speculative
        'denied',        -- explicitly contradicted
        'conditional',   -- if-then claim
        'inferred',      -- derived from other claims
        'obligatory',    -- must be true
        'permitted'      -- may be true
    )),

    -- Temporal claims (moved from nodes/edges)
    claimed_time_start TIMESTAMPTZ,
    claimed_time_end TIMESTAMPTZ,
    claimed_time_text TEXT,

    -- Geospatial claims
    claimed_geo_region TEXT,
    claimed_geo_text TEXT,

    -- Attribution: who made this claim (for human decisions)
    claimed_by UUID REFERENCES nodes(id),

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add updated_at trigger function for nodes
CREATE OR REPLACE FUNCTION update_nodes_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER nodes_updated_at
    BEFORE UPDATE ON nodes
    FOR EACH ROW
    EXECUTE FUNCTION update_nodes_updated_at();

-- Add updated_at trigger function for edges
CREATE OR REPLACE FUNCTION update_edges_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER edges_updated_at
    BEFORE UPDATE ON edges
    FOR EACH ROW
    EXECUTE FUNCTION update_edges_updated_at();

-- Add updated_at trigger function for provenance
CREATE OR REPLACE FUNCTION update_provenance_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER provenance_updated_at
    BEFORE UPDATE ON provenance
    FOR EACH ROW
    EXECUTE FUNCTION update_provenance_updated_at();
