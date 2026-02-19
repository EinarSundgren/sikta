import React, { useState } from 'react';
import { Inconsistency } from '../../types';
import { timelineApi } from '../../api/timeline';

interface InconsistencyPanelProps {
  inconsistencies: Inconsistency[];
  onInconsistencyUpdated: (id: string, status: string, note: string) => void;
  onHighlightClaims?: (claimIds: string[]) => void;
}

const SEVERITY_STYLES: Record<string, { badge: string; icon: string; border: string }> = {
  conflict: { badge: 'bg-red-100 text-red-700', icon: '⚡', border: 'border-l-red-500' },
  warning:  { badge: 'bg-amber-100 text-amber-700', icon: '⚠️', border: 'border-l-amber-400' },
  info:     { badge: 'bg-slate-100 text-slate-600', icon: 'ℹ️', border: 'border-l-slate-300' },
};

const RESOLUTION_STYLES: Record<string, string> = {
  unresolved: 'text-slate-500',
  resolved:   'text-green-600',
  noted:      'text-blue-600',
  dismissed:  'text-slate-400 line-through',
};

const TYPE_LABELS: Record<string, string> = {
  narrative_chronological_mismatch: 'Narrative/Chronological mismatch',
  temporal_impossibility: 'Temporal impossibility',
  contradiction: 'Contradiction',
  cross_reference: 'Cross-reference',
  duplicate_entity: 'Duplicate entity',
  data_mismatch: 'Data mismatch',
};

interface NoteFormProps {
  id: string;
  currentNote: string | null;
  onSave: (note: string) => void;
  onCancel: () => void;
}

const NoteForm: React.FC<NoteFormProps> = ({ currentNote, onSave, onCancel }) => {
  const [note, setNote] = useState(currentNote ?? '');
  return (
    <div className="mt-2">
      <textarea
        value={note}
        onChange={e => setNote(e.target.value)}
        placeholder="Add a note explaining this inconsistency..."
        rows={2}
        autoFocus
        className="w-full px-2.5 py-1.5 text-xs border border-slate-200 rounded focus:outline-none focus:ring-1 focus:ring-blue-400 resize-none"
      />
      <div className="flex gap-2 mt-1.5">
        <button
          onClick={() => onSave(note)}
          className="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Save note
        </button>
        <button
          onClick={onCancel}
          className="px-3 py-1 text-xs text-slate-500 hover:text-slate-700"
        >
          Cancel
        </button>
      </div>
    </div>
  );
};

