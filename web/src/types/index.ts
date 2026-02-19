export interface TimelineEvent {
  id: string;
  source_id: string;
  title: string;
  description: string | null;
  event_type: string | null;
  date_text: string | null;
  date_start: string | null;
  date_end: string | null;
  date_precision: string | null;
  narrative_position: number;
  chronological_position: number | null;
  confidence: number;
  confidence_reason: string | null;
  review_status: string;
  metadata: Record<string, any> | null;
  entities: TimelineEntity[];
  source_references: TimelineSource[];
  inconsistencies: TimelineInconsistency[];
  chapter_title: string | null;
  chapter_number: number | null;
}

export interface TimelineEntity {
  id: string;
  name: string;
  entity_type: string;
  role: string | null;
}

export interface TimelineSource {
  id: string;
  chunk_id: string;
  excerpt: string;
  chapter_title: string | null;
  chapter_number: number | null;
  page_start: number | null;
  page_end: number | null;
}

export interface TimelineInconsistency {
  id: string;
  inconsistency_type: string;
  severity: string;
  title: string;
}

export interface Source {
  id: string;
  title: string;
  filename: string;
  file_type: string;
  total_pages: number | null;
  upload_status: string;
  is_demo: boolean;
  source_trust: number | null;
  trust_reason: string | null;
  created_at: string;
  updated_at: string;
}

/** @deprecated Use Source instead */
export type Document = Source;

export interface Entity {
  id: string;
  source_id: string;
  name: string;
  entity_type: string;
  aliases: string[] | null;
  description: string | null;
  first_appearance_chunk: number | null;
  last_appearance_chunk: number | null;
  confidence: number;
  review_status: string;
  created_at: string;
  updated_at: string;
}

export interface Relationship {
  id: string;
  source_id: string;
  entity_a_id: string;
  entity_b_id: string;
  relationship_type: string;
  description: string | null;
  start_claim_id: string | null;
  end_claim_id: string | null;
  confidence: number;
  review_status: string;
  created_at: string;
  updated_at: string;
}
