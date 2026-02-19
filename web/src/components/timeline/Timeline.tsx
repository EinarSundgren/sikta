import React, { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import { TimelineEvent } from '../../types';

interface TimelineProps {
  events: TimelineEvent[];
  onEventClick?: (event: TimelineEvent) => void;
  highlightedIds?: string[];
  width?: number;
  height?: number;
}

const Timeline: React.FC<TimelineProps> = ({
  events,
  onEventClick,
  highlightedIds = [],
  width = 1200,
  height = 400
}) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const [hoveredEvent, setHoveredEvent] = useState<string | null>(null);

  useEffect(() => {
    if (!svgRef.current || events.length === 0) return;

    // Clear previous content
    d3.select(svgRef.current).selectAll('*').remove();

    const svg = d3.select(svgRef.current);
    const margin = { top: 60, right: 150, bottom: 60, left: 150 };
    const innerWidth = width - margin.left - margin.right;
    const innerHeight = height - margin.top - margin.bottom;

    const g = svg
      .append('g')
      .attr('transform', `translate(${margin.left},${margin.top})`);

    // Split events into two lanes
    const eventsWithChrono = events.filter(e => e.chronological_position !== null);
    const eventsWithNarrative = events.filter(e => e.narrative_position !== null);

    // Scales
    const maxChronoPos = Math.max(...eventsWithChrono.map(e => e.chronological_position || 0));
    const maxNarrativePos = Math.max(...eventsWithNarrative.map(e => e.narrative_position));

    const chronoScale = d3.scaleLinear()
      .domain([0, maxChronoPos + 1])
      .range([0, innerWidth]);

    const narrativeScale = d3.scaleLinear()
      .domain([0, maxNarrativePos + 1])
      .range([0, innerWidth]);

    // Lane backgrounds
    g.append('rect')
      .attr('x', 0)
      .attr('y', 0)
      .attr('width', innerWidth)
      .attr('height', innerHeight / 2 - 10)
      .attr('fill', '#f8fafc')
      .attr('rx', 4);

    g.append('rect')
      .attr('x', 0)
      .attr('y', innerHeight / 2 + 10)
      .attr('width', innerWidth)
      .attr('height', innerHeight / 2 - 10)
      .attr('fill', '#f1f5f9')
      .attr('rx', 4);

    // Lane labels
    g.append('text')
      .attr('x', -10)
      .attr('y', innerHeight / 4)
      .attr('text-anchor', 'end')
      .attr('alignment-baseline', 'middle')
      .attr('font-size', '14px')
      .attr('font-weight', '600')
      .attr('fill', '#475569')
      .text('Chronological');

    g.append('text')
      .attr('x', -10)
      .attr('y', 3 * innerHeight / 4)
      .attr('text-anchor', 'end')
      .attr('alignment-baseline', 'middle')
      .attr('font-size', '14px')
      .attr('font-weight', '600')
      .attr('fill', '#475569')
      .text('Narrative');

    // Draw events in chronological lane
    eventsWithChrono.forEach(event => {
      const x = chronoScale(event.chronological_position || 0);
      const y = innerHeight / 4;

      const isHighlighted = highlightedIds.includes(event.id);

      // Event circle
      g.append('circle')
        .attr('cx', x)
        .attr('cy', y)
        .attr('r', isHighlighted ? 11 : 8)
        .attr('fill', eventHasInconsistency(event) ? '#ef4444' : '#3b82f6')
        .attr('stroke', isHighlighted ? '#f59e0b' : '#fff')
        .attr('stroke-width', isHighlighted ? 3 : 2)
        .attr('cursor', 'pointer')
        .attr('data-event-id', event.id)
        .on('click', () => onEventClick?.(event))
        .on('mouseenter', () => setHoveredEvent(event.id))
        .on('mouseleave', () => setHoveredEvent(null));

      // Inconsistency marker
      if (eventHasInconsistency(event)) {
        g.append('text')
          .attr('x', x)
          .attr('y', y - 15)
          .attr('text-anchor', 'middle')
          .attr('font-size', '16px')
          .text('⚡');
      }

      // Confidence indicator (dot size based on confidence)
      const confidenceSize = 3 + event.confidence * 5;
      g.append('circle')
        .attr('cx', x)
        .attr('cy', y)
        .attr('r', confidenceSize)
        .attr('fill', 'none')
        .attr('stroke', event.confidence > 0.8 ? '#22c55e' : event.confidence > 0.6 ? '#eab308' : '#f97316')
        .attr('stroke-width', 2)
        .attr('opacity', 0.6);
    });

    // Draw events in narrative lane
    eventsWithNarrative.forEach(event => {
      const x = narrativeScale(event.narrative_position);
      const y = 3 * innerHeight / 4;

      // Event circle
      g.append('circle')
        .attr('cx', x)
        .attr('cy', y)
        .attr('r', 8)
        .attr('fill', eventHasInconsistency(event) ? '#ef4444' : '#8b5cf6')
        .attr('stroke', '#fff')
        .attr('stroke-width', 2)
        .attr('cursor', 'pointer')
        .attr('data-event-id', event.id)
        .on('click', () => onEventClick?.(event))
        .on('mouseenter', () => setHoveredEvent(event.id))
        .on('mouseleave', () => setHoveredEvent(null));

      // Inconsistency marker
      if (eventHasInconsistency(event)) {
        g.append('text')
          .attr('x', x)
          .attr('y', y - 15)
          .attr('text-anchor', 'middle')
          .attr('font-size', '16px')
          .text('⚡');
      }
    });

    // Draw connector lines between same event in both lanes
    events.forEach(event => {
      if (event.chronological_position !== null && event.narrative_position !== null) {
        const x1 = chronoScale(event.chronological_position);
        const y1 = innerHeight / 4;
        const x2 = narrativeScale(event.narrative_position);
        const y2 = 3 * innerHeight / 4;

        g.append('line')
          .attr('x1', x1)
          .attr('y1', y1 + 10)
          .attr('x2', x2)
          .attr('y2', y2 - 10)
          .attr('stroke', '#cbd5e1')
          .attr('stroke-width', 1)
          .attr('stroke-dasharray', '4,4')
          .attr('opacity', 0.6);
      }
    });

  }, [events, width, height, onEventClick, highlightedIds]);

  const eventHasInconsistency = (event: TimelineEvent): boolean => {
    return event.inconsistencies && event.inconsistencies.length > 0;
  };

  const hoveredEventData = events.find(e => e.id === hoveredEvent);

  return (
    <div className="relative">
      <svg
        ref={svgRef}
        width={width}
        height={height}
        className="border border-slate-200 rounded-lg bg-white"
      />
      {hoveredEventData && (
        <div className="absolute top-4 right-4 bg-white border border-slate-200 rounded-lg shadow-lg p-4 max-w-sm z-10">
          <h3 className="font-semibold text-lg text-slate-800 mb-2">
            {hoveredEventData.title}
          </h3>
          <p className="text-sm text-slate-600 mb-2">
            {hoveredEventData.description}
          </p>
          <div className="flex items-center gap-2 text-xs text-slate-500">
            <span className="px-2 py-1 bg-blue-100 text-blue-700 rounded">
              {hoveredEventData.event_type}
            </span>
            <span>Confidence: {Math.round(hoveredEventData.confidence * 100)}%</span>
          </div>
          {hoveredEventData.inconsistencies && hoveredEventData.inconsistencies.length > 0 && (
            <div className="mt-2 pt-2 border-t border-slate-200">
              <span className="text-xs font-medium text-red-600">
                ⚡ {hoveredEventData.inconsistencies.length} inconsistency(es)
              </span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default Timeline;
