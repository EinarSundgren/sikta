// API client for project operations

export interface Project {
  id: string;
  title: string;
  description: string;
  created_at: string;
  updated_at: string;
  stats?: {
    doc_count: number;
    node_count: number;
    edge_count: number;
  };
}

export interface Document {
  id: string;
  title: string;
  filename: string;
  file_type: string;
  total_pages: number;
  upload_status: string;
  is_demo: boolean;
  created_at: string;
  updated_at: string;
}

export interface Node {
  id: string;
  node_type: string;
  label: string;
  properties: Record<string, unknown>;
  created_at: string;
}

export interface Edge {
  id: string;
  edge_type: string;
  source_node: string;
  target_node: string;
  properties: Record<string, unknown>;
  created_at: string;
}

export interface Graph {
  nodes: Node[];
  edges: Edge[];
}

export interface PostProcessResult {
  deduplication?: { matches: number };
  inconsistencies?: { count: number };
}

const API_BASE = '';

export async function listProjects(): Promise<Project[]> {
  const res = await fetch(`${API_BASE}/api/projects`);
  if (!res.ok) throw new Error('Failed to fetch projects');
  return res.json();
}

export async function getProject(id: string): Promise<Project> {
  const res = await fetch(`${API_BASE}/api/projects/${id}`);
  if (!res.ok) throw new Error('Failed to fetch project');
  return res.json();
}

export async function createProject(title: string, description: string): Promise<Project> {
  const res = await fetch(`${API_BASE}/api/projects`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title, description }),
  });
  if (!res.ok) throw new Error('Failed to create project');
  return res.json();
}

export async function getProjectDocuments(projectId: string): Promise<Document[]> {
  const res = await fetch(`${API_BASE}/api/projects/${projectId}/documents`);
  if (!res.ok) throw new Error('Failed to fetch project documents');
  return res.json();
}

export async function addDocumentToProject(projectId: string, documentId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/projects/${projectId}/documents`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ document_id: documentId }),
  });
  if (!res.ok) throw new Error('Failed to add document to project');
}

export async function getProjectGraph(projectId: string): Promise<Graph> {
  const res = await fetch(`${API_BASE}/api/projects/${projectId}/graph`);
  if (!res.ok) throw new Error('Failed to fetch project graph');
  return res.json();
}

export async function runPostProcessing(projectId: string): Promise<PostProcessResult> {
  const res = await fetch(`${API_BASE}/api/projects/${projectId}/postprocess`, {
    method: 'POST',
  });
  if (!res.ok) throw new Error('Failed to run post-processing');
  return res.json();
}

export async function listDocuments(): Promise<Document[]> {
  const res = await fetch(`${API_BASE}/api/documents`);
  if (!res.ok) throw new Error('Failed to fetch documents');
  return res.json();
}

export async function uploadDocument(file: File): Promise<Document> {
  const formData = new FormData();
  formData.append('file', file);

  const res = await fetch(`${API_BASE}/api/documents`, {
    method: 'POST',
    body: formData,
  });
  if (!res.ok) throw new Error('Failed to upload document');
  return res.json();
}

export async function extractDocument(docId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/documents/${docId}/extract`, {
    method: 'POST',
  });
  if (!res.ok) throw new Error('Failed to trigger extraction');
}