const InconsistencyPanel: React.FC<InconsistencyPanelProps> = ({
  inconsistencies,
  onInconsistencyUpdated,
  onHighlightClaims,
}) => {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [noteFormId, setNoteFormId] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [severityFilter, setSeverityFilter] = useState<string>('all');

  const resolve = async (inc: Inconsistency, status: string, note?: string) => {
    if (saving) return;
    setSaving(true);
    try {
      const noteText = note ?? inc.resolution_note ?? '';
      await timelineApi.resolveInconsistency(inc.id, status, noteText);
      onInconsistencyUpdated(inc.id, status, noteText);
    } finally {
      setSaving(false);
      setNoteFormId(null);
    }
  };

  const filtered = severityFilter === 'all'
    ? inconsistencies
    : inconsistencies.filter(i => i.severity === severityFilter);

  const counts = {
    conflict: inconsistencies.filter(i => i.severity === 'conflict').length,
    warning: inconsistencies.filter(i => i.severity === 'warning').length,
    info: inconsistencies.filter(i => i.severity === 'info').length,
  };

  return (
    <div className="flex flex-col h-full bg-white overflow-hidden">
      {/* Header with severity filter */}
      <div className="px-5 py-3 border-b border-slate-200 shrink-0">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-sm font-semibold text-slate-700">
            {inconsistencies.length} inconsistenc{inconsistencies.length === 1 ? 'y' : 'ies'}
          </h2>
        </div>
        <div className="flex gap-2">
          {([
            { key: 'all', label: `All (${inconsistencies.length})` },
            { key: 'conflict', label: `⚡ Conflicts (${counts.conflict})` },
            { key: 'warning', label: `⚠️ Warnings (${counts.warning})` },
            { key: 'info', label: `ℹ️ Info (${counts.info})` },
          ] as { key: string; label: string }[]).map(f => (
            <button
              key={f.key}
              onClick={() => setSeverityFilter(f.key)}
              className={`px-2.5 py-1 text-xs rounded transition-colors ${
                severityFilter === f.key
                  ? 'bg-slate-900 text-white'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {/* List */}
      <div className="flex-1 overflow-y-auto">
        {filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center px-8">
            <div className="text-4xl mb-3">✓</div>
            <p className="text-slate-500 text-sm">No inconsistencies to show</p>
          </div>
        ) : (
          filtered.map(inc => {
            const styles = SEVERITY_STYLES[inc.severity] ?? SEVERITY_STYLES.info;
            const isExpanded = expandedId === inc.id;
            const isDismissed = inc.resolution_status === 'dismissed';

            return (
              <div
                key={inc.id}
                className={`border-b border-slate-100 border-l-2 ${styles.border} ${isDismissed ? 'opacity-50' : ''}`}
              >
                <button
                  className="w-full text-left px-4 py-3 hover:bg-slate-50 transition-colors"
                  onClick={() => setExpandedId(isExpanded ? null : inc.id)}
                >
                  <div className="flex items-start gap-2">
                    <span className="shrink-0 text-sm mt-0.5">{styles.icon}</span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <p className={`text-sm font-medium text-slate-800 ${isDismissed ? 'line-through' : ''}`}>
                          {inc.title}
                        </p>
                        <span className={`text-xs px-1.5 py-0.5 rounded ${styles.badge}`}>
                          {inc.severity}
                        </span>
                        {inc.resolution_status !== 'unresolved' && (
                          <span className={`text-xs ${RESOLUTION_STYLES[inc.resolution_status]}`}>
                            {inc.resolution_status}
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-slate-400 mt-0.5">
                        {TYPE_LABELS[inc.inconsistency_type] || inc.inconsistency_type}
                      </p>
                    </div>
                    <span className="text-slate-300 text-xs shrink-0">{isExpanded ? '▲' : '▼'}</span>
                  </div>
                </button>

                {isExpanded && (
                  <div className="px-4 pb-3 ml-6">
                    <p className="text-xs text-slate-600 mb-3 leading-relaxed">{inc.description}</p>

                    {inc.resolution_note && (
                      <div className="mb-3 px-2.5 py-2 bg-blue-50 rounded text-xs text-blue-700">
                        <span className="font-semibold">Note: </span>{inc.resolution_note}
                      </div>
                    )}

                    {/* Action buttons */}
                    {!isDismissed && (
                      <div className="flex flex-wrap gap-2 mb-2">
                        {inc.resolution_status !== 'resolved' && (
                          <button
                            onClick={() => resolve(inc, 'resolved')}
                            disabled={saving}
                            className="px-2.5 py-1 text-xs bg-green-100 text-green-700 rounded hover:bg-green-200 disabled:opacity-40 transition-colors"
                          >
                            ✓ Mark resolved
                          </button>
                        )}
                        <button
                          onClick={() => setNoteFormId(noteFormId === inc.id ? null : inc.id)}
                          className="px-2.5 py-1 text-xs bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition-colors"
                        >
                          ✎ {inc.resolution_note ? 'Edit note' : 'Add note'}
                        </button>
                        {inc.resolution_status !== 'dismissed' && (
                          <button
                            onClick={() => resolve(inc, 'dismissed')}
                            disabled={saving}
                            className="px-2.5 py-1 text-xs bg-slate-100 text-slate-500 rounded hover:bg-slate-200 disabled:opacity-40 transition-colors"
                          >
                            Dismiss
                          </button>
                        )}
                        {onHighlightClaims && (() => {
                          const claimIds: string[] = [];
                          if (inc.metadata?.['claim_ids']) {
                            claimIds.push(...(inc.metadata['claim_ids'] as string[]));
                          } else if (inc.metadata?.['claim_id']) {
                            claimIds.push(inc.metadata['claim_id'] as string);
                          }
                          return claimIds.length > 0 ? (
                            <button
                              onClick={() => onHighlightClaims(claimIds)}
                              className="px-2.5 py-1 text-xs bg-purple-100 text-purple-700 rounded hover:bg-purple-200 transition-colors"
                            >
                              → Show on timeline
                            </button>
                          ) : null;
                        })()}
                      </div>
                    )}

                    {noteFormId === inc.id && (
                      <NoteForm
                        id={inc.id}
                        currentNote={inc.resolution_note}
                        onSave={note => resolve(inc, 'noted', note)}
                        onCancel={() => setNoteFormId(null)}
                      />
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};

export default InconsistencyPanel;
