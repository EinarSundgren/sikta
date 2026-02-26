import { useEffect, useRef, useState, useMemo } from 'react';
import * as d3 from 'd3';
import { Node, Edge, Document } from '../../api/projects';

// Design tokens
const tokens = {
  surface0: '#FAFBFC',
  surface1: '#FFFFFF',
  surface2: '#F3F4F6',
  textPrimary: '#1A1D23',
  textSecondary: '#4B5162',
  textTertiary: '#8B90A0',
  borderDefault: '#E2E4E9',
  borderSubtle: '#ECEEF1',
  accentPrimary: '#3B6FED',
};

const DOC_COLORS = [
  '#3B6FED', // Blue
  '#0D9488', // Teal
  '#D97706', // Amber
  '#7C3AED', // Purple
  '#DB2777', // Rose
];

const ENTITY_TYPE_COLORS: Record<string, string> = {
  person: '#3B6FED',
  organization: '#D97706',
  place: '#0D9488',
  object: '#8B5CF6',
  amount: '#DC2626',
  event: '#64748b',
};

const ENTITY_TYPE_BG: Record<string, string> = {
  person: '#EEF2FF',
  organization: '#FFFBEB',
  place: '#F0FDFA',
  object: '#F5F3FF',
  amount: '#FEF2F2',
  event: '#F8FAFC',
};

const fontDisplay = "'DM Sans', sans-serif";
const fontMono = "'JetBrains Mono', monospace";

interface SwimlaneEvent {
  id: string;
  label: string;
  description?: string;
  date?: string;
  confidence?: number;
  documentId?: string;
  documentTitle?: string;
  entityIds: string[];
  properties?: Record<string, unknown>;
}

interface SwimlaneEntity {
  id: string;
  label: string;
  node_type: string;
  documentIds: string[];
}

interface SwimlaneTimelineProps {
  nodes: Node[];
  edges: Edge[];
  documents: Document[];
  onEventClick?: (event: SwimlaneEvent) => void;
  width?: number;
  height?: number;
}

// Extract entities and events from nodes
function processData(nodes: Node[], documents: Document[]): {
  entities: SwimlaneEntity[];
  events: SwimlaneEvent[];
  docColorMap: Map<string, string>;
} {
  const entities: SwimlaneEntity[] = [];
  const events: SwimlaneEvent[] = [];
  const docColorMap = new Map<string, string>();

  // Create document color map
  documents.forEach((doc, i) => {
    docColorMap.set(doc.id, DOC_COLORS[i % DOC_COLORS.length]);
  });

  // Separate entities and events
  nodes.forEach(node => {
    if (node.node_type === 'event') {
      // Extract entity IDs from properties if available
      const props = node.properties || {};
      const entityIds: string[] = [];

      // Check for involved_entities in properties
      if (Array.isArray(props.involved_entities)) {
        entityIds.push(...props.involved_entities);
      }
      if (typeof props.entity_id === 'string') {
        entityIds.push(props.entity_id);
      }

      events.push({
        id: node.id,
        label: node.label,
        description: props.description as string | undefined,
        date: props.date as string | undefined,
        confidence: props.confidence as number | undefined,
        documentId: props.document_id as string | undefined,
        documentTitle: props.document_title as string | undefined,
        entityIds,
        properties: props,
      });
    } else {
      // Entity node
      const props = node.properties || {};
      const docIds: string[] = [];
      if (typeof props.document_id === 'string') {
        docIds.push(props.document_id);
      }

      entities.push({
        id: node.id,
        label: node.label,
        node_type: node.node_type,
        documentIds: docIds,
      });
    }
  });

  return { entities, events, docColorMap };
}

