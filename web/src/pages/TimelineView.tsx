import React, { useEffect, useState, useMemo, useRef } from 'react';
import Timeline from '../components/timeline/Timeline';
import EntityPanel from '../components/entities/EntityPanel';
import RelationshipGraph from '../components/graph/RelationshipGraph';
import { timelineApi } from '../api/timeline';
import { TimelineEvent, Entity, Relationship } from '../types';

type ActiveView = 'timeline' | 'graph';

const TimelineView: React.FC = () => {
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [entities, setEntities] = useState<Entity[]>([]);
  const [relationships, setRelationships] = useState<Relationship[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedEvent, setSelectedEvent] = useState<TimelineEvent | null>(null);
  const [selectedEntityId, setSelectedEntityId] = useState<string | null>(null);
  const [activeView, setActiveView] = useState<ActiveView>('timeline');
  const mainRef = useRef<HTMLDivElement>(null);
  const [mainWidth, setMainWidth] = useState(900);

  // Load data
  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        const res = await fetch('/api/documents');
        if (!res.ok) throw new Error('Failed to fetch documents');
        const docs = await res.json();
        if (!docs || docs.length === 0) throw new Error('No documents found');
        const documentId = docs[0].id;

        const [eventsData, entitiesData, relationshipsData] = await Promise.all([
          timelineApi.getTimeline(documentId),
          timelineApi.getEntities(documentId),
          timelineApi.getRelationships(documentId),
        ]);

        setEvents(eventsData);
        setEntities(entitiesData);
        setRelationships(relationshipsData);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  // Measure main content area for D3 components
  useEffect(() => {
    if (!mainRef.current) return;
    const observer = new ResizeObserver(entries => {
      for (const entry of entries) {
        setMainWidth(entry.contentRect.width);
      }
    });
    observer.observe(mainRef.current);
    return () => observer.disconnect();
  }, [loading]);

  // Filter events by selected entity.
  // Since claim_entities table may be empty, fall back to name matching in title/description.
  const selectedEntity = useMemo(
    () => entities.find(e => e.id === selectedEntityId) ?? null,
    [entities, selectedEntityId]
  );

  const filteredEvents = useMemo(() => {
    if (!selectedEntityId || !selectedEntity) return events;

    const terms = [
      selectedEntity.name.toLowerCase(),
      ...(selectedEntity.aliases || []).map(a => a.toLowerCase()),
    ];

    return events.filter(event => {
      // Prefer explicit entity associations if available
      if (event.entities?.some(e => e.id === selectedEntityId)) return true;
      // Fall back to text matching
      const title = event.title.toLowerCase();
      const desc = (event.description || '').toLowerCase();
      return terms.some(term => title.includes(term) || desc.includes(term));
    });
  }, [events, selectedEntityId, selectedEntity]);

  const handleEntitySelect = (id: string | null) => {
    setSelectedEntityId(id);
    setSelectedEvent(null);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-slate-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600 mx-auto mb-3" />
          <p className="text-slate-500 text-sm">Loading timeline...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen bg-slate-50">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen bg-slate-100 overflow-hidden">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 shadow-sm shrink-0">
        <div className="px-5 py-3 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-slate-900 leading-tight">Sikta</h1>
            <p className="text-xs text-slate-500">Document Timeline Intelligence</p>
          </div>

          {/* View tabs */}
          <div className="flex items-center gap-1 bg-slate-100 rounded-lg p-1">
            <button
              onClick={() => setActiveView('timeline')}
              className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
                activeView === 'timeline'
                  ? 'bg-white text-slate-900 shadow-sm'
                  : 'text-slate-500 hover:text-slate-700'
              }`}
            >
              Timeline
            </button>
            <button
              onClick={() => setActiveView('graph')}
              className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
                activeView === 'graph'
                  ? 'bg-white text-slate-900 shadow-sm'
                  : 'text-slate-500 hover:text-slate-700'
              }`}
            >
              Graph
            </button>
          </div>

          <div className="text-right">
            <p className="text-sm font-medium text-slate-800">Pride and Prejudice</p>
            <p className="text-xs text-slate-500">
              {selectedEntityId
                ? `${filteredEvents.length} of ${events.length} events`
                : `${events.length} events · ${entities.length} entities`}
            </p>
          </div>
        </div>
      </header>

      {/* Body: entity panel + main content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Entity sidebar */}
        <EntityPanel
          entities={entities}
          relationships={relationships}
          selectedEntityId={selectedEntityId}
          onEntitySelect={handleEntitySelect}
        />

        {/* Main content */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {activeView === 'timeline' ? (
            <>
              {/* Legend bar */}
              <div className="bg-white border-b border-slate-200 px-5 py-2 shrink-0 flex items-center gap-6 text-xs text-slate-600">
                <div className="flex items-center gap-1.5">
                  <div className="w-3 h-3 rounded-full bg-blue-500" />
                  <span>Chronological</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <div className="w-3 h-3 rounded-full bg-purple-500" />
                  <span>Narrative</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span>⚡</span>
                  <span>Inconsistency</span>
                </div>
                {selectedEntity && (
                  <div className="ml-auto flex items-center gap-1.5 text-blue-600 font-medium">
                    <span>👤</span>
                    <span>Filtered: {selectedEntity.name}</span>
                  </div>
                )}
              </div>

              {/* Timeline */}
              <div ref={mainRef} className="flex-1 overflow-auto p-4">
                <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-4">
                  <Timeline
                    events={filteredEvents}
                    onEventClick={setSelectedEvent}
                    width={Math.max(mainWidth - 48, 600)}
                    height={420}
                  />
                </div>

                {/* Event detail panel */}
                {selectedEvent && (
                  <div className="mt-4 bg-white rounded-lg shadow-sm border border-slate-200 p-5">
                    <div className="flex items-start justify-between mb-3">
                      <h2 className="text-lg font-bold text-slate-900 pr-4">
                        {selectedEvent.title}
                      </h2>
                      <button
                        onClick={() => setSelectedEvent(null)}
                        className="text-slate-400 hover:text-slate-600 shrink-0 text-lg leading-none"
                      >
                        ✕
                      </button>
                    </div>
                    {selectedEvent.description && (
                      <p className="text-slate-700 text-sm mb-4 leading-relaxed">
                        {selectedEvent.description}
                      </p>
                    )}
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 text-sm mb-3">
                      {selectedEvent.event_type && (
                        <div>
                          <span className="font-medium text-slate-500 text-xs uppercase tracking-wide">Type</span>
                          <p className="text-slate-800">{selectedEvent.event_type}</p>
                        </div>
                      )}
                      <div>
                        <span className="font-medium text-slate-500 text-xs uppercase tracking-wide">Confidence</span>
                        <p className="text-slate-800">{Math.round(selectedEvent.confidence * 100)}%</p>
                      </div>
                      {selectedEvent.date_text && (
                        <div>
                          <span className="font-medium text-slate-500 text-xs uppercase tracking-wide">Date</span>
                          <p className="text-slate-800">{selectedEvent.date_text}</p>
                        </div>
                      )}
                      <div>
                        <span className="font-medium text-slate-500 text-xs uppercase tracking-wide">Narrative pos.</span>
                        <p className="text-slate-800">{selectedEvent.narrative_position}</p>
                      </div>
                      {selectedEvent.chronological_position !== null && (
                        <div>
                          <span className="font-medium text-slate-500 text-xs uppercase tracking-wide">Chronological pos.</span>
                          <p className="text-slate-800">{selectedEvent.chronological_position}</p>
                        </div>
                      )}
                    </div>

                    {selectedEvent.entities && selectedEvent.entities.length > 0 && (
                      <div className="mb-3">
                        <span className="font-medium text-slate-500 text-xs uppercase tracking-wide block mb-1">Entities</span>
                        <div className="flex flex-wrap gap-1.5">
                          {selectedEvent.entities.map(e => (
                            <button
                              key={e.id}
                              onClick={() => handleEntitySelect(e.id)}
                              className="px-2.5 py-1 bg-slate-100 hover:bg-blue-100 text-slate-700 hover:text-blue-700 rounded text-xs transition-colors"
                            >
                              {e.name}
                            </button>
                          ))}
                        </div>
                      </div>
                    )}

                    {selectedEvent.inconsistencies && selectedEvent.inconsistencies.length > 0 && (
                      <div className="pt-3 border-t border-slate-100">
                        <span className="font-medium text-slate-500 text-xs uppercase tracking-wide block mb-1">Inconsistencies</span>
                        {selectedEvent.inconsistencies.map(inc => (
                          <div key={inc.id} className="flex items-center gap-2 text-sm py-0.5">
                            <span className="text-amber-500">⚡</span>
                            <span className="text-slate-700">{inc.title}</span>
                            <span className={`ml-auto px-2 py-0.5 rounded text-xs font-medium ${
                              inc.severity === 'conflict'
                                ? 'bg-red-100 text-red-700'
                                : inc.severity === 'warning'
                                ? 'bg-amber-100 text-amber-700'
                                : 'bg-slate-100 text-slate-600'
                            }`}>
                              {inc.severity}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </>
          ) : (
            /* Graph view */
            <div ref={mainRef} className="flex-1 p-4 overflow-hidden">
              <div className="w-full h-full rounded-lg overflow-hidden shadow-sm border border-slate-200">
                <RelationshipGraph
                  entities={entities}
                  relationships={relationships}
                  selectedEntityId={selectedEntityId}
                  onEntitySelect={handleEntitySelect}
                />
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default TimelineView;
