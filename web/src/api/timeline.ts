import { TimelineEvent } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const timelineApi = {
  getTimeline: async (documentId: string): Promise<TimelineEvent[]> => {
    const response = await fetch(`${API_BASE_URL}/api/documents/${documentId}/timeline`);
    if (!response.ok) {
      throw new Error(`Failed to fetch timeline: ${response.statusText}`);
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
};
