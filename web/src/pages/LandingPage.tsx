import { useState, useRef, useEffect, useCallback } from 'react';
import {
  listProjects,
  createProject,
  listDocuments,
  addDocumentToProject,
  Project,
  Document,
} from '../api/projects';

interface Props {
  onNavigateToProject: (projectId: string) => void;
}

type UploadPhase =
  | { name: 'idle' }
  | { name: 'uploading' }
  | { name: 'chunking'; docId: string }
  | { name: 'extracting'; docId: string; currentChunk: number; totalChunks: number; events: number; entities: number; percent: number; elapsedSec: number }
  | { name: 'error'; message: string; docId?: string; canRetry?: boolean };

interface ExtractionProgress {
  source_id: string;
  status: string;
  current_chunk: number;
  total_chunks: number;
  events_found: number;
  entities_found: number;
  percent_complete: number;
  elapsed_time_sec: number;
  error_message?: string;
}

// Design tokens (matching CSS variables)
const tokens = {
  surface0: '#FAFBFC',
  surface1: '#FFFFFF',
  surface2: '#F3F4F6',
  surface3: '#E8EAED',
  textPrimary: '#1A1D23',
  textSecondary: '#4B5162',
  textTertiary: '#8B90A0',
  borderDefault: '#E2E4E9',
  borderSubtle: '#ECEEF1',
  accentPrimary: '#3B6FED',
  accentPrimaryBg: '#EEF2FF',
  confHigh: '#16A34A',
  confConflict: '#DC2626',
  confConflictBg: '#FEF2F2',
  shadowSm: '0 1px 2px rgba(0,0,0,0.04)',
  shadowMd: '0 2px 8px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04)',
};