const SwimlaneTimeline: React.FC<SwimlaneTimelineProps> = ({
  nodes,
  edges,
  documents,
  onEventClick,
  width = 1100,
  height = 600,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);
  const [hoveredEvent, setHoveredEvent] = useState<string | null>(null);

  const { entities, events, docColorMap } = useMemo(
    () => processData(nodes, documents),
    [nodes, documents]
  );

  // Build event-to-entities map from edges
  const eventEntityMap = useMemo(() => {
    const map = new Map<string, string[]>();
    edges.forEach(edge => {
      // Handle various edge types that connect entities to events
      if (edge.edge_type === 'involves' || edge.edge_type === 'has_participant' ||
          edge.edge_type === 'involved_in' || edge.edge_type === 'related_to') {
        // For involved_in: source is entity, target is event
        // For involves: source is event, target is entity
        if (edge.edge_type === 'involved_in' || edge.edge_type === 'related_to') {
          // Source is entity, target is event
          const existing = map.get(edge.target_node) || [];
          existing.push(edge.source_node);
          map.set(edge.target_node, existing);
        } else {
          // Source is event, target is entity
          const existing = map.get(edge.source_node) || [];
          existing.push(edge.target_node);
          map.set(edge.source_node, existing);
        }
      }
    });
    return map;
  }, [edges]);

  // Update events with entity info from edges
  const enrichedEvents = useMemo(() => {
    return events.map(event => ({
      ...event,
      entityIds: event.entityIds.length > 0
        ? event.entityIds
        : (eventEntityMap.get(event.id) || []),
    }));
  }, [events, eventEntityMap]);

  // Filter to show only entities with events, or create a default lane
  const activeEntities = useMemo(() => {
    const entityIdsWithEvents = new Set<string>();
    enrichedEvents.forEach(event => {
      event.entityIds.forEach(id => entityIdsWithEvents.add(id));
    });
    const filtered = entities.filter(e => entityIdsWithEvents.has(e.id));

    // If no entities are connected to events but we have events,
    // create a virtual "All Events" entity to show them
    if (filtered.length === 0 && events.length > 0) {
      return [{
        id: '__all_events__',
        label: 'All Events',
        node_type: 'event',
        documentIds: [],
      }];
    }
    return filtered;
  }, [entities, enrichedEvents, events.length]);

  // For the default lane, assign all events to it
  const finalEvents = useMemo(() => {
    if (activeEntities.length === 1 && activeEntities[0].id === '__all_events__') {
      return enrichedEvents.map(event => ({
        ...event,
        entityIds: ['__all_events__'],
      }));
    }
    return enrichedEvents;
  }, [activeEntities, enrichedEvents]);

  // Sort events by date
  const sortedEvents = useMemo(() => {
    return [...finalEvents].sort((a, b) => {
      if (a.date && b.date) return a.date.localeCompare(b.date);
      if (a.date) return -1;
      if (b.date) return 1;
      return 0;
    });
  }, [finalEvents]);

  useEffect(() => {
    if (!svgRef.current || sortedEvents.length === 0) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    const margin = { top: 40, right: 40, bottom: 40, left: 200 };
    const laneHeight = 80;
    const laneGap = 4;
    const eventWidth = 160;
    const eventGap = 12;

    // Calculate dimensions
    const totalWidth = Math.max(
      width,
      sortedEvents.length * (eventWidth + eventGap) + margin.left + margin.right
    );
    const totalHeight = activeEntities.length * (laneHeight + laneGap) + margin.top + margin.bottom;

    svg.attr('width', totalWidth).attr('height', Math.max(height, totalHeight));

    const g = svg.append('g').attr('transform', `translate(${margin.left},${margin.top})`);

    // Create entity position map
    const entityY = new Map<string, number>();
    activeEntities.forEach((entity, i) => {
      entityY.set(entity.id, i * (laneHeight + laneGap));
    });

    // Create event position map
    const eventX = new Map<string, number>();
    sortedEvents.forEach((event, i) => {
      eventX.set(event.id, i * (eventWidth + eventGap));
    });

    // Draw lane backgrounds and headers
    activeEntities.forEach((entity, i) => {
      const y = i * (laneHeight + laneGap);
      const bgColor = ENTITY_TYPE_BG[entity.node_type] || tokens.surface2;
      const typeColor = ENTITY_TYPE_COLORS[entity.node_type] || tokens.textTertiary;

      // Lane background
      g.append('rect')
        .attr('x', 0)
        .attr('y', y)
        .attr('width', totalWidth - margin.left - margin.right)
        .attr('height', laneHeight)
        .attr('rx', 6)
        .attr('fill', bgColor)
        .attr('opacity', 0.5);

      // Lane header (left side)
      const headerG = svg.append('g')
        .attr('transform', `translate(12, ${margin.top + y + laneHeight / 2})`);

      // Entity type icon
      headerG.append('circle')
        .attr('cx', 0)
        .attr('cy', 0)
        .attr('r', 6)
        .attr('fill', typeColor);

      // Entity name
      const displayName = entity.label.length > 24
        ? entity.label.slice(0, 22) + '...'
        : entity.label;
      headerG.append('text')
        .attr('x', 14)
        .attr('y', 4)
        .attr('font-family', fontDisplay)
        .attr('font-size', 12)
        .attr('font-weight', 600)
        .attr('fill', tokens.textPrimary)
        .text(displayName);

      // Entity type label
      headerG.append('text')
        .attr('x', 14)
        .attr('y', 18)
        .attr('font-family', fontMono)
        .attr('font-size', 9)
        .attr('fill', tokens.textTertiary)
        .text(entity.node_type);
    });

    // Draw time axis at top
    const axisG = svg.append('g')
      .attr('transform', `translate(${margin.left}, ${margin.top - 20})`);

    // Document markers on time axis
    const docPositions = new Map<string, number[]>();
    sortedEvents.forEach((event, i) => {
      if (event.documentId) {
        const positions = docPositions.get(event.documentId) || [];
        positions.push(i * (eventWidth + eventGap) + eventWidth / 2);
        docPositions.set(event.documentId, positions);
      }
    });

    // Draw document labels
    let docLabelX = 0;
    documents.forEach((doc, idx) => {
      const positions = docPositions.get(doc.id);
      if (!positions || positions.length === 0) return;

      const avgX = positions.reduce((a, b) => a + b, 0) / positions.length;
      const color = docColorMap.get(doc.id) || DOC_COLORS[idx % DOC_COLORS.length];

      // Only draw if it doesn't overlap with previous
      if (avgX > docLabelX + 100) {
        axisG.append('rect')
          .attr('x', avgX - 40)
          .attr('y', -12)
          .attr('width', 80)
          .attr('height', 20)
          .attr('rx', 4)
          .attr('fill', color + '20');

        axisG.append('text')
          .attr('x', avgX)
          .attr('y', 2)
          .attr('text-anchor', 'middle')
          .attr('font-family', fontMono)
          .attr('font-size', 10)
          .attr('font-weight', 500)
          .attr('fill', color)
          .text(doc.title.substring(0, 12));

        docLabelX = avgX;
      }
    });

    // Draw events
    const eventsG = g.append('g').attr('class', 'events');

    sortedEvents.forEach((event, eventIdx) => {
      const x = eventIdx * (eventWidth + eventGap);

      // Find which lanes this event belongs to
      event.entityIds.forEach(entityId => {
        const y = entityY.get(entityId);
        if (y === undefined) return;

        const isHovered = hoveredEvent === event.id;
        const docColor = event.documentId
          ? (docColorMap.get(event.documentId) || tokens.accentPrimary)
          : tokens.accentPrimary;

        const cardG = eventsG.append('g')
          .attr('transform', `translate(${x}, ${y + 8})`)
          .attr('cursor', 'pointer')
          .on('click', () => onEventClick?.(event))
          .on('mouseenter', () => setHoveredEvent(event.id))
          .on('mouseleave', () => setHoveredEvent(null));

        // Card background
        const cardHeight = laneHeight - 16;
        cardG.append('rect')
          .attr('width', eventWidth)
          .attr('height', cardHeight)
          .attr('rx', 6)
          .attr('fill', tokens.surface1)
          .attr('stroke', isHovered ? docColor : tokens.borderDefault)
          .attr('stroke-width', isHovered ? 2 : 1)
          .style('filter', isHovered
            ? 'drop-shadow(0 4px 12px rgba(0,0,0,0.15))'
            : 'drop-shadow(0 1px 2px rgba(0,0,0,0.04))');

        // Document color bar on left
        cardG.append('rect')
          .attr('width', 3)
          .attr('height', cardHeight)
          .attr('rx', 1.5)
          .attr('fill', docColor);

        // Event title
        const titleText = event.label.length > 28
          ? event.label.slice(0, 26) + '...'
          : event.label;
        cardG.append('text')
          .attr('x', 10)
          .attr('y', 18)
          .attr('font-family', fontDisplay)
          .attr('font-size', 11)
          .attr('font-weight', 600)
          .attr('fill', tokens.textPrimary)
          .text(titleText);

        // Date if available
        if (event.date) {
          cardG.append('text')
            .attr('x', 10)
            .attr('y', 32)
            .attr('font-family', fontMono)
            .attr('font-size', 9)
            .attr('fill', tokens.textTertiary)
            .text(event.date.substring(0, 16));
        }

        // Document badge
        if (event.documentTitle) {
          const badgeText = event.documentTitle.length > 14
            ? event.documentTitle.slice(0, 12) + '..'
            : event.documentTitle;
          cardG.append('rect')
            .attr('x', 10)
            .attr('y', cardHeight - 22)
            .attr('width', badgeText.length * 6 + 8)
            .attr('height', 14)
            .attr('rx', 3)
            .attr('fill', docColor + '20');
          cardG.append('text')
            .attr('x', 14)
            .attr('y', cardHeight - 12)
            .attr('font-family', fontMono)
            .attr('font-size', 8)
            .attr('font-weight', 500)
            .attr('fill', docColor)
            .text(badgeText);
        }

        // Entrance animation
        cardG.attr('opacity', 0)
          .transition()
          .delay(100 + eventIdx * 30)
          .duration(300)
          .attr('opacity', 1);
      });
    });

  }, [activeEntities, sortedEvents, documents, docColorMap, hoveredEvent, width, height, onEventClick]);

  if (events.length === 0) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: 400,
          backgroundColor: tokens.surface0,
          borderRadius: 12,
          border: `1px solid ${tokens.borderSubtle}`,
        }}
      >
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: 32, marginBottom: 12 }}>📊</div>
          <p style={{ color: tokens.textSecondary, fontSize: 14 }}>
            No events extracted yet.
          </p>
          <p style={{ color: tokens.textTertiary, fontSize: 12, marginTop: 4 }}>
            Add documents and run extraction to see the swimlane timeline.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      style={{
        overflow: 'auto',
        backgroundColor: tokens.surface0,
        borderRadius: 12,
        border: `1px solid ${tokens.borderSubtle}`,
      }}
    >
      <svg ref={svgRef} style={{ display: 'block', minWidth: width }} />

      {/* Legend */}
      <div
        style={{
          position: 'absolute',
          bottom: 12,
          right: 12,
          display: 'flex',
          flexWrap: 'wrap',
          gap: 8,
          maxWidth: 300,
        }}
      >
        {documents.slice(0, 5).map((doc, i) => (
          <div
            key={doc.id}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 4,
              backgroundColor: 'rgba(255,255,255,0.9)',
              padding: '2px 8px',
              borderRadius: 4,
              fontSize: 10,
            }}
          >
            <div
              style={{
                width: 8,
                height: 8,
                borderRadius: 2,
                backgroundColor: docColorMap.get(doc.id) || DOC_COLORS[i],
              }}
            />
            <span style={{ color: tokens.textSecondary }}>{doc.title.substring(0, 15)}</span>
          </div>
        ))}
      </div>

      {/* Hint */}
      <div
        style={{
          position: 'absolute',
          top: 12,
          right: 12,
          fontSize: 11,
          color: tokens.textTertiary,
          backgroundColor: 'rgba(255,255,255,0.8)',
          padding: '4px 8px',
          borderRadius: 4,
        }}
      >
        Scroll to explore · Click event for details
      </div>
    </div>
  );
};

export default SwimlaneTimeline;
