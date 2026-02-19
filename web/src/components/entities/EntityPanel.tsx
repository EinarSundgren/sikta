import React, { useState, useMemo } from 'react';
import { Entity, Relationship } from '../../types';

interface EntityPanelProps {
  entities: Entity[];
  relationships: Relationship[];
  selectedEntityId: string | null;
  onEntitySelect: (id: string | null) => void;
}

const TYPE_ICONS: Record<string, string> = {
  person: '👤',
  place: '📍',
  organization: '🏛️',
  object: '📦',
  amount: '💰',
};

const TYPE_LABELS: Record<string, string> = {
  person: 'People',
  place: 'Places',
  organization: 'Organizations',
  object: 'Objects',
  amount: 'Amounts',
};

const TYPE_ORDER = ['person', 'place', 'organization', 'object', 'amount'];

const EntityPanel: React.FC<EntityPanelProps> = ({
  entities,
  relationships,
  selectedEntityId,
  onEntitySelect,
}) => {
  const [search, setSearch] = useState('');

  const relCountById = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const rel of relationships) {
      counts[rel.entity_a_id] = (counts[rel.entity_a_id] || 0) + 1;
      counts[rel.entity_b_id] = (counts[rel.entity_b_id] || 0) + 1;
    }
    return counts;
  }, [relationships]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return entities;
    return entities.filter(e =>
      e.name.toLowerCase().includes(q) ||
      (e.aliases || []).some(a => a.toLowerCase().includes(q))
    );
  }, [entities, search]);

  const grouped = useMemo(() => {
    const groups: Record<string, Entity[]> = {};
    for (const e of filtered) {
      const type = e.entity_type || 'other';
      if (!groups[type]) groups[type] = [];
      groups[type].push(e);
    }
    for (const type of Object.keys(groups)) {
      groups[type].sort((a, b) => (relCountById[b.id] || 0) - (relCountById[a.id] || 0));
    }
    return groups;
  }, [filtered, relCountById]);

  const orderedTypes = [
    ...TYPE_ORDER.filter(t => grouped[t]),
    ...Object.keys(grouped).filter(t => !TYPE_ORDER.includes(t)),
  ];

  return (
    <div className="flex flex-col h-full bg-white border-r border-slate-200 overflow-hidden" style={{ width: 272 }}>
      <div className="px-3 py-3 border-b border-slate-200 shrink-0">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold text-slate-700">Entities</h2>
          <span className="text-xs text-slate-400">{entities.length} total</span>
        </div>
        <input
          type="text"
          placeholder="Search entities..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          className="w-full px-2.5 py-1.5 text-sm border border-slate-200 rounded focus:outline-none focus:ring-1 focus:ring-blue-400 bg-slate-50"
        />
        {selectedEntityId && (
          <button
            onClick={() => onEntitySelect(null)}
            className="mt-1.5 text-xs text-blue-600 hover:text-blue-800 flex items-center gap-1"
          >
            <span>✕</span> Clear filter
          </button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto">
        {orderedTypes.map(type => {
          const group = grouped[type];
          if (!group?.length) return null;
          const icon = TYPE_ICONS[type] || '•';
          const label = TYPE_LABELS[type] || type;
          return (
            <div key={type}>
              <div className="px-3 py-1.5 bg-slate-50 border-b border-slate-100 sticky top-0 z-10">
                <span className="text-xs font-semibold text-slate-400 uppercase tracking-wide">
                  {icon} {label} ({group.length})
                </span>
              </div>
              {group.map(entity => {
                const relCount = relCountById[entity.id] || 0;
                const isSelected = entity.id === selectedEntityId;
                return (
                  <button
                    key={entity.id}
                    onClick={() => onEntitySelect(isSelected ? null : entity.id)}
                    className={`w-full px-3 py-2 text-left border-b border-slate-100 transition-colors ${
                      isSelected
                        ? 'bg-blue-50 border-l-2 border-l-blue-500 pl-[10px]'
                        : 'hover:bg-slate-50'
                    }`}
                  >
                    <div className="flex items-center justify-between gap-1">
                      <span
                        className={`text-sm truncate ${
                          isSelected ? 'text-blue-700 font-medium' : 'text-slate-800'
                        }`}
                      >
                        {entity.name}
                      </span>
                      {relCount > 0 && (
                        <span className="text-xs text-slate-400 shrink-0">{relCount}</span>
                      )}
                    </div>
                  </button>
                );
              })}
            </div>
          );
        })}
        {filtered.length === 0 && (
          <div className="px-3 py-6 text-center text-sm text-slate-400">
            No entities match "{search}"
          </div>
        )}
      </div>
    </div>
  );
};

export default EntityPanel;
