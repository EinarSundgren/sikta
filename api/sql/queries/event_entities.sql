-- name: CreateEventEntity :one
INSERT INTO event_entities (event_id, entity_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (event_id, entity_id) DO UPDATE SET role = EXCLUDED.role
RETURNING *;

-- name: ListEventEntities :many
SELECT e.*, ee.role
FROM event_entities ee
JOIN entities e ON ee.entity_id = e.id
WHERE ee.event_id = $1;

-- name: ListEntityEvents :many
SELECT ev.*, ee.role
FROM event_entities ee
JOIN events ev ON ee.event_id = ev.id
WHERE ee.entity_id = $1
ORDER BY ev.narrative_position;
