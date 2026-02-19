import React, { useEffect, useState, useMemo, useRef, useCallback } from 'react';
import Timeline from '../components/timeline/Timeline';
import EntityPanel from '../components/entities/EntityPanel';
import RelationshipGraph from '../components/graph/RelationshipGraph';
import ReviewPanel from '../components/review/ReviewPanel';
import InconsistencyPanel from '../components/inconsistencies/InconsistencyPanel';
import { timelineApi } from '../api/timeline';
import { TimelineEvent, Entity, Relationship, Inconsistency, ReviewProgress } from '../types';

type ActiveView = 'timeline' | 'graph' | 'review' | 'inconsistencies';

interface TimelineViewProps {
  docId?: string;
  onNavigateHome?: () => void;
}

const TimelineView: React.FC<TimelineViewProps> = ({ docId: propDocId, onNavigateHome }) => {
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [entities, setEntities] = useState<Entity[]>([]);
  const [relationships, setRelationships] = useState<Relationship[]>([]);
  const [inconsistencies, setInconsistencies] = useState<Inconsistency[]>([]);
  const [progress, setProgress] = useState<ReviewProgress | null>(null);
  const [documentId, setDocumentId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedEvent, setSelectedEvent] = useState<TimelineEvent | null>(null);
  const [selectedEntityId, setSelectedEntityId] = useState<string | null>(null);
  const [highlightedClaimIds, setHighlightedClaimIds] = useState<string[]>([]);
  const [activeView, setActiveView] = useState<ActiveView>('timeline');
  const mainRef = useRef<HTMLDivElement>(null);
  const [mainWidth, setMainWidth] = useState(900);

  // Load all data
  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        let docId: string;
        if (propDocId) {
          docId = propDocId;
        } else {
          const res = await fetch('/api/documents');
          if (!res.ok) throw new Error('Failed to fetch documents');
          const docs = await res.json();
          if (!docs || docs.length === 0) throw new Error('No documents found');
          docId = docs[0].id;
        }
        setDocumentId(docId);

        const [eventsData, entitiesData, relationshipsData, inconsistenciesData, progressData] =
          await Promise.all([
            timelineApi.getTimeline(docId),
            timelineApi.getEntities(docId),
            timelineApi.getRelationships(docId),
            timelineApi.getInconsistencies(docId),
            timelineApi.getReviewProgress(docId),
          ]);

        setEvents(eventsData);
        setEntities(entitiesData);
        setRelationships(relationshipsData);
        setInconsistencies(inconsistenciesData);
        setProgress(progressData);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [propDocId]);

  // Refresh progress after review actions
  const refreshProgress = useCallback(async () => {
    if (!documentId) return;
    try {
      const p = await timelineApi.getReviewProgress(documentId);
      setProgress(p);
    } catch {
      // non-fatal
    }
  }, [documentId]);

  // Handle review status / data updates optimistically
  const handleEventUpdated = useCallback((id: string, newStatus: string, newData?: Partial<TimelineEvent>) => {
    setEvents(prev =>
      prev.map(e =>
        e.id === id ? { ...e, review_status: newStatus, ...newData } : e
      )
    );
    if (selectedEvent?.id === id) {
      setSelectedEvent(prev => prev ? { ...prev, review_status: newStatus, ...newData } : null);
    }
  }, [selectedEvent]);

  // Handle inconsistency resolution
  const handleInconsistencyUpdated = useCallback((id: string, status: string, note: string) => {
    setInconsistencies(prev =>
      prev.map(i =>
        i.id === id ? { ...i, resolution_status: status, resolution_note: note || null } : i
      )
    );
  }, []);

  // Highlight claims on timeline from inconsistency panel
  const handleHighlightClaims = useCallback((claimIds: string[]) => {
    setHighlightedClaimIds(claimIds);
    setActiveView('timeline');
  }, []);

  // Measure main content area
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

  // Entity selection
  const selectedEntity = useMemo(
    () => entities.find(e => e.id === selectedEntityId) ?? null,
    [entities, selectedEntityId]
  );

  // Filter events by selected entity (text-match fallback since claim_entities is empty)
  const filteredEvents = useMemo(() => {
    if (!selectedEntityId || !selectedEntity) return events;
    const terms = [
      selectedEntity.name.toLowerCase(),
      ...(selectedEntity.aliases || []).map(a => a.toLowerCase()),
    ];
    return events.filter(event => {
      if (event.entities?.some(e => e.id === selectedEntityId)) return true;
      const title = event.title.toLowerCase();
      const desc = (event.description || '').toLowerCase();
      return terms.some(term => title.includes(term) || desc.includes(term));
    });
  }, [events, selectedEntityId, selectedEntity]);

  const handleEntitySelect = (id: string | null) => {
    setSelectedEntityId(id);
    setSelectedEvent(null);
    setHighlightedClaimIds([]);
  };

  const pendingCount = events.filter(e => e.review_status === 'pending').length;
  const unresolvedCount = inconsistencies.filter(i => i.resolution_status === 'unresolved').length;

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

  const TABS: { id: ActiveView; label: string; badge?: number }[] = [
    { id: 'timeline', label: 'Timeline' },
    { id: 'graph', label: 'Graph' },
    { id: 'review', label: 'Review', badge: pendingCount },
    { id: 'inconsistencies', label: 'Inconsistencies', badge: unresolvedCount },
  ];

  return (
    <div className="flex flex-col h-screen bg-slate-100 overflow-hidden">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 shadow-sm shrink-0">
        <div className="px-5 py-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            {onNavigateHome && (
              <button
                onClick={onNavigateHome}
                className="text-slate-400 hover:text-slate-600 transition-colors text-sm leading-none"
                title="Back to home"
              >
                ←
              </button>
            )}
            <div>
              <h1 className="text-xl font-bold text-slate-900 leading-tight">Sikta</h1>
              <p className="text-xs text-slate-500">Document Timeline Intelligence</p>
            </div>
          </div>

          {/* View tabs */}
          <div className="flex items-center gap-1 bg-slate-100 rounded-lg p-1">
            {TABS.map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveView(tab.id)}
                className={`relative px-4 py-1.5 rounded text-sm font-medium transition-colors ${
                  activeView === tab.id
                    ? 'bg-white text-slate-900 shadow-sm'
                    : 'text-slate-500 hover:text-slate-700'
                }`}
              >
                {tab.label}
                {tab.badge !== undefined && tab.badge > 0 && (
                  <span className="ml-1.5 px-1.5 py-0.5 bg-amber-500 text-white text-xs rounded-full leading-none">
                    {tab.badge}
                  </span>
                )}
              </button>
            ))}
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
        {/* Entity sidebar — always visible */}
        <EntityPanel
          entities={entities}
          relationships={relationships}
          selectedEntityId={selectedEntityId}
          onEntitySelect={handleEntitySelect}
        />

        {/* Main content area */}
        <div className="flex-1 flex flex-col overflow-hidden">

          {/* ── TIMELINE ── */}
          {activeView === 'timeline' && (
            <>
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

              <div ref={mainRef} className="flex-1 overflow-auto p-4">
                <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-4">
                  <Timeline
                    events={filteredEvents}
                    highlightedIds={highlightedClaimIds}
                    onEventClick={e => {
                      setSelectedEvent(e);
                      setHighlightedClaimIds([]);
                    }}
                    width={Math.max(mainWidth - 48, 600)}
                    height={420}
                  />
                </div>

                {/* Event detail */}
                {selectedEvent && (
                  <div className="mt-4 bg-white rounded-lg shadow-sm border border-slate-200 p-5">
                    <div className="flex items-start justify-between mb-3">
                      <h2 className="text-lg font-bold text-slate-900 pr-4">{selectedEvent.title}</h2>
                      <button
                        onClick={() => setSelectedEvent(null)}
                        className="text-slate-400 hover:text-slate-600 shrink-0 text-lg leading-none"
                      >
                        ✕
                      </button>
                    </div>
                    {selectedEvent.description && (
                      <p className="text-slate-700 text-sm mb-4 leading-relaxed">{selectedEvent.description}</p>
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
                        <span className="font-medium text-slate-500 text-xs uppercase tracking-wide">Status</span>
                        <p className={`capitalize font-medium ${
                          selectedEvent.review_status === 'approved' ? 'text-green-600' :
                          selectedEvent.review_status === 'rejected' ? 'text-red-500' :
                          selectedEvent.review_status === 'edited' ? 'text-blue-600' :
                          'text-slate-500'
                        }`}>
                          {selectedEvent.review_status}
                        </p>
                      </div>
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
                              inc.severity === 'conflict' ? 'bg-red-100 text-red-700' :
                              inc.severity === 'warning' ? 'bg-amber-100 text-amber-700' :
                              'bg-slate-100 text-slate-600'
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
          )}

          {/* ── GRAPH ── */}
          {activeView === 'graph' && (
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

          {/* ── REVIEW ── */}
          {activeView === 'review' && (
            <div ref={mainRef} className="flex-1 overflow-hidden">
              <ReviewPanel
                events={events}
                progress={progress}
                onEventUpdated={handleEventUpdated}
                onProgressRefresh={refreshProgress}
              />
            </div>
          )}

          {/* ── INCONSISTENCIES ── */}
          {activeView === 'inconsistencies' && (
            <div ref={mainRef} className="flex-1 overflow-hidden">
              <InconsistencyPanel
                inconsistencies={inconsistencies}
                onInconsistencyUpdated={handleInconsistencyUpdated}
                onHighlightClaims={handleHighlightClaims}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default TimelineView;
