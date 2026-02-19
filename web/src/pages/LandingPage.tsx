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
  | { name: 'extracting'; docId: string; currentChunk: number; totalChunks: number; events: number; entities: number; relationships: number; percent: number; elapsedSec: number }
  | { name: 'error'; message: string };

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
}

interface Props {
  onNavigate: (docId: string) => void;
}

const FEATURES = [
  { icon: '⟷', label: 'Dual-lane timeline', detail: 'Chronological and narrative order simultaneously' },
  { icon: '◎', label: 'Entity relationship graph', detail: 'D3 force-directed network of characters and places' },
  { icon: '⚡', label: 'Inconsistency detection', detail: 'Contradictions and temporal impossibilities surfaced automatically' },
  { icon: '✓', label: 'Human review workflow', detail: 'Keyboard-driven approve / reject / edit with J K A R E' },
];

export default function LandingPage({ onNavigate }: Props) {
  const [demo, setDemo] = useState<DemoInfo | null>(null);
  const [demoLoading, setDemoLoading] = useState(true);
  const [phase, setPhase] = useState<UploadPhase>({ name: 'idle' });
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const pollTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Load demo document info
  useEffect(() => {
    fetch('/api/documents')
      .then(r => (r.ok ? r.json() : []))
      .then(async (docs: { id: string; title: string }[]) => {
        if (!docs || docs.length === 0) return;
        const doc = docs[0];
        // Fetch extraction status to get counts
        const status = await fetch(`/api/documents/${doc.id}/extract/status`).then(r => r.json());
        setDemo({
          id: doc.id,
          title: doc.title,
          events: status.events || 0,
          entities: status.entities || 0,
          relationships: status.relationships || 0,
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

    eventSource.onmessage = (event) => {
      try {
        const progress: ExtractionProgress = JSON.parse(event.data);

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

        if (progress.status === 'complete') {
          eventSource.close();
          onNavigate(docId);
        } else if (progress.status === 'error') {
          eventSource.close();
          setPhase({ name: 'error', message: 'Extraction failed' });
        }
      } catch {
        // Ignore parse errors
      }
    };

    eventSource.onerror = () => {
      eventSource.close();
      // Fallback to polling if SSE fails
      pollExtractionStatus(docId);
    };

    // Store for cleanup
    pollTimerRef.current = setInterval(() => {}, 1000); // Dummy for cleanup tracking
    (window as unknown as { sseRef?: EventSource }).sseRef = eventSource;
  }, [onNavigate]);

  const pollExtractionStatus = useCallback((docId: string) => {
    // Fallback polling if SSE doesn't work
    pollTimerRef.current = setInterval(async () => {
      try {
        const status = await fetch(`/api/documents/${docId}/extract/status`).then(r => r.json());
        const events = status.events || 0;
        const totalChunks = status.total_chunks || 0;
        setPhase({
          name: 'extracting',
          docId,
          currentChunk: 0,
          totalChunks,
          events,
          entities: 0,
          relationships: 0,
          percent: events > 0 ? 50 : 0,
          elapsedSec: 0,
        });
        if (status.status === 'complete' && events > 0) {
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

  const isProcessing = phase.name === 'uploading' || phase.name === 'chunking' || phase.name === 'extracting';

  return (
    <div className="min-h-screen bg-slate-50 flex flex-col">
      {/* Nav */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between">
          <div>
            <span className="text-xl font-bold text-slate-900 tracking-tight">Sikta</span>
            <span className="ml-2 text-sm text-slate-400">Document Timeline Intelligence</span>
          </div>
          {demo && (
            <button
              onClick={() => onNavigate(demo.id)}
              className="px-4 py-1.5 text-sm font-medium bg-slate-900 text-white rounded-lg hover:bg-slate-700 transition-colors"
            >
              Explore Demo →
            </button>
          )}
        </div>
      </header>

      {/* Hero */}
      <main className="flex-1 max-w-4xl mx-auto px-6 py-16 w-full">
        <div className="text-center mb-14">
          <h1 className="text-5xl font-bold text-slate-900 tracking-tight mb-5 leading-tight">
            Make documents<br />
            <span className="text-blue-600">auditable and navigable</span>
          </h1>
          <p className="text-xl text-slate-500 max-w-xl mx-auto leading-relaxed">
            Extract structured timelines from any text. Every event traced to its source.
            Contradictions surfaced automatically.
          </p>
        </div>

        {/* Demo card */}
        <div className="mb-12">
          {demoLoading ? (
            <div className="bg-white border border-slate-200 rounded-xl p-8 flex items-center justify-center">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600" />
            </div>
          ) : demo ? (
            <div className="bg-white border border-slate-200 rounded-xl p-8 shadow-sm">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs font-semibold text-green-600 bg-green-50 px-2 py-0.5 rounded-full uppercase tracking-wide">
                      Ready
                    </span>
                    <span className="text-xs text-slate-400">Pre-extracted demo</span>
                  </div>
                  <h2 className="text-2xl font-bold text-slate-900">{demo.title}</h2>
                </div>
                <span className="text-4xl select-none">📖</span>
              </div>
              <div className="flex gap-6 mb-6 text-sm text-slate-500">
                <span><strong className="text-slate-800">{demo.events}</strong> events</span>
                <span><strong className="text-slate-800">{demo.entities}</strong> entities</span>
                <span><strong className="text-slate-800">{demo.relationships}</strong> relationships</span>
              </div>
              <button
                onClick={() => onNavigate(demo.id)}
                className="w-full py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg transition-colors text-sm"
              >
                Explore Demo →
              </button>
            </div>
          ) : (
            <div className="bg-white border border-slate-200 rounded-xl p-8 text-center">
              <p className="text-slate-500 text-sm">No demo document loaded. Upload a document below.</p>
            </div>
          )}
        </div>

        {/* Divider */}
        <div className="flex items-center gap-4 mb-8">
          <div className="flex-1 h-px bg-slate-200" />
          <span className="text-sm text-slate-400 shrink-0">or upload your own document</span>
          <div className="flex-1 h-px bg-slate-200" />
        </div>

        {/* Upload zone */}
        {!isProcessing && phase.name !== 'error' && (
          <div
            onDragOver={e => { e.preventDefault(); setIsDragOver(true); }}
            onDragLeave={() => setIsDragOver(false)}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            className={`border-2 border-dashed rounded-xl p-12 text-center cursor-pointer transition-colors mb-12 ${
              isDragOver
                ? 'border-blue-400 bg-blue-50'
                : 'border-slate-300 bg-white hover:border-slate-400 hover:bg-slate-50'
            }`}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept=".txt,.pdf"
              className="hidden"
              onChange={handleFileInput}
            />
            <div className="text-4xl mb-3 select-none">📄</div>
            <p className="text-slate-700 font-medium mb-1">Drag & drop a file here</p>
            <p className="text-sm text-slate-400">.txt or .pdf · up to 50 MB · click to browse</p>
          </div>
        )}

        {/* Processing state */}
        {isProcessing && (
          <div className="bg-white border border-slate-200 rounded-xl p-8 mb-12">
            <div className="flex items-center gap-4 mb-4">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 shrink-0" />
              <div className="flex-1">
                <p className="font-semibold text-slate-800">
                  {phase.name === 'uploading' && 'Uploading document...'}
                  {phase.name === 'chunking' && 'Splitting into sections...'}
                  {phase.name === 'extracting' && 'Extracting events with AI...'}
                </p>
                {phase.name === 'extracting' && (
                  <p className="text-sm text-slate-500 mt-0.5">
                    {phase.totalChunks > 0
                      ? `Processing chunk ${phase.currentChunk} of ${phase.totalChunks}`
                      : 'Starting extraction pipeline...'}
                  </p>
                )}
              </div>
              {phase.name === 'extracting' && phase.elapsedSec > 0 && (
                <span className="text-sm text-slate-400">
                  {Math.floor(phase.elapsedSec / 60)}:{(phase.elapsedSec % 60).toString().padStart(2, '0')}
                </span>
              )}
            </div>

            {/* Progress bar for extraction */}
            {phase.name === 'extracting' && phase.totalChunks > 0 && (
              <div className="mb-4">
                <div className="h-2 bg-slate-100 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-blue-500 rounded-full transition-all duration-300"
                    style={{ width: `${phase.percent}%` }}
                  />
                </div>
                <div className="flex justify-between mt-1 text-xs text-slate-500">
                  <span>{phase.percent}% complete</span>
                  <span>{phase.totalChunks - phase.currentChunk} chunks remaining</span>
                </div>
              </div>
            )}

            {/* Extraction stats */}
            {phase.name === 'extracting' && (phase.events > 0 || phase.entities > 0) && (
              <div className="flex gap-6 mb-4 text-sm">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-blue-500"></span>
                  <span className="text-slate-600"><strong className="text-slate-800">{phase.events}</strong> events</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-green-500"></span>
                  <span className="text-slate-600"><strong className="text-slate-800">{phase.entities}</strong> entities</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-purple-500"></span>
                  <span className="text-slate-600"><strong className="text-slate-800">{phase.relationships}</strong> relationships</span>
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
                className="w-full py-2 border border-blue-300 text-blue-700 text-sm font-medium rounded-lg hover:bg-blue-50 transition-colors"
              >
                View partial results now →
              </button>
            )}
          </div>
        )}

        {/* Error state */}
        {phase.name === 'error' && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-6 mb-12 flex items-start gap-3">
            <span className="text-red-500 text-lg shrink-0">✕</span>
            <div className="flex-1">
              <p className="text-red-700 font-medium text-sm">{phase.message}</p>
              <button
                onClick={() => setPhase({ name: 'idle' })}
                className="mt-2 text-xs text-red-600 hover:text-red-800 underline"
              >
                Try again
              </button>
            </div>
          </div>
        )}

        {/* Feature list */}
        <div className="grid grid-cols-2 gap-4">
          {FEATURES.map(f => (
            <div key={f.label} className="flex items-start gap-3 bg-white border border-slate-200 rounded-lg p-4">
              <span className="text-blue-500 text-lg shrink-0 mt-0.5">{f.icon}</span>
              <div>
                <p className="text-sm font-semibold text-slate-800">{f.label}</p>
                <p className="text-xs text-slate-500 mt-0.5">{f.detail}</p>
              </div>
            </div>
          ))}
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-200 bg-white">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between text-xs text-slate-400">
          <span>Sikta — Document Timeline Intelligence</span>
          <span>Demo: <em>Pride and Prejudice</em>, Jane Austen (public domain)</span>
        </div>
      </footer>
    </div>
  );
}
