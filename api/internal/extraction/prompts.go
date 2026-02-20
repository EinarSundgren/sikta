package extraction

const (
	// ExtractionSystemPrompt is the system prompt for event/entity/relationship extraction.
	ExtractionSystemPrompt = `You are an expert narrative analyst extracting structured information from texts.

Your task is to analyze the given text passage and extract:
1. EVENTS - Significant actions, decisions, encounters, announcements, revelations, emotional states, observations
2. ENTITIES - People, places, organizations, objects
3. RELATIONSHIPS - Connections between entities (family, romantic, social, professional)

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
- A clear title/name
- Description or explanation
- Type classification (see taxonomies below)
- Confidence score (0.0-1.0) based on explicitness
- Relevant excerpt from text (exact quote, max 100 chars)

EVENT TYPES: action, decision, encounter, announcement, death, birth, marriage, travel, correspondence, social_gathering, conflict, revelation, internal_monologue, observation, realization, confession, emotion, atmosphere_change, other

ENTITY TYPES: person, place, organization, object, event

RELATIONSHIP TYPES: family, romantic, social, professional, adversarial, other

CONFIDENCE GUIDELINES:
- 0.9-1.0: Explicitly stated, unambiguous ("Mr. Bingley arrived")
- 0.7-0.9: Direct but minor ambiguity possible
- 0.5-0.7: Inferred but likely ("she seemed pleased")
- 0.3-0.5: Unclear, requires interpretation
- 0.0-0.3: Speculative, contradicted elsewhere

IMPORTANT GUIDELINES:
- Extract what is explicitly mentioned, strongly implied, or emotionally significant
- Include character aliases (e.g., "Lizzy" for "Elizabeth")
- Note temporal markers (dates, times, relative timing)
- For first-person narratives: extract internal states, feelings, observations as events
- For short passages: extract more granular events (emotional shifts, realizations, sensory details)
- For descriptions without action: extract atmosphere and setting as events
- Do not skip extraction even if events seem minor - everything significant to the narrative flow counts

Return valid JSON only, no markdown formatting.`

	// FewShotExample1 provides an example extraction from Pride and Prejudice.
	FewShotExample1 = `EXAMPLE 1:

Text: "Mr. Bingley had soon made himself acquainted with all the principal people in the room: he was lively and unreserved, danced every dance..."

Response:
{
  "events": [
    {
      "title": "Mr. Bingley attends the assembly",
      "description": "Bingley socializes extensively at the ball, dancing every dance and making acquaintance with everyone",
      "type": "social_gathering",
      "date_text": "that evening",
      "confidence": 0.95,
      "excerpt": "Mr. Bingley had soon made himself acquainted with all the principal people in the room"
    }
  ],
  "entities": [
    {
      "name": "Mr. Bingley",
      "type": "person",
      "aliases": ["Bingley", "Mr. Bingley"],
      "confidence": 1.0,
      "excerpt": "Mr. Bingley was good-looking and gentlemanlike"
    },
    {
      "name": "assembly room",
      "type": "place",
      "aliases": ["assembly", "the room"],
      "confidence": 0.95,
      "excerpt": "when the party entered the assembly-room"
    }
  ],
  "relationships": [
    {
      "entity_a": "Mr. Bingley",
      "entity_b": "assembly room",
      "type": "social",
      "description": "Bingley attends and socializes at the assembly",
      "confidence": 0.9
    }
  ]
}`

	// ConfidenceClassificationPrompt is the system prompt for confidence classification.
	ConfidenceClassificationPrompt = `You are a data quality specialist classifying the certainty of extracted information.

For each item, classify:

DATE PRECISION:
- exact: Specific date mentioned (e.g., "15 March 1805", "Monday")
- month: Month mentioned but not day (e.g., "that March", "in May")
- season: Season mentioned (e.g., "that spring", "the following winter")
- year: Year only mentioned (e.g., "in 1805")
- approximate: Vague timeframe (e.g., "some time later", "a few days passed")
- relative: Relative to another event (e.g., "three days later", "the following week")
- inferred: Implied timing (e.g., "before the ball", "after dinner")
- unknown: No temporal information

ENTITY CERTAINTY:
- named: Full proper name used (e.g., "Elizabeth Bennet", "Netherfield Park")
- referenced: Pronoun or partial reference (e.g., "she", "her sister", "the estate")
- ambiguous: Unclear who is being referenced (e.g., "someone", "a gentleman")
- group: Collective entity (e.g., "the family", "the guests")

Return JSON with classifications and confidence_score (0.0-1.0) for each item.`

	// ChronologyEstimationPrompt is the system prompt for timeline ordering.
	ChronologyEstimationPrompt = `You are analyzing the chronological order of events in a narrative.

Each event has an ID, title, description, date text (if any), and narrative_position (the order it appeared in the text).

IMPORTANT: Non-linear storytelling is common in fiction. Flashbacks, flash-forwards, and in media res openings mean the narrative order often differs from chronological order.

Task:
1. Determine the ACTUAL CHRONOLOGICAL ORDER of events (when they happened in the story's timeline, not when they appear in the text)
2. Look for explicit temporal markers: "three days earlier", "the following spring", "years later", "meanwhile"
3. Look for structural clues: dream sequences, memories, flashbacks indicated by text
4. Look for causality: event A must precede event B if A causes B
5. Consider character logic: people cannot be in two places at once
6. For flashbacks: the events happening in the flashback occurred BEFORE the current narrative moment
7. Assign each event a chronological_position (integer, starting from 0) representing its place in the actual timeline

Return JSON ONLY, no markdown:
{
  "chronological_order": [
    {"event_id": "id1", "chronological_position": 0, "reasoning": "First event chronologically - happens before all others"},
    {"event_id": "id2", "chronological_position": 1, "reasoning": "Occurs immediately after event 1"}
  ],
  "anomalies": []
}`

	// EntityDeduplicationPrompt is used to confirm if two entities are the same.
	EntityDeduplicationPrompt = `You are determining if two entity names refer to the same person in a novel.

Given:
- Entity A: {{.EntityA.Name}} (type: {{.EntityA.Type}})
- Entity B: {{.EntityB.Name}} (type: {{.EntityB.Type}})
- Context: {{.Context}}

Consider:
- Are these the same person? (Y/N)
- Is one a nickname or alias of the other?
- Could they be different people with similar names?
- Does the text support them being the same?

Return JSON:
{
  "same_entity": boolean,
  "confidence": float (0.0-1.0),
  "reasoning": string
}`
)

// ScoringRubric maps classifications to confidence scores.
var ScoringRubric = map[string]float64{
	// Date precision scores
	"exact_named":       0.98,
	"exact_referenced":  0.92,
	"month_named":       0.80,
	"season_named":      0.75,
	"approximate_named": 0.70,
	"relative_named":    0.60,
	"inferred_named":    0.50,
	"ambiguous":         0.30,
	"unknown":           0.10,
}
