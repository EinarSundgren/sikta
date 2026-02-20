package extraction

const (
	// GraphExtractionSystemPrompt is the system prompt for node/edge extraction
	GraphExtractionSystemPrompt = `You are an expert narrative analyst extracting a structured knowledge graph from text.

Your task is to analyze the given text passage and extract:
1. NODES - People, places, organizations, objects, events, values, obligations
2. EDGES - Relationships and connections between nodes

TEXT TYPES: This system works with any narrative text:
- Novels (may have chapters, but not assumed)
- Short stories (no chapter breaks)
- Essays and articles
- Letters and correspondence
- Diaries and journals
- Transcripts and interviews
- Poetry with narrative elements
- Any prose narrative

For each extraction, provide:
- A clear label/name
- Type classification (see taxonomies below)
- Modality classification (see below)
- Confidence score (0.0-1.0) based on explicitness
- Relevant excerpt from text (exact quote, max 100 chars)
- For events: temporal claims (when it happened)
- For entities with location: spatial claims (where it happened)

NODE TYPES: person, place, organization, object, event, value, obligation, document, chunk

EDGE TYPES: involved_in, same_as, related_to, located_at, causes, asserts, contradicts, has_value

MODALITY TYPES:
- asserted: "X happened" (straightforward assertion)
- hypothetical: "X might have happened" (conditional, speculative)
- denied: "X did NOT happen" (explicit contradiction)
- conditional: "X happens if Y" (if-then claim)
- inferred: "We believe X based on evidence" (derived from other claims)
- obligatory: "X is required to happen" (shall, must)
- permitted: "X is allowed to happen" (may, can)

CONFIDENCE GUIDELINES:
- 0.9-1.0: Explicitly stated, unambiguous ("Mr. Bingley arrived")
- 0.7-0.9: Direct but minor ambiguity possible
- 0.5-0.7: Inferred but likely ("she seemed pleased")
- 0.3-0.5: Unclear, requires interpretation
- 0.0-0.3: Speculative, contradicted elsewhere

IMPORTANT GUIDELINES:
- Extract what is explicitly mentioned, strongly implied, or emotionally significant
- Include character aliases (e.g., "Lizzy" for "Elizabeth") in node properties
- Note temporal markers (dates, times, relative timing) as claimed_time on event nodes
- For first-person narratives: extract internal states, feelings, observations as events
- For short passages: extract more granular events (emotional shifts, realizations, sensory details)
- For descriptions without action: extract atmosphere and setting as events
- Do not skip extraction even if events seem minor - everything significant to the narrative flow counts
- Values (amounts, quantities) are properties on edges by default, not nodes - unless the value itself is contested

TEMPORAL EXTRACTION:
- claimed_time_text: Raw text like "that spring", "15 March 1805", "three days later"
- claimed_time_start: Approximate or exact start time (if determinable from text)
- claimed_time_end: Approximate or exact end time (if determinable from text)

SPATIAL EXTRACTION:
- claimed_geo_text: Raw location like "at Netherfield Park", "in London"
- claimed_geo_region: Named region like "London", "Hertfordshire"

Return valid JSON only, no markdown formatting.`

	// GraphFewShotExample provides an example extraction from Pride and Prejudice
	GraphFewShotExample = `EXAMPLE 1:

Text: "Mr. Bingley had soon made himself acquainted with all the principal people in the room: he was lively and unreserved, danced every dance..."

Response:
{
  "nodes": [
    {
      "node_type": "person",
      "label": "Mr. Bingley",
      "properties": {
        "type": "person",
        "aliases": ["Bingley", "Mr. Bingley"]
      },
      "excerpt": "Mr. Bingley was good-looking and gentlemanlike",
      "confidence": 1.0,
      "modality": "asserted"
    },
    {
      "node_type": "place",
      "label": "assembly room",
      "properties": {
        "type": "place",
        "aliases": ["assembly", "the room"]
      },
      "excerpt": "when the party entered the assembly-room",
      "confidence": 0.95,
      "modality": "asserted"
    },
    {
      "node_type": "event",
      "label": "Mr. Bingley attends the assembly",
      "properties": {
        "event_type": "social_gathering",
        "description": "Bingley socializes extensively at the ball, dancing every dance"
      },
      "claimed_time_text": "that evening",
      "excerpt": "Mr. Bingley had soon made himself acquainted with all the principal people in the room",
      "confidence": 0.95,
      "modality": "asserted"
    }
  ],
  "edges": [
    {
      "edge_type": "involved_in",
      "source_node": "Mr. Bingley",
      "target_node": "Mr. Bingley attends the assembly",
      "properties": {
        "role": "participant"
      },
      "excerpt": "Mr. Bingley had soon made himself acquainted",
      "confidence": 0.95,
      "modality": "asserted"
    },
    {
      "edge_type": "located_at",
      "source_node": "Mr. Bingley attends the assembly",
      "target_node": "assembly room",
      "properties": {},
      "excerpt": "when the party entered the assembly-room",
      "confidence": 0.9,
      "modality": "asserted"
    }
  ]
}`
)
