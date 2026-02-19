import { TimelineEvent, Entity, Relationship, ReviewProgress } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const timelineApi = {
  getTimeline: async (documentId: string): Promise<TimelineEvent[]> => {
    const response = await fetch(`${API_BASE_URL}/api/documents/${documentId}/timeline`);
    if (!response.ok) {
      throw new Error(`Failed to fetch timeline: ${response.statusText}`);
    }
    return response.json();
  },

  getEntities: async (documentId: string): Promise<Entity[]> => {
    const response = await fetch(`${API_BASE_URL}/api/documents/${documentId}/entities`);
    if (!response.ok) {
      throw new Error(`Failed to fetch entities: ${response.statusText}`);
    }
    return response.json();
  },

  getRelationships: async (documentId: string): Promise<Relationship[]> => {
    const response = await fetch(`${API_BASE_URL}/api/documents/${documentId}/relationships`);
    if (!response.ok) {
      throw new Error(`Failed to fetch relationships: ${response.statusText}`);
    }
    return response.json();
  },

  getInconsistencies: async (documentId: string) => {
    const response = await fetch(`${API_BASE_URL}/api/documents/${documentId}/inconsistencies`);
    if (!response.ok) {
      throw new Error(`Failed to fetch inconsistencies: ${response.statusText}`);
    }
    return response.json();
  },

  getReviewProgress: async (documentId: string): Promise<ReviewProgress> => {
    const response = await fetch(`${API_BASE_URL}/api/documents/${documentId}/review-progress`);
    if (!response.ok) throw new Error('Failed to fetch review progress');
    return response.json();
  },

  updateClaimReview: async (claimId: string, status: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/claims/${claimId}/review`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ review_status: status }),
    });
    if (!response.ok) throw new Error('Failed to update claim review status');
  },

  updateClaimData: async (
    claimId: string,
    data: { title: string; description: string | null; date_text: string | null; event_type: string | null; confidence: number }
  ): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/claims/${claimId}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    if (!response.ok) throw new Error('Failed to update claim');
  },

  updateEntityReview: async (entityId: string, status: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/entities/${entityId}/review`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ review_status: status }),
    });
    if (!response.ok) throw new Error('Failed to update entity review status');
  },

  updateRelationshipReview: async (relId: string, status: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/relationships/${relId}/review`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ review_status: status }),
    });
    if (!response.ok) throw new Error('Failed to update relationship review status');
  },

  resolveInconsistency: async (id: string, status: string, note: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/inconsistencies/${id}/resolve`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status, note }),
    });
    if (!response.ok) throw new Error('Failed to resolve inconsistency');
  },
};
