package extraction

import (
	"encoding/json"
	"sync"
	"time"
)

// ProgressState tracks the current state of an extraction.
type ProgressState struct {
	SourceID        string `json:"source_id"`
	Status          string `json:"status"` // "idle", "processing", "complete", "error"
	CurrentChunk    int    `json:"current_chunk"`
	TotalChunks     int    `json:"total_chunks"`
	EventsFound     int    `json:"events_found"`
	EntitiesFound   int    `json:"entities_found"`
	RelationsFound  int    `json:"relationships_found"`
	ErrorMessage    string `json:"error_message,omitempty"`
	PercentComplete int    `json:"percent_complete"`
	ElapsedTimeSec  int    `json:"elapsed_time_sec"`
}

// ProgressEvent is sent to SSE subscribers.
type ProgressEvent struct {
	Type    string        `json:"type"` // "progress", "complete", "error"
	Payload ProgressState `json:"payload"`
}

// ProgressTracker manages extraction progress state and subscribers.
type ProgressTracker struct {
	mu          sync.RWMutex
	states      map[string]*ProgressState
	startTimes  map[string]time.Time
	subscribers map[string]map[chan ProgressEvent]struct{}
}

// NewProgressTracker creates a new progress tracker.
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		states:      make(map[string]*ProgressState),
		startTimes:  make(map[string]time.Time),
		subscribers: make(map[string]map[chan ProgressEvent]struct{}),
	}
}

// Start marks an extraction as started.
func (t *ProgressTracker) Start(sourceID string, totalChunks int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	t.startTimes[sourceID] = now
	t.states[sourceID] = &ProgressState{
		SourceID:       sourceID,
		Status:         "processing",
		CurrentChunk:   0,
		TotalChunks:    totalChunks,
		ElapsedTimeSec: 0,
	}
	t.broadcast(sourceID)
}

// Update updates the progress for an extraction.
func (t *ProgressTracker) Update(sourceID string, currentChunk, events, entities, relationships int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state, ok := t.states[sourceID]
	if !ok {
		return
	}

	state.CurrentChunk = currentChunk
	state.EventsFound = events
	state.EntitiesFound = entities
	state.RelationsFound = relationships

	if state.TotalChunks > 0 {
		state.PercentComplete = (currentChunk * 100) / state.TotalChunks
		if state.PercentComplete > 95 {
			state.PercentComplete = 95
		}
	}

	if startTime, ok := t.startTimes[sourceID]; ok {
		state.ElapsedTimeSec = int(time.Since(startTime).Seconds())
	}

	t.broadcast(sourceID)
}

// Complete marks an extraction as complete.
func (t *ProgressTracker) Complete(sourceID string, events, entities, relationships int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state, ok := t.states[sourceID]
	if !ok {
		return
	}

	state.Status = "complete"
	state.PercentComplete = 100
	state.EventsFound = events
	state.EntitiesFound = entities
	state.RelationsFound = relationships

	if startTime, ok := t.startTimes[sourceID]; ok {
		state.ElapsedTimeSec = int(time.Since(startTime).Seconds())
	}

	t.broadcast(sourceID)
	delete(t.startTimes, sourceID)
}

// Error marks an extraction as failed.
func (t *ProgressTracker) Error(sourceID string, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state, ok := t.states[sourceID]
	if !ok {
		return
	}

	state.Status = "error"
	state.ErrorMessage = errMsg

	if startTime, ok := t.startTimes[sourceID]; ok {
		state.ElapsedTimeSec = int(time.Since(startTime).Seconds())
	}

	t.broadcast(sourceID)
	delete(t.startTimes, sourceID)
}

// Get returns the current state for an extraction.
func (t *ProgressTracker) Get(sourceID string) *ProgressState {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, ok := t.states[sourceID]
	if !ok {
		return &ProgressState{SourceID: sourceID, Status: "idle"}
	}

	// Update elapsed time for ongoing extractions
	if state.Status == "processing" {
		if startTime, ok := t.startTimes[sourceID]; ok {
			elapsed := int(time.Since(startTime).Seconds())
			stateCopy := *state
			stateCopy.ElapsedTimeSec = elapsed
			return &stateCopy
		}
	}

	stateCopy := *state
	return &stateCopy
}

// Subscribe registers a channel to receive progress updates.
func (t *ProgressTracker) Subscribe(sourceID string) chan ProgressEvent {
	t.mu.Lock()
	defer t.mu.Unlock()

	ch := make(chan ProgressEvent, 10)
	if t.subscribers[sourceID] == nil {
		t.subscribers[sourceID] = make(map[chan ProgressEvent]struct{})
	}
	t.subscribers[sourceID][ch] = struct{}{}

	// Send current state immediately if exists
	if state, ok := t.states[sourceID]; ok {
		go func() {
			event := ProgressEvent{Type: state.Status, Payload: *state}
			select {
			case ch <- event:
			default:
			}
		}()
	}

	return ch
}

// Unsubscribe removes a subscriber.
func (t *ProgressTracker) Unsubscribe(sourceID string, ch chan ProgressEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if subs, ok := t.subscribers[sourceID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(t.subscribers, sourceID)
		}
	}
	close(ch)
}

// broadcast sends the current state to all subscribers.
func (t *ProgressTracker) broadcast(sourceID string) {
	state, ok := t.states[sourceID]
	if !ok {
		return
	}

	event := ProgressEvent{Type: state.Status, Payload: *state}
	for ch := range t.subscribers[sourceID] {
		select {
		case ch <- event:
		default:
		}
	}
}

// ToJSON returns the state as JSON.
func (s *ProgressState) ToJSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}
