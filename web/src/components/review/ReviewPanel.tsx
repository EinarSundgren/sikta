import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { TimelineEvent, ReviewProgress } from '../../types';
import EditModal from './EditModal';
import { timelineApi } from '../../api/timeline';

type ReviewFilter = 'pending' | 'all' | 'approved' | 'rejected' | 'edited';

interface ReviewPanelProps {
  events: TimelineEvent[];
  progress: ReviewProgress | null;
  onEventUpdated: (id: string, newStatus: string, newData?: Partial<TimelineEvent>) => void;
  onProgressRefresh: () => void;
}

const STATUS_COLORS: Record<string, string> = {
  pending: 'bg-slate-100 text-slate-600',
  approved: 'bg-green-100 text-green-700',
  rejected: 'bg-red-100 text-red-700',
  edited: 'bg-blue-100 text-blue-700',
};

const STATUS_ICONS: Record<string, string> = {
  pending: '○',
  approved: '✓',
  rejected: '✗',
  edited: '✎',
};

const ReviewPanel: React.FC<ReviewPanelProps> = ({
  events,
  progress,
  onEventUpdated,
  onProgressRefresh,
}) => {
  const [filter, setFilter] = useState<ReviewFilter>('pending');
  const [cursor, setCursor] = useState(0);
  const [editingEvent, setEditingEvent] = useState<TimelineEvent | null>(null);
  const [saving, setSaving] = useState(false);

  const queue = useMemo(() => {
    const filtered = filter === 'all'
      ? events
      : events.filter(e => e.review_status === filter);
    // Sort by confidence ascending (most uncertain first for pending, otherwise by narrative pos)
    return [...filtered].sort((a, b) =>
      filter === 'pending' || filter === 'all'
        ? a.confidence - b.confidence
        : a.narrative_position - b.narrative_position
    );
  }, [events, filter]);

  // Clamp cursor when queue changes
  useEffect(() => {
    setCursor(c => Math.min(c, Math.max(queue.length - 1, 0)));
  }, [queue.length]);

  const currentEvent = queue[cursor] ?? null;

  const approve = useCallback(async (event: TimelineEvent) => {
    if (saving) return;
    setSaving(true);
    try {
      await timelineApi.updateClaimReview(event.id, 'approved');
      onEventUpdated(event.id, 'approved');
      onProgressRefresh();
    } catch {
      // silently ignore for now
    } finally {
      setSaving(false);
    }
  }, [saving, onEventUpdated, onProgressRefresh]);

  const reject = useCallback(async (event: TimelineEvent) => {
    if (saving) return;
    setSaving(true);
    try {
      await timelineApi.updateClaimReview(event.id, 'rejected');
      onEventUpdated(event.id, 'rejected');
      onProgressRefresh();
    } catch {
      // silently ignore
    } finally {
      setSaving(false);
    }
  }, [saving, onEventUpdated, onProgressRefresh]);

  const handleSaveEdit = useCallback(async (data: Parameters<typeof timelineApi.updateClaimData>[1]) => {
    if (!editingEvent || saving) return;
    setSaving(true);
    try {
      await timelineApi.updateClaimData(editingEvent.id, data);
      onEventUpdated(editingEvent.id, 'edited', {
        title: data.title,
        description: data.description,
        date_text: data.date_text,
        event_type: data.event_type,
        confidence: data.confidence,
        review_status: 'edited',
      });
      onProgressRefresh();
    } catch {
      // silently ignore
    } finally {
      setSaving(false);
      setEditingEvent(null);
    }
  }, [editingEvent, saving, onEventUpdated, onProgressRefresh]);

  // Keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (editingEvent) return;
      // Don't fire when typing in an input
      if (['INPUT', 'TEXTAREA', 'SELECT'].includes((e.target as HTMLElement).tagName)) return;

      switch (e.key) {
        case 'j':
        case 'ArrowDown':
          e.preventDefault();
          setCursor(c => Math.min(c + 1, queue.length - 1));
          break;
        case 'k':
        case 'ArrowUp':
          e.preventDefault();
          setCursor(c => Math.max(c - 1, 0));
          break;
        case 'a':
          if (currentEvent) approve(currentEvent);
          break;
        case 'r':
          if (currentEvent) reject(currentEvent);
          break;
        case 'e':
          if (currentEvent) setEditingEvent(currentEvent);
          break;
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [queue, cursor, currentEvent, editingEvent, approve, reject]);

  const reviewed = progress ? progress.total_reviewed : 0;
  const total = progress ? progress.total_items : 0;
  const pct = total > 0 ? Math.round((reviewed / total) * 100) : 0;

  return (
    <div className="flex flex-col h-full bg-white overflow-hidden">
      {/* Progress bar */}
      <div className="px-5 py-3 border-b border-slate-200 shrink-0">
        <div className="flex items-center justify-between mb-1.5">
          <span className="text-sm font-medium text-slate-700">
            {reviewed} of {total} reviewed
          </span>
          <span className="text-sm font-semibold text-slate-900">{pct}%</span>
        </div>
        <div className="h-2 bg-slate-100 rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-blue-500 to-green-500 rounded-full transition-all duration-300"
            style={{ width: `${pct}%` }}
          />
        </div>
        {progress && (
          <div className="flex gap-3 mt-1.5 text-xs">
            <span className="text-green-600">✓ {progress.claims.approved} approved</span>
            <span className="text-red-500">✗ {progress.claims.rejected} rejected</span>
            <span className="text-blue-600">✎ {progress.claims.edited} edited</span>
            <span className="text-slate-400">○ {progress.claims.pending} pending</span>
          </div>
        )}
      </div>

      {/* Filter tabs */}
      <div className="flex border-b border-slate-200 shrink-0 px-4 pt-2 gap-1">
        {(['pending', 'all', 'approved', 'rejected', 'edited'] as ReviewFilter[]).map(f => (
          <button
            key={f}
            onClick={() => { setFilter(f); setCursor(0); }}
            className={`px-3 py-1.5 text-xs font-medium rounded-t transition-colors capitalize ${
              filter === f
                ? 'text-blue-700 border-b-2 border-blue-600 bg-blue-50'
                : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            {f}
            <span className="ml-1 text-slate-400">
              ({f === 'all'
                ? events.length
                : f === 'pending'
                ? events.filter(e => e.review_status === 'pending').length
                : f === 'approved'
                ? events.filter(e => e.review_status === 'approved').length
                : f === 'rejected'
                ? events.filter(e => e.review_status === 'rejected').length
                : events.filter(e => e.review_status === 'edited').length})
            </span>
          </button>
        ))}
      </div>

      {/* Keyboard hint */}
      <div className="px-5 py-2 bg-slate-50 border-b border-slate-100 shrink-0">
        <p className="text-xs text-slate-400">
          <kbd className="px-1 py-0.5 bg-white border border-slate-200 rounded text-slate-600">J/K</kbd> navigate
          &nbsp;·&nbsp;
          <kbd className="px-1 py-0.5 bg-white border border-slate-200 rounded text-green-600">A</kbd> approve
          &nbsp;·&nbsp;
          <kbd className="px-1 py-0.5 bg-white border border-slate-200 rounded text-red-500">R</kbd> reject
          &nbsp;·&nbsp;
          <kbd className="px-1 py-0.5 bg-white border border-slate-200 rounded text-blue-600">E</kbd> edit
        </p>
      </div>

      {/* Queue */}
      <div className="flex-1 overflow-y-auto">
        {queue.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center px-8">
            <div className="text-4xl mb-3">✓</div>
            <p className="text-slate-600 font-medium">
              {filter === 'pending' ? 'All events reviewed!' : `No ${filter} events`}
            </p>
          </div>
        ) : (
          queue.map((event, idx) => (
            <div
              key={event.id}
              onClick={() => setCursor(idx)}
              className={`border-b border-slate-100 cursor-pointer transition-colors ${
                idx === cursor
                  ? 'bg-blue-50 border-l-2 border-l-blue-500'
                  : 'hover:bg-slate-50'
              }`}
            >
              <div className="px-4 py-3">
                <div className="flex items-start gap-2 mb-1">
                  <span
                    className={`shrink-0 mt-0.5 text-xs font-semibold px-1.5 py-0.5 rounded ${
                      STATUS_COLORS[event.review_status] || STATUS_COLORS.pending
                    }`}
                  >
                    {STATUS_ICONS[event.review_status] || '○'}
                  </span>
                  <p className="text-sm font-medium text-slate-800 leading-snug">{event.title}</p>
                </div>
                <div className="flex items-center gap-3 ml-7 text-xs text-slate-400">
                  {event.event_type && <span>{event.event_type}</span>}
                  {event.date_text && <span>📅 {event.date_text}</span>}
                  <span className={`font-medium ${
                    event.confidence > 0.8 ? 'text-green-600' :
                    event.confidence > 0.6 ? 'text-amber-600' : 'text-red-500'
                  }`}>
                    {Math.round(event.confidence * 100)}%
                  </span>
                  {event.inconsistencies?.length > 0 && (
                    <span className="text-amber-500">⚡ {event.inconsistencies.length}</span>
                  )}
                </div>
              </div>

              {/* Expanded detail for focused item */}
              {idx === cursor && (
                <div className="px-4 pb-3 ml-7">
                  {event.description && (
                    <p className="text-xs text-slate-600 mb-2 leading-relaxed">{event.description}</p>
                  )}
                  <div className="flex gap-2">
                    <button
                      onClick={e => { e.stopPropagation(); approve(event); }}
                      disabled={saving || event.review_status === 'approved'}
                      className="px-3 py-1 text-xs font-semibold bg-green-100 text-green-700 rounded hover:bg-green-200 disabled:opacity-40 transition-colors"
                    >
                      ✓ Approve
                    </button>
                    <button
                      onClick={e => { e.stopPropagation(); reject(event); }}
                      disabled={saving || event.review_status === 'rejected'}
                      className="px-3 py-1 text-xs font-semibold bg-red-100 text-red-600 rounded hover:bg-red-200 disabled:opacity-40 transition-colors"
                    >
                      ✗ Reject
                    </button>
                    <button
                      onClick={e => { e.stopPropagation(); setEditingEvent(event); }}
                      className="px-3 py-1 text-xs font-semibold bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition-colors"
                    >
                      ✎ Edit
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {/* Edit modal */}
      {editingEvent && (
        <EditModal
          event={editingEvent}
          onSave={handleSaveEdit}
          onClose={() => setEditingEvent(null)}
        />
      )}
    </div>
  );
};

export default ReviewPanel;
