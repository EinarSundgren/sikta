import React, { useEffect, useState } from 'react';
import Timeline from '../components/timeline/Timeline';
import { timelineApi } from '../api/timeline';
import { TimelineEvent } from '../types';

const TimelineView: React.FC = () => {
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedEvent, setSelectedEvent] = useState<TimelineEvent | null>(null);

  // Hardcoded document ID for Pride and Prejudice
  const DOCUMENT_ID = '01234567-89ab-cdef-0123-456789abcdef';

  useEffect(() => {
    const loadTimeline = async () => {
      try {
        setLoading(true);
        const data = await timelineApi.getTimeline(DOCUMENT_ID);
        setEvents(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load timeline');
      } finally {
        setLoading(false);
      }
    };

    loadTimeline();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-slate-600">Loading timeline...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <p className="text-red-600 mb-4">Error: {error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-slate-900">Sikta</h1>
              <p className="text-sm text-slate-600">Document Timeline Intelligence</p>
            </div>
            <div className="flex items-center gap-4">
              <div className="text-right">
                <p className="text-sm font-medium text-slate-900">Pride and Prejudice</p>
                <p className="text-xs text-slate-600">{events.length} events</p>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Legend */}
        <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-4 mb-6">
          <div className="flex items-center justify-between flex-wrap gap-4">
            <div className="flex items-center gap-6 text-sm">
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 rounded-full bg-blue-500"></div>
                <span className="text-slate-700">Chronological</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 rounded-full bg-purple-500"></div>
                <span className="text-slate-700">Narrative</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-lg">⚡</span>
                <span className="text-slate-700">Inconsistency</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full border-2 border-green-500"></div>
                <span className="text-slate-700">High confidence</span>
              </div>
            </div>
            <div className="text-sm text-slate-600">
              Click on events to see details
            </div>
          </div>
        </div>

        {/* Timeline */}
        <div className="bg-white rounded-lg shadow-lg border border-slate-200 p-6 mb-6">
          <Timeline
            events={events}
            onEventClick={setSelectedEvent}
            width={1100}
            height={500}
          />
        </div>

        {/* Event Details Panel */}
        {selectedEvent && (
          <div className="bg-white rounded-lg shadow-lg border border-slate-200 p-6">
            <div className="flex items-start justify-between mb-4">
              <h2 className="text-xl font-bold text-slate-900">{selectedEvent.title}</h2>
              <button
                onClick={() => setSelectedEvent(null)}
                className="text-slate-400 hover:text-slate-600"
              >
                ✕
              </button>
            </div>
            {selectedEvent.description && (
              <p className="text-slate-700 mb-4">{selectedEvent.description}</p>
            )}
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="font-medium text-slate-900">Type:</span>{' '}
                <span className="text-slate-700">{selectedEvent.event_type}</span>
              </div>
              <div>
                <span className="font-medium text-slate-900">Confidence:</span>{' '}
                <span className="text-slate-700">{Math.round(selectedEvent.confidence * 100)}%</span>
              </div>
              {selectedEvent.date_text && (
                <div>
                  <span className="font-medium text-slate-900">Date:</span>{' '}
                  <span className="text-slate-700">{selectedEvent.date_text}</span>
                </div>
              )}
              <div>
                <span className="font-medium text-slate-900">Narrative Position:</span>{' '}
                <span className="text-slate-700">{selectedEvent.narrative_position}</span>
              </div>
              {selectedEvent.chronological_position !== null && (
                <div>
                  <span className="font-medium text-slate-900">Chronological Position:</span>{' '}
                  <span className="text-slate-700">{selectedEvent.chronological_position}</span>
                </div>
              )}
            </div>
            {selectedEvent.inconsistencies && selectedEvent.inconsistencies.length > 0 && (
              <div className="mt-4 pt-4 border-t border-slate-200">
                <h3 className="font-medium text-slate-900 mb-2">Inconsistencies:</h3>
                {selectedEvent.inconsistencies.map(inc => (
                  <div key={inc.id} className="flex items-center gap-2 text-sm mb-1">
                    <span className="text-red-600">⚡</span>
                    <span className="text-slate-700">{inc.title}</span>
                    <span className="px-2 py-0.5 bg-red-100 text-red-700 rounded text-xs">
                      {inc.severity}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  );
};

export default TimelineView;