export default function LandingPage({ onNavigateToProject }: Props) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newProjectTitle, setNewProjectTitle] = useState('');
  const [newProjectDesc, setNewProjectDesc] = useState('');
  const [creating, setCreating] = useState(false);

  // Upload state
  const [phase, setPhase] = useState<UploadPhase>({ name: 'idle' });
  const [isDragOver, setIsDragOver] = useState(false);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [availableDocs, setAvailableDocs] = useState<Document[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string>('');
  const fileInputRef = useRef<HTMLInputElement>(null);
  const pollTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Load projects on mount
  useEffect(() => {
    loadProjects();
  }, []);

  async function loadProjects() {
    try {
      setLoading(true);
      const data = await listProjects();
      setProjects(data);
    } catch (err) {
      console.error('Failed to load projects');
    } finally {
      setLoading(false);
    }
  }

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (pollTimerRef.current) clearInterval(pollTimerRef.current);
      const sseRef = (window as unknown as { sseRef?: EventSource }).sseRef;
      if (sseRef) sseRef.close();
    };
  }, []);

  async function handleCreateProject() {
    if (!newProjectTitle.trim()) return;

    try {
      setCreating(true);
      const project = await createProject(newProjectTitle.trim(), newProjectDesc.trim());
      setProjects([...projects, project]);
      setShowCreateModal(false);
      setNewProjectTitle('');
      setNewProjectDesc('');
    } catch (err) {
      alert('Failed to create project');
    } finally {
      setCreating(false);
    }
  }

  async function handleOpenUploadModal() {
    try {
      const docs = await listDocuments();
      setAvailableDocs(docs);
      setSelectedProjectId(projects.length > 0 ? projects[0].id : '');
      setShowUploadModal(true);
    } catch (err) {
      console.error('Failed to load documents');
    }
  }

  const streamExtractionProgress = useCallback((docId: string, projectId: string) => {
    const eventSource = new EventSource(`/api/documents/${docId}/extract/progress`);
    let hasCompleted = false;

    eventSource.onmessage = (event) => {
      try {
        const progress: ExtractionProgress = JSON.parse(event.data);

        if (progress.status === 'complete' || progress.percent_complete >= 100) {
          hasCompleted = true;
          eventSource.close();
          // Add to project and navigate
          addDocumentToProject(projectId, docId).then(() => {
            onNavigateToProject(projectId);
          });
          return;
        }

        if (progress.status === 'error') {
          eventSource.close();
          setPhase({
            name: 'error',
            message: progress.error_message || 'Extraction failed',
            docId,
            canRetry: true,
          });
          return;
        }

        setPhase({
          name: 'extracting',
          docId,
          currentChunk: progress.current_chunk,
          totalChunks: progress.total_chunks,
          events: progress.events_found,
          entities: progress.entities_found,
          percent: progress.percent_complete,
          elapsedSec: progress.elapsed_time_sec,
        });
      } catch {
        // Ignore parse errors
      }
    };

    eventSource.onerror = () => {
      if (hasCompleted) {
        eventSource.close();
        return;
      }
      eventSource.close();
      setPhase({
        name: 'error',
        message: 'Connection lost during extraction',
        docId,
        canRetry: true,
      });
    };

    (window as unknown as { sseRef?: EventSource }).sseRef = eventSource;
  }, [onNavigateToProject]);

  const processFile = useCallback(async (file: File, projectId: string) => {
    if (!file.name.match(/\.(txt|pdf|md)$/i)) {
      setPhase({ name: 'error', message: 'Only .txt, .pdf, and .md files are supported.' });
      return;
    }
    if (file.size > 50 * 1024 * 1024) {
      setPhase({ name: 'error', message: 'File must be under 50 MB.' });
      return;
    }

    setPhase({ name: 'uploading' });

    try {
      // Upload
      const formData = new FormData();
      formData.append('file', file);
      const uploadRes = await fetch('/api/documents', { method: 'POST', body: formData });
      if (!uploadRes.ok) {
        const msg = await uploadRes.text();
        throw new Error(msg || 'Upload failed');
      }
      const doc = await uploadRes.json() as { id: string };
      const docId = doc.id;

      setPhase({ name: 'chunking', docId });

      // Poll until chunking complete
      await new Promise<void>((resolve, reject) => {
        const t = setInterval(async () => {
          try {
            const d = await fetch(`/api/documents/${docId}/status`).then(r => r.json()) as { upload_status: string };
            if (d.upload_status === 'ready') {
              clearInterval(t);
              resolve();
            } else if (d.upload_status === 'error') {
              clearInterval(t);
              reject(new Error('Document processing failed'));
            }
          } catch (e) {
            clearInterval(t);
            reject(e);
          }
        }, 1500);
      });

      // Trigger extraction
      await fetch(`/api/documents/${docId}/extract`, { method: 'POST' });
      setPhase({
        name: 'extracting',
        docId,
        currentChunk: 0,
        totalChunks: 0,
        events: 0,
        entities: 0,
        percent: 0,
        elapsedSec: 0,
      });

      streamExtractionProgress(docId, projectId);
    } catch (err) {
      setPhase({ name: 'error', message: err instanceof Error ? err.message : 'Upload failed' });
    }
  }, [streamExtractionProgress]);

  async function handleAddExistingDoc(docId: string, projectId: string) {
    try {
      await addDocumentToProject(projectId, docId);
      setShowUploadModal(false);
      onNavigateToProject(projectId);
    } catch (err) {
      alert('Failed to add document to project');
    }
  }

  const isProcessing = phase.name === 'uploading' || phase.name === 'chunking' || phase.name === 'extracting';

  return (
    <div style={{ fontFamily: "'DM Sans', sans-serif", backgroundColor: tokens.surface0, minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '16px 32px',
        borderBottom: `1px solid ${tokens.borderSubtle}`,
        backgroundColor: tokens.surface1,
      }}>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 10 }}>
          <span style={{ fontSize: 22, fontWeight: 700, color: tokens.textPrimary, letterSpacing: '-0.02em' }}>Sikta</span>
          <span style={{ fontSize: 13, color: tokens.textTertiary, fontWeight: 500 }}>Evidence Synthesis Engine</span>
        </div>
        <div style={{ display: 'flex', gap: 12 }}>
          <button
            onClick={handleOpenUploadModal}
            style={{
              padding: '8px 20px',
              borderRadius: 8,
              backgroundColor: tokens.surface2,
              color: tokens.textPrimary,
              border: `1px solid ${tokens.borderDefault}`,
              fontFamily: "'DM Sans', sans-serif",
              fontSize: 13,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Add Document
          </button>
          <button
            onClick={() => setShowCreateModal(true)}
            style={{
              padding: '8px 20px',
              borderRadius: 8,
              backgroundColor: tokens.accentPrimary,
              color: '#fff',
              border: 'none',
              fontFamily: "'DM Sans', sans-serif",
              fontSize: 13,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            + New Project
          </button>
        </div>
      </div>

      {/* Main content */}
      <div style={{ maxWidth: 1100, margin: '0 auto', padding: '40px 32px', flex: 1 }}>
        <h1 style={{
          fontSize: 36,
          fontWeight: 700,
          color: tokens.textPrimary,
          lineHeight: 1.2,
          letterSpacing: '-0.02em',
          marginBottom: 8,
        }}>
          Projects
        </h1>
        <p style={{
          fontSize: 16,
          color: tokens.textSecondary,
          marginBottom: 32,
        }}>
          Organize documents into projects for cross-document analysis.
        </p>

        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: '60px 0' }}>
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
          </div>
        ) : projects.length === 0 ? (
          <div style={{
            padding: '60px 40px',
            backgroundColor: tokens.surface1,
            borderRadius: 12,
            border: `1px dashed ${tokens.borderDefault}`,
            textAlign: 'center',
          }}>
            <div style={{ fontSize: 48, marginBottom: 16 }}>📁</div>
            <h3 style={{ fontSize: 18, fontWeight: 600, color: tokens.textPrimary, marginBottom: 8 }}>
              No projects yet
            </h3>
            <p style={{ fontSize: 14, color: tokens.textTertiary, marginBottom: 24 }}>
              Create a project to start organizing your documents.
            </p>
            <button
              onClick={() => setShowCreateModal(true)}
              style={{
                padding: '10px 24px',
                borderRadius: 8,
                backgroundColor: tokens.accentPrimary,
                color: '#fff',
                border: 'none',
                fontFamily: "'DM Sans', sans-serif",
                fontSize: 14,
                fontWeight: 600,
                cursor: 'pointer',
              }}
            >
              Create Your First Project
            </button>
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 16 }}>
            {projects.map(project => (
              <div
                key={project.id}
                onClick={() => onNavigateToProject(project.id)}
                style={{
                  padding: '20px 24px',
                  backgroundColor: tokens.surface1,
                  borderRadius: 12,
                  border: `1px solid ${tokens.borderDefault}`,
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                  boxShadow: tokens.shadowSm,
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = tokens.accentPrimary;
                  e.currentTarget.style.boxShadow = tokens.shadowMd;
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = tokens.borderDefault;
                  e.currentTarget.style.boxShadow = tokens.shadowSm;
                }}
              >
                <h3 style={{ fontSize: 16, fontWeight: 600, color: tokens.textPrimary, marginBottom: 4 }}>
                  {project.title}
                </h3>
                {project.description && (
                  <p style={{ fontSize: 13, color: tokens.textTertiary, marginBottom: 16, lineHeight: 1.5 }}>
                    {project.description}
                  </p>
                )}
                {project.stats && (
                  <div style={{ display: 'flex', gap: 16, fontSize: 12, color: tokens.textTertiary }}>
                    <span>
                      <strong style={{ color: tokens.textPrimary }}>{project.stats.doc_count}</strong> docs
                    </span>
                    <span>
                      <strong style={{ color: tokens.textPrimary }}>{project.stats.node_count}</strong> nodes
                    </span>
                    <span>
                      <strong style={{ color: tokens.textPrimary }}>{project.stats.edge_count}</strong> edges
                    </span>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create Project Modal */}
      {showCreateModal && (
        <div
          style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50 }}
          onClick={(e) => { if (e.target === e.currentTarget) setShowCreateModal(false); }}
        >
          <div style={{ backgroundColor: tokens.surface1, borderRadius: 12, width: 420, maxWidth: '90vw', boxShadow: tokens.shadowMd }}>
            <div style={{ padding: '20px 24px', borderBottom: `1px solid ${tokens.borderDefault}` }}>
              <h2 style={{ fontSize: 18, fontWeight: 600, color: tokens.textPrimary }}>Create New Project</h2>
            </div>
            <div style={{ padding: 24 }}>
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: tokens.textSecondary, marginBottom: 6 }}>
                Project Title
              </label>
              <input
                type="text"
                value={newProjectTitle}
                onChange={(e) => setNewProjectTitle(e.target.value)}
                placeholder="e.g., Q1 Board Meetings"
                style={{
                  width: '100%',
                  padding: '10px 12px',
                  borderRadius: 8,
                  border: `1px solid ${tokens.borderDefault}`,
                  fontSize: 14,
                  marginBottom: 16,
                  boxSizing: 'border-box',
                }}
                autoFocus
              />
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: tokens.textSecondary, marginBottom: 6 }}>
                Description (optional)
              </label>
              <textarea
                value={newProjectDesc}
                onChange={(e) => setNewProjectDesc(e.target.value)}
                placeholder="Brief description of this project..."
                rows={3}
                style={{
                  width: '100%',
                  padding: '10px 12px',
                  borderRadius: 8,
                  border: `1px solid ${tokens.borderDefault}`,
                  fontSize: 14,
                  resize: 'vertical',
                  boxSizing: 'border-box',
                }}
              />
            </div>
            <div style={{ padding: '16px 24px', borderTop: `1px solid ${tokens.borderDefault}`, display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <button
                onClick={() => setShowCreateModal(false)}
                style={{
                  padding: '8px 16px',
                  borderRadius: 8,
                  backgroundColor: tokens.surface2,
                  color: tokens.textPrimary,
                  border: `1px solid ${tokens.borderDefault}`,
                  fontSize: 13,
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                Cancel
              </button>
              <button
                onClick={handleCreateProject}
                disabled={!newProjectTitle.trim() || creating}
                style={{
                  padding: '8px 16px',
                  borderRadius: 8,
                  backgroundColor: tokens.accentPrimary,
                  color: '#fff',
                  border: 'none',
                  fontSize: 13,
                  fontWeight: 500,
                  cursor: creating ? 'wait' : 'pointer',
                  opacity: newProjectTitle.trim() ? 1 : 0.5,
                }}
              >
                {creating ? 'Creating...' : 'Create Project'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Document Modal */}
      {showUploadModal && !isProcessing && phase.name !== 'error' && (
        <div
          style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50 }}
          onClick={(e) => { if (e.target === e.currentTarget) setShowUploadModal(false); }}
        >
          <div style={{ backgroundColor: tokens.surface1, borderRadius: 12, width: 520, maxWidth: '90vw', maxHeight: '80vh', overflow: 'hidden', boxShadow: tokens.shadowMd }}>
            <div style={{ padding: '20px 24px', borderBottom: `1px solid ${tokens.borderDefault}` }}>
              <h2 style={{ fontSize: 18, fontWeight: 600, color: tokens.textPrimary }}>Add Document to Project</h2>
            </div>
            <div style={{ padding: 24, overflowY: 'auto', maxHeight: 'calc(80vh - 140px)' }}>
              {/* Select project */}
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: tokens.textSecondary, marginBottom: 6 }}>
                Select Project
              </label>
              {projects.length === 0 ? (
                <p style={{ color: tokens.confConflict, fontSize: 14, marginBottom: 16 }}>
                  No projects available. Create a project first.
                </p>
              ) : (
                <select
                  value={selectedProjectId}
                  onChange={(e) => setSelectedProjectId(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '10px 12px',
                    borderRadius: 8,
                    border: `1px solid ${tokens.borderDefault}`,
                    fontSize: 14,
                    marginBottom: 20,
                    backgroundColor: tokens.surface1,
                  }}
                >
                  {projects.map(p => (
                    <option key={p.id} value={p.id}>{p.title}</option>
                  ))}
                </select>
              )}

              {/* Upload new document */}
              <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: tokens.textSecondary, marginBottom: 6 }}>
                Upload New Document
              </label>
              <div
                onDragOver={e => { e.preventDefault(); setIsDragOver(true); }}
                onDragLeave={() => setIsDragOver(false)}
                onDrop={(e) => {
                  e.preventDefault();
                  setIsDragOver(false);
                  const file = e.dataTransfer.files[0];
                  if (file && selectedProjectId) {
                    setShowUploadModal(false);
                    processFile(file, selectedProjectId);
                  }
                }}
                onClick={() => fileInputRef.current?.click()}
                style={{
                  border: `2px dashed ${isDragOver ? tokens.accentPrimary : tokens.borderDefault}`,
                  borderRadius: 8,
                  padding: '32px 20px',
                  textAlign: 'center',
                  cursor: 'pointer',
                  backgroundColor: isDragOver ? tokens.accentPrimaryBg : tokens.surface2,
                  marginBottom: 20,
                }}
              >
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".txt,.pdf,.md"
                  style={{ display: 'none' }}
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file && selectedProjectId) {
                      setShowUploadModal(false);
                      processFile(file, selectedProjectId);
                    }
                  }}
                />
                <div style={{ fontSize: 24, marginBottom: 8 }}>📄</div>
                <p style={{ color: tokens.textPrimary, fontSize: 14, fontWeight: 500 }}>Click or drag to upload</p>
                <p style={{ fontSize: 12, color: tokens.textTertiary }}>.txt, .pdf, or .md up to 50 MB</p>
              </div>

              {/* Existing documents */}
              {availableDocs.length > 0 && (
                <>
                  <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: tokens.textSecondary, marginBottom: 6 }}>
                    Or Add Existing Document
                  </label>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                    {availableDocs.slice(0, 5).map(doc => (
                      <div
                        key={doc.id}
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'center',
                          padding: '10px 12px',
                          backgroundColor: tokens.surface2,
                          borderRadius: 8,
                        }}
                      >
                        <div>
                          <div style={{ fontSize: 14, fontWeight: 500, color: tokens.textPrimary }}>{doc.title}</div>
                          <div style={{ fontSize: 12, color: tokens.textTertiary }}>{doc.filename}</div>
                        </div>
                        <button
                          onClick={() => handleAddExistingDoc(doc.id, selectedProjectId)}
                          disabled={!selectedProjectId}
                          style={{
                            padding: '6px 12px',
                            borderRadius: 6,
                            backgroundColor: tokens.accentPrimary,
                            color: '#fff',
                            border: 'none',
                            fontSize: 12,
                            fontWeight: 500,
                            cursor: 'pointer',
                            opacity: selectedProjectId ? 1 : 0.5,
                          }}
                        >
                          Add
                        </button>
                      </div>
                    ))}
                  </div>
                </>
              )}
            </div>
            <div style={{ padding: '16px 24px', borderTop: `1px solid ${tokens.borderDefault}`, display: 'flex', justifyContent: 'flex-end' }}>
              <button
                onClick={() => setShowUploadModal(false)}
                style={{
                  padding: '8px 16px',
                  borderRadius: 8,
                  backgroundColor: tokens.surface2,
                  color: tokens.textPrimary,
                  border: `1px solid ${tokens.borderDefault}`,
                  fontSize: 13,
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Processing overlay */}
      {isProcessing && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50 }}>
          <div style={{ backgroundColor: tokens.surface1, borderRadius: 12, width: 420, maxWidth: '90vw', boxShadow: tokens.shadowMd, padding: 24 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 16 }}>
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 shrink-0" />
              <div style={{ flex: 1 }}>
                <p style={{ fontSize: 15, fontWeight: 600, color: tokens.textPrimary }}>
                  {phase.name === 'uploading' && 'Uploading document...'}
                  {phase.name === 'chunking' && 'Splitting into sections...'}
                  {phase.name === 'extracting' && 'Extracting events with AI...'}
                </p>
                {phase.name === 'extracting' && (
                  <p style={{ fontSize: 13, color: tokens.textTertiary, marginTop: 2 }}>
                    {phase.totalChunks > 0
                      ? `Processing chunk ${phase.currentChunk} of ${phase.totalChunks}`
                      : 'Starting extraction pipeline...'}
                  </p>
                )}
              </div>
              {phase.name === 'extracting' && phase.elapsedSec > 0 && (
                <span style={{ fontSize: 13, color: tokens.textTertiary }}>
                  {Math.floor(phase.elapsedSec / 60)}:{(phase.elapsedSec % 60).toString().padStart(2, '0')}
                </span>
              )}
            </div>

            {phase.name === 'extracting' && phase.totalChunks > 0 && (
              <div style={{ marginBottom: 16 }}>
                <div style={{ height: 8, backgroundColor: tokens.surface2, borderRadius: 4, overflow: 'hidden' }}>
                  <div
                    style={{
                      height: '100%',
                      backgroundColor: tokens.accentPrimary,
                      borderRadius: 4,
                      transition: 'all 0.3s ease',
                      width: `${phase.percent}%`
                    }}
                  />
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 6, fontSize: 11, color: tokens.textTertiary }}>
                  <span>{phase.percent}% complete</span>
                  <span>{phase.events} events, {phase.entities} entities</span>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Error overlay */}
      {phase.name === 'error' && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50 }}>
          <div style={{ backgroundColor: tokens.surface1, borderRadius: 12, width: 420, maxWidth: '90vw', boxShadow: tokens.shadowMd }}>
            <div style={{ padding: 24 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
                <span style={{ color: tokens.confConflict, fontSize: 24 }}>✕</span>
                <p style={{ color: '#991B1B', fontSize: 15, fontWeight: 500 }}>{phase.message}</p>
              </div>
              <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                <button
                  onClick={() => setPhase({ name: 'idle' })}
                  style={{
                    padding: '8px 16px',
                    borderRadius: 8,
                    backgroundColor: tokens.surface2,
                    color: tokens.textPrimary,
                    border: `1px solid ${tokens.borderDefault}`,
                    fontSize: 13,
                    fontWeight: 500,
                    cursor: 'pointer',
                  }}
                >
                  Close
                </button>
                {phase.canRetry && phase.docId && (
                  <button
                    onClick={() => {
                      // Retry extraction
                      setPhase({ name: 'idle' });
                      handleOpenUploadModal();
                    }}
                    style={{
                      padding: '8px 16px',
                      borderRadius: 8,
                      backgroundColor: tokens.confConflict,
                      color: '#fff',
                      border: 'none',
                      fontSize: 13,
                      fontWeight: 500,
                      cursor: 'pointer',
                    }}
                  >
                    Try Again
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Footer */}
      <footer style={{
        borderTop: `1px solid ${tokens.borderDefault}`,
        backgroundColor: tokens.surface1,
        padding: '16px 32px',
        display: 'flex',
        justifyContent: 'space-between',
        fontSize: 12,
        color: tokens.textTertiary,
      }}>
        <span>Sikta — Evidence Synthesis</span>
        <span>Organize documents into projects for cross-document analysis</span>
      </footer>
    </div>
  );
}
