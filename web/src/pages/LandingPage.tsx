import React, { useState, useRef, useEffect, useCallback } from 'react';

interface DemoInfo {
  id: string;
  title: string;
  events: number;
  entities: number;
  relationships: number;
}

type UploadPhase =
  | { name: 'idle' }
  | { name: 'uploading' }
  | { name: 'chunking'; docId: string }
  | { name: 'extracting'; docId: string; currentChunk: number; totalChunks: number; events: number; entities: number; relationships: number; percent: number; elapsedSec: number; failedChunk?: number; errorMessage?: string }
  | { name: 'error'; message: string; docId?: string; canRetry?: boolean };

interface ExtractionProgress {
  source_id: string;
  status: string;
  current_chunk: number;
  total_chunks: number;
  events_found: number;
  entities_found: number;
  relationships_found: number;
  percent_complete: number;
  elapsed_time_sec: number;
  error_message?: string;
}

interface DocumentInfo {
  id: string;
  title: string;
  upload_status: string;
}

interface Props {
  onNavigate: (docId: string) => void;
  onNavigateToProjects?: () => void;
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

export default function LandingPage({ onNavigate, onNavigateToProjects }: Props) {
  const [demo, setDemo] = useState<DemoInfo | null>(null);
  const [demoLoading, setDemoLoading] = useState(true);
  const [allDocs, setAllDocs] = useState<DocumentInfo[]>([]);
  const [phase, setPhase] = useState<UploadPhase>({ name: 'idle' });
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const pollTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Load demo document info and all documents
  useEffect(() => {
    fetch('/api/documents')
      .then(r => (r.ok ? r.json() : []))
      .then(async (docs: DocumentInfo[]) => {
        setAllDocs(docs || []);
        if (!docs || docs.length === 0) return;
        const doc = docs[0];
        // Fetch review progress to get counts (extract/status endpoint was removed in G6)
        const progress = await fetch(`/api/documents/${doc.id}/review-progress`).then(r => r.json());
        setDemo({
          id: doc.id,
          title: doc.title,
          events: progress.claims?.total || 0,
          entities: progress.entities?.total || 0,
          relationships: 0, // No quick count available for relationships
        });
      })
      .catch(() => {})
      .finally(() => setDemoLoading(false));
  }, []);

  // Cleanup poll timer and SSE on unmount
  useEffect(() => {
    return () => {
      if (pollTimerRef.current) clearInterval(pollTimerRef.current);
      const sseRef = (window as unknown as { sseRef?: EventSource }).sseRef;
      if (sseRef) sseRef.close();
    };
  }, []);

  const streamExtractionProgress = useCallback((docId: string) => {
    const eventSource = new EventSource(`/api/documents/${docId}/extract/progress`);
    let hasCompleted = false;

    eventSource.onmessage = (event) => {
      try {
        const progress: ExtractionProgress = JSON.parse(event.data);

        // Check for completion first - most important
        if (progress.status === 'complete' || progress.percent_complete >= 100) {
          hasCompleted = true;
          eventSource.close();
          // Small delay to ensure data is committed
          setTimeout(() => onNavigate(docId), 500);
          return;
        }

        // Check for errors
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

        // Normal progress update
        setPhase({
          name: 'extracting',
          docId,
          currentChunk: progress.current_chunk,
          totalChunks: progress.total_chunks,
          events: progress.events_found,
          entities: progress.entities_found,
          relationships: progress.relationships_found,
          percent: progress.percent_complete,
          elapsedSec: progress.elapsed_time_sec,
        });
      } catch {
        // Ignore parse errors, keep listening
      }
    };

    eventSource.onerror = () => {
      // Don't close immediately on error - check if we already completed
      if (hasCompleted) {
        eventSource.close();
        return;
      }
      eventSource.close();
      // Fallback to polling if SSE fails
      pollExtractionStatus(docId);
    };

    // Store for cleanup
    pollTimerRef.current = setInterval(() => {}, 1000); // Dummy for cleanup tracking
    (window as unknown as { sseRef?: EventSource }).sseRef = eventSource;
  }, [onNavigate]);

  const pollExtractionStatus = useCallback((docId: string) => {
    let pollCount = 0;
    const maxPolls = 120; // 6 minutes max at 3s intervals

    // Fallback polling if SSE doesn't work
    pollTimerRef.current = setInterval(async () => {
      pollCount++;
      if (pollCount > maxPolls) {
        if (pollTimerRef.current) clearInterval(pollTimerRef.current);
        setPhase({ name: 'error', message: 'Extraction timed out', docId, canRetry: true });
        return;
      }

      try {
        const status = await fetch(`/api/documents/${docId}/extract/status`).then(r => r.json());
        const events = status.events || 0;
        const totalChunks = status.total_chunks || 0;
        const percent = status.percent_complete || (events > 0 ? 50 : 0);

        setPhase(prev => ({
          ...prev,
          name: 'extracting',
          docId,
          currentChunk: status.current_chunk || 0,
          totalChunks,
          events,
          entities: status.entities_found || 0,
          relationships: status.relationships_found || 0,
          percent,
          elapsedSec: status.elapsed_time_sec || 0,
        } as UploadPhase));

        // Check for completion - be more lenient
        if (status.status === 'complete' || percent >= 100 || (events > 0 && status.status !== 'processing')) {
          if (pollTimerRef.current) clearInterval(pollTimerRef.current);
          onNavigate(docId);
        }
      } catch {
        // keep polling
      }
    }, 3000);
  }, [onNavigate]);

  const processFile = useCallback(async (file: File) => {
    if (!file.name.match(/\.(txt|pdf)$/i)) {
      setPhase({ name: 'error', message: 'Only .txt and .pdf files are supported.' });
      return;
    }
    if (file.size > 50 * 1024 * 1024) {
      setPhase({ name: 'error', message: 'File must be under 50 MB.' });
      return;
    }

    setPhase({ name: 'uploading' });

    try {
      // 1. Upload
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

      // 2. Poll /status until chunking complete (status = 'ready')
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

      // 3. Trigger extraction
      await fetch(`/api/documents/${docId}/extract`, { method: 'POST' });
      setPhase({
        name: 'extracting',
        docId,
        currentChunk: 0,
        totalChunks: 0,
        events: 0,
        entities: 0,
        relationships: 0,
        percent: 0,
        elapsedSec: 0,
      });

      // 4. Stream extraction progress via SSE
      streamExtractionProgress(docId);
    } catch (err) {
      setPhase({ name: 'error', message: err instanceof Error ? err.message : 'Upload failed' });
    }
  }, [streamExtractionProgress]);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    const file = e.dataTransfer.files[0];
    if (file) processFile(file);
  }, [processFile]);

  const handleFileInput = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) processFile(file);
    e.target.value = '';
  }, [processFile]);

  // Retry extraction after failure
  const retryExtraction = useCallback(async (docId: string) => {
    setPhase({
      name: 'extracting',
      docId,
      currentChunk: 0,
      totalChunks: 0,
      events: 0,
      entities: 0,
      relationships: 0,
      percent: 0,
      elapsedSec: 0,
    });

    try {
      // Trigger extraction again
      await fetch(`/api/documents/${docId}/extract`, { method: 'POST' });
      streamExtractionProgress(docId);
    } catch (err) {
      setPhase({
        name: 'error',
        message: err instanceof Error ? err.message : 'Retry failed',
        docId,
        canRetry: true,
      });
    }
  }, [streamExtractionProgress]);

  const handleDeleteDocument = useCallback(async (docId: string) => {
    if (!confirm('Are you sure you want to delete this document? This cannot be undone.')) {
      return;
    }
    try {
      const res = await fetch(`/api/documents/${docId}`, { method: 'DELETE' });
      if (!res.ok) {
        throw new Error('Failed to delete document');
      }
      // Remove from local state
      setAllDocs(prev => prev.filter(d => d.id !== docId));
      // If we deleted the demo doc, clear it and try to set a new one
      if (demo?.id === docId) {
        setDemo(null);
        const remaining = allDocs.filter(d => d.id !== docId);
        if (remaining.length > 0) {
          const newDemo = remaining[0];
          const progress = await fetch(`/api/documents/${newDemo.id}/review-progress`).then(r => r.json());
          setDemo({
            id: newDemo.id,
            title: newDemo.title,
            events: progress.claims?.total || 0,
            entities: progress.entities?.total || 0,
            relationships: 0,
          });
        }
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete document');
    }
  }, [demo, allDocs]);

  const isProcessing = phase.name === 'uploading' || phase.name === 'chunking' || phase.name === 'extracting';

  // Other ready documents (excluding the main demo)
  const otherDocs = allDocs.filter(d => d.upload_status === 'ready' && d.id !== demo?.id);

  // Helper for StatCard
  const StatCard = ({ value, label, color, icon }: { value: number | string; label: string; color?: string; icon?: string }) => (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      padding: '20px 32px',
      backgroundColor: tokens.surface1,
      borderRadius: 12,
      border: `1px solid ${tokens.borderDefault}`,
      minWidth: 140,
      boxShadow: tokens.shadowSm,
    }}>
      <span style={{
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: 36,
        fontWeight: 700,
        color: color || tokens.textPrimary,
        lineHeight: 1,
      }}>
        {value}
      </span>
      <span style={{
        fontFamily: "'DM Sans', sans-serif",
        fontSize: 13,
        color: tokens.textTertiary,
        marginTop: 4,
        display: 'flex',
        alignItems: 'center',
        gap: 4,
      }}>
        {label} {icon}
      </span>
    </div>
  );

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
        {demo && (
          <div style={{ display: 'flex', gap: 12 }}>
            {onNavigateToProjects && (
              <button
                onClick={onNavigateToProjects}
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
                Projects
              </button>
            )}
            <button
              onClick={() => onNavigate(demo.id)}
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
              Explore Demo →
            </button>
          </div>
        )}
      </div>

      {/* Hero */}
      <div style={{ maxWidth: 1100, margin: '0 auto', padding: '60px 32px 40px', flex: 1 }}>
        <h1 style={{
          fontSize: 48,
          fontWeight: 700,
          color: tokens.textPrimary,
          lineHeight: 1.1,
          letterSpacing: '-0.03em',
          marginBottom: 16,
        }}>
          Turn documents into<br />
          <span style={{ color: tokens.accentPrimary }}>auditable evidence.</span>
        </h1>
        <p style={{
          fontSize: 18,
          color: tokens.textSecondary,
          lineHeight: 1.6,
          maxWidth: 520,
          marginBottom: 48,
        }}>
          Every claim traced to its source.<br />
          Every contradiction surfaced.
        </p>

        {/* Stats */}
        {demoLoading ? (
          <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 40 }}>
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600" />
          </div>
        ) : demo ? (
          <div style={{ marginBottom: 40 }}>
            {/* Demo document title with delete */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span style={{ fontSize: 24 }}>📖</span>
                <span style={{ fontSize: 15, fontWeight: 600, color: tokens.textPrimary }}>{demo.title}</span>
              </div>
              <button
                onClick={() => handleDeleteDocument(demo.id)}
                style={{
                  padding: '4px 8px',
                  borderRadius: 6,
                  backgroundColor: tokens.confConflictBg,
                  color: tokens.confConflict,
                  border: 'none',
                  fontFamily: "'DM Sans', sans-serif",
                  fontSize: 11,
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
                title="Delete document"
              >
                ✕ Delete
              </button>
            </div>
            <div style={{ display: 'flex', gap: 16 }}>
              <StatCard value={demo.events} label="claims extracted" />
              <StatCard value={demo.entities} label="entities identified" />
              <StatCard value="—" label="conflicts" color={tokens.textTertiary} />
            </div>
          </div>
        ) : null}

        {/* Other novels available */}
        {otherDocs.length > 0 && (
          <div style={{
            padding: '20px 24px',
            backgroundColor: tokens.surface1,
            borderRadius: 12,
            border: `1px solid ${tokens.borderDefault}`,
            marginBottom: 32,
          }}>
            <div style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 10, fontWeight: 600, color: tokens.textTertiary, letterSpacing: '0.08em', marginBottom: 12 }}>
              OTHER NOVELS AVAILABLE
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {otherDocs.map(doc => (
                <div key={doc.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '8px 12px', borderRadius: 8, backgroundColor: tokens.surface2 }}>
                  <span style={{ fontSize: 14, color: tokens.textPrimary, flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{doc.title}</span>
                  <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                    <button
                      onClick={() => onNavigate(doc.id)}
                      style={{
                        padding: '4px 12px',
                        borderRadius: 6,
                        backgroundColor: tokens.accentPrimary,
                        color: '#fff',
                        border: 'none',
                        fontFamily: "'DM Sans', sans-serif",
                        fontSize: 11,
                        fontWeight: 600,
                        cursor: 'pointer',
                      }}
                    >
                      View →
                    </button>
                    <button
                      onClick={() => handleDeleteDocument(doc.id)}
                      style={{
                        padding: '4px 8px',
                        borderRadius: 6,
                        backgroundColor: tokens.confConflictBg,
                        color: tokens.confConflict,
                        border: 'none',
                        fontFamily: "'DM Sans', sans-serif",
                        fontSize: 11,
                        fontWeight: 500,
                        cursor: 'pointer',
                      }}
                      title="Delete document"
                    >
                      ✕
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Upload zone */}
        {!isProcessing && phase.name !== 'error' && (
          <div
            onDragOver={e => { e.preventDefault(); setIsDragOver(true); }}
            onDragLeave={() => setIsDragOver(false)}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            style={{
              border: `2px dashed ${isDragOver ? '#3B6FED' : tokens.borderDefault}`,
              borderRadius: 12,
              padding: '48px 32px',
              textAlign: 'center',
              cursor: 'pointer',
              transition: 'all 0.15s ease',
              marginBottom: 32,
              backgroundColor: isDragOver ? '#EEF2FF' : tokens.surface1,
            }}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept=".txt,.pdf"
              style={{ display: 'none' }}
              onChange={handleFileInput}
            />
            <div style={{ fontSize: 32, marginBottom: 12 }}>📄</div>
            <p style={{ color: tokens.textPrimary, fontSize: 15, fontWeight: 500, marginBottom: 4 }}>
              Drag & drop a file here
            </p>
            <p style={{ fontSize: 13, color: tokens.textTertiary }}>
              .txt or .pdf · up to 50 MB · click to browse
            </p>
          </div>
        )}

        {/* Processing state */}
        {isProcessing && (
          <div style={{
            padding: '24px',
            backgroundColor: tokens.surface1,
            borderRadius: 12,
            border: `1px solid ${tokens.borderDefault}`,
            marginBottom: 32,
            boxShadow: tokens.shadowSm,
          }}>
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

            {/* Progress bar */}
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
                  <span>
                    {phase.totalChunks - phase.currentChunk} chunks remaining
                    {phase.elapsedSec > 0 && phase.currentChunk > 0 && phase.currentChunk < phase.totalChunks && (
                      <> · ~{Math.round((phase.elapsedSec / phase.currentChunk) * (phase.totalChunks - phase.currentChunk) / 60)}m left</>
                    )}
                  </span>
                </div>
              </div>
            )}

            {/* Extraction stats */}
            {phase.name === 'extracting' && (phase.events > 0 || phase.entities > 0) && (
              <div style={{ display: 'flex', gap: 24, marginBottom: 16 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: tokens.accentPrimary }}></span>
                  <span style={{ color: tokens.textSecondary, fontSize: 14 }}><strong style={{ color: tokens.textPrimary }}>{phase.events}</strong> events</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: tokens.confHigh }}></span>
                  <span style={{ color: tokens.textSecondary, fontSize: 14 }}><strong style={{ color: tokens.textPrimary }}>{phase.entities}</strong> entities</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: '#8B5CF6' }}></span>
                  <span style={{ color: tokens.textSecondary, fontSize: 14 }}><strong style={{ color: tokens.textPrimary }}>{phase.relationships}</strong> relationships</span>
                </div>
              </div>
            )}

            {/* Early access button */}
            {phase.name === 'extracting' && phase.events > 0 && (
              <button
                onClick={() => {
                  if (pollTimerRef.current) clearInterval(pollTimerRef.current);
                  const sseRef = (window as unknown as { sseRef?: EventSource }).sseRef;
                  if (sseRef) sseRef.close();
                  onNavigate((phase as { name: 'extracting'; docId: string }).docId);
                }}
                style={{
                  width: '100%',
                  padding: '10px 16px',
                  border: `1px solid ${tokens.accentPrimary}`,
                  borderRadius: 8,
                  backgroundColor: tokens.surface1,
                  color: tokens.accentPrimary,
                  fontFamily: "'DM Sans', sans-serif",
                  fontSize: 13,
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                View partial results now →
              </button>
            )}
          </div>
        )}

        {/* Error state */}
        {phase.name === 'error' && (
          <div style={{
            padding: '20px 24px',
            backgroundColor: '#FEF2F2',
            borderRadius: 12,
            border: '1px solid #FECACA',
            marginBottom: 32,
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <span style={{ color: '#DC2626', fontSize: 18 }}>✕</span>
              <div style={{ flex: 1 }}>
                <p style={{ color: '#991B1B', fontSize: 14, fontWeight: 500 }}>{phase.message}</p>
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8, marginTop: 12, justifyContent: 'flex-end' }}>
              <button
                onClick={() => setPhase({ name: 'idle' })}
                style={{
                  padding: '6px 12px',
                  border: '1px solid #E5E7EB',
                  borderRadius: 6,
                  backgroundColor: '#fff',
                  color: '#6B7280',
                  fontFamily: "'DM Sans', sans-serif",
                  fontSize: 12,
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                Upload different file
              </button>
              {phase.canRetry && phase.docId && (
                <button
                  onClick={() => retryExtraction(phase.docId!)}
                  style={{
                    padding: '6px 12px',
                    border: 'none',
                    borderRadius: 6,
                    backgroundColor: '#DC2626',
                    color: '#fff',
                    fontFamily: "'DM Sans', sans-serif",
                    fontSize: 12,
                    fontWeight: 500,
                    cursor: 'pointer',
                  }}
                >
                  Retry extraction
                </button>
              )}
            </div>
          </div>
        )}

        {/* Works with section */}
        <div style={{
          padding: '24px 32px',
          backgroundColor: tokens.surface1,
          borderRadius: 12,
          border: `1px solid ${tokens.borderDefault}`,
        }}>
          <div style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 10, fontWeight: 600, color: tokens.textTertiary, letterSpacing: '0.08em', marginBottom: 12 }}>
            WORKS WITH
          </div>
          <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
            {["Board protocols", "Contracts & M&A", "Case files", "Research papers", "Novels & narratives", "Any text"].map(t => (
              <span key={t} style={{ fontFamily: "'DM Sans', sans-serif", fontSize: 14, color: tokens.textSecondary }}>
                {t}
              </span>
            ))}
          </div>
        </div>
      </div>

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
        <span>Demo: <em>Pride and Prejudice</em>, Jane Austen (public domain)</span>
      </footer>
    </div>
  );
}
