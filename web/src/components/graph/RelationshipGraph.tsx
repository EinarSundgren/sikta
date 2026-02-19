import React, { useEffect, useRef, useMemo } from 'react';
import * as d3 from 'd3';
import { Entity, Relationship } from '../../types';

interface GraphNode extends d3.SimulationNodeDatum {
  id: string;
  name: string;
  entity_type: string;
  relCount: number;
}

interface GraphLink extends d3.SimulationLinkDatum<GraphNode> {
  id: string;
  relationship_type: string;
  description: string | null;
}

interface RelationshipGraphProps {
  entities: Entity[];
  relationships: Relationship[];
  selectedEntityId: string | null;
  onEntitySelect: (id: string | null) => void;
}

const TYPE_COLORS: Record<string, string> = {
  person: '#3b82f6',
  place: '#10b981',
  organization: '#f59e0b',
  object: '#8b5cf6',
  amount: '#ef4444',
};

function nodeRadius(relCount: number): number {
  return Math.max(5, Math.sqrt(Math.max(relCount, 1)) * 5);
}

const RelationshipGraph: React.FC<RelationshipGraphProps> = ({
  entities,
  relationships,
  selectedEntityId,
  onEntitySelect,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);
  const simulationRef = useRef<d3.Simulation<GraphNode, GraphLink> | null>(null);

  const relCountById = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const rel of relationships) {
      counts[rel.entity_a_id] = (counts[rel.entity_a_id] || 0) + 1;
      counts[rel.entity_b_id] = (counts[rel.entity_b_id] || 0) + 1;
    }
    return counts;
  }, [relationships]);

  useEffect(() => {
    if (!svgRef.current || !containerRef.current || entities.length === 0) return;

    const container = containerRef.current;
    const width = container.clientWidth || 900;
    const height = container.clientHeight || 600;

    // Stop previous simulation
    simulationRef.current?.stop();
    d3.select(svgRef.current).selectAll('*').remove();

    const entitySet = new Set(entities.map(e => e.id));

    const nodes: GraphNode[] = entities.map(e => ({
      id: e.id,
      name: e.name,
      entity_type: e.entity_type,
      relCount: relCountById[e.id] || 0,
    }));

    const links: GraphLink[] = relationships
      .filter(r => entitySet.has(r.entity_a_id) && entitySet.has(r.entity_b_id))
      .map(r => ({
        id: r.id,
        source: r.entity_a_id,
        target: r.entity_b_id,
        relationship_type: r.relationship_type,
        description: r.description,
      }));

    const svg = d3.select(svgRef.current)
      .attr('width', width)
      .attr('height', height);

    const g = svg.append('g');

    // Zoom behavior
    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.15, 5])
      .on('zoom', event => g.attr('transform', event.transform));
    svg.call(zoom);

    // Initial zoom-to-fit after simulation settles
    const fitView = () => {
      const bounds = (g.node() as SVGGElement).getBBox();
      if (bounds.width === 0 || bounds.height === 0) return;
      const padding = 40;
      const scale = Math.min(
        (width - padding * 2) / bounds.width,
        (height - padding * 2) / bounds.height,
        1.2
      );
      const tx = width / 2 - scale * (bounds.x + bounds.width / 2);
      const ty = height / 2 - scale * (bounds.y + bounds.height / 2);
      svg.transition().duration(600).call(
        zoom.transform,
        d3.zoomIdentity.translate(tx, ty).scale(scale)
      );
    };

    // Draw links
    const link = g.append('g')
      .selectAll<SVGLineElement, GraphLink>('line')
      .data(links)
      .join('line')
      .attr('stroke', '#e2e8f0')
      .attr('stroke-width', 1.5)
      .attr('opacity', 0.8);

    // Draw nodes
    const node = g.append('g')
      .selectAll<SVGCircleElement, GraphNode>('circle')
      .data(nodes)
      .join('circle')
      .attr('r', d => nodeRadius(d.relCount))
      .attr('fill', d => TYPE_COLORS[d.entity_type] || '#64748b')
      .attr('stroke', d => d.id === selectedEntityId ? '#1e3a8a' : '#fff')
      .attr('stroke-width', d => d.id === selectedEntityId ? 3 : 1.5)
      .attr('cursor', 'pointer');

    // Labels for prominent nodes (≥3 relationships)
    const label = g.append('g')
      .selectAll<SVGTextElement, GraphNode>('text')
      .data(nodes.filter(n => n.relCount >= 3))
      .join('text')
      .attr('text-anchor', 'middle')
      .attr('font-size', '10px')
      .attr('fill', '#374151')
      .attr('pointer-events', 'none')
      .text(d => d.name);

    // Tooltip
    const tooltip = d3.select(document.body)
      .append('div')
      .style('position', 'fixed')
      .style('background', 'white')
      .style('border', '1px solid #e2e8f0')
      .style('border-radius', '6px')
      .style('padding', '8px 12px')
      .style('font-size', '12px')
      .style('pointer-events', 'none')
      .style('opacity', '0')
      .style('z-index', '1000')
      .style('box-shadow', '0 4px 6px -1px rgba(0,0,0,0.1)')
      .style('max-width', '200px');

    node
      .on('click', (event, d) => {
        event.stopPropagation();
        onEntitySelect(d.id === selectedEntityId ? null : d.id);
      })
      .on('mouseenter', (_event, d) => {
        tooltip
          .style('opacity', '1')
          .html(
            `<strong>${d.name}</strong><br/><span style="color:#64748b">${d.entity_type} · ${d.relCount} relationship${d.relCount !== 1 ? 's' : ''}</span>`
          );
      })
      .on('mousemove', event => {
        tooltip
          .style('left', event.clientX + 14 + 'px')
          .style('top', event.clientY - 32 + 'px');
      })
      .on('mouseleave', () => tooltip.style('opacity', '0'));

    // Edge tooltip on hover
    link
      .on('mouseenter', (_event, d) => {
        const src = d.source as GraphNode;
        const tgt = d.target as GraphNode;
        tooltip
          .style('opacity', '1')
          .html(
            `<strong>${src.name}</strong><span style="color:#64748b"> ${d.relationship_type} </span><strong>${tgt.name}</strong>${d.description ? `<br/><span style="color:#94a3b8;font-size:11px">${d.description}</span>` : ''}`
          );
      })
      .on('mousemove', event => {
        tooltip
          .style('left', event.clientX + 14 + 'px')
          .style('top', event.clientY - 32 + 'px');
      })
      .on('mouseleave', () => tooltip.style('opacity', '0'));

    svg.on('click', () => onEntitySelect(null));

    // Drag
    node.call(
      d3.drag<SVGCircleElement, GraphNode>()
        .on('start', (event, d) => {
          if (!event.active) simulation.alphaTarget(0.3).restart();
          d.fx = d.x;
          d.fy = d.y;
        })
        .on('drag', (event, d) => {
          d.fx = event.x;
          d.fy = event.y;
        })
        .on('end', (event, d) => {
          if (!event.active) simulation.alphaTarget(0);
          d.fx = null;
          d.fy = null;
        })
    );

    // Force simulation
    const simulation = d3.forceSimulation<GraphNode>(nodes)
      .force(
        'link',
        d3.forceLink<GraphNode, GraphLink>(links)
          .id(d => d.id)
          .distance(d => {
            const src = d.source as GraphNode;
            const tgt = d.target as GraphNode;
            return 50 + nodeRadius(src.relCount) + nodeRadius(tgt.relCount);
          })
      )
      .force('charge', d3.forceManyBody<GraphNode>().strength(d => -80 - d.relCount * 10))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collide', d3.forceCollide<GraphNode>(d => nodeRadius(d.relCount) + 6));

    simulationRef.current = simulation;

    simulation.on('tick', () => {
      link
        .attr('x1', d => (d.source as GraphNode).x ?? 0)
        .attr('y1', d => (d.source as GraphNode).y ?? 0)
        .attr('x2', d => (d.target as GraphNode).x ?? 0)
        .attr('y2', d => (d.target as GraphNode).y ?? 0);

      node
        .attr('cx', d => d.x ?? 0)
        .attr('cy', d => d.y ?? 0);

      label
        .attr('x', d => d.x ?? 0)
        .attr('y', d => (d.y ?? 0) + nodeRadius(d.relCount) + 11);
    });

    simulation.on('end', fitView);

    return () => {
      simulation.stop();
      tooltip.remove();
    };
  }, [entities, relationships]);

  // Update node highlight separately when selection changes (no need to re-run simulation)
  useEffect(() => {
    if (!svgRef.current) return;
    d3.select(svgRef.current)
      .selectAll<SVGCircleElement, GraphNode>('circle')
      .attr('stroke', d => d.id === selectedEntityId ? '#1e3a8a' : '#fff')
      .attr('stroke-width', d => d.id === selectedEntityId ? 3 : 1.5)
      .attr('opacity', d =>
        selectedEntityId === null || d.id === selectedEntityId ? 1 : 0.35
      );
  }, [selectedEntityId]);

  return (
    <div ref={containerRef} className="relative w-full h-full bg-slate-50 rounded-lg overflow-hidden">
      <svg ref={svgRef} className="w-full h-full" style={{ cursor: 'grab' }} />
      {/* Legend */}
      <div className="absolute bottom-3 right-3 flex flex-wrap gap-1.5 max-w-xs">
        {Object.entries(TYPE_COLORS).map(([type, color]) => (
          <div key={type} className="flex items-center gap-1 bg-white/90 px-2 py-0.5 rounded text-xs shadow-sm">
            <div className="w-2 h-2 rounded-full shrink-0" style={{ background: color }} />
            <span className="text-slate-600 capitalize">{type}</span>
          </div>
        ))}
      </div>
      {/* Hint */}
      <div className="absolute top-3 right-3 text-xs text-slate-400 bg-white/80 px-2 py-1 rounded">
        Scroll to zoom · Drag to pan · Click node to filter
      </div>
    </div>
  );
};

export default RelationshipGraph;
