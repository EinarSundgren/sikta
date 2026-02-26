import { useState, useEffect } from 'react';
import {
  getProject,
  getProjectDocuments,
  getProjectGraph,
  addDocumentToProject,
  runPostProcessing,
  listDocuments,
  uploadDocument,
  extractDocument,
  Project,
  Document,
  Graph,
} from '../api/projects';
import SwimlaneTimeline from '../components/swimlane/SwimlaneTimeline';

interface Props {
  projectId: string;
  onNavigateBack: () => void;
}

type Tab = 'swimlane' | 'documents' | 'entities' | 'events' | 'relationships';

export default function ProjectView({ projectId, onNavigateBack }: Props) {
  const [project, setProject] = useState<Project | null>(null);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [graph, setGraph] = useState<Graph | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<Tab>('swimlane');
  const [showAddDoc, setShowAddDoc] = useState(false);
  const [availableDocs, setAvailableDocs] = useState<Document[]>([]);
  const [processing, setProcessing] = useState(false);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
    loadProjectData();
  }, [projectId]);

  async function loadProjectData() {
    try {
      setLoading(true);
      const [projectData, docs, graphData] = await Promise.all([
        getProject(projectId),
        getProjectDocuments(projectId),
        getProjectGraph(projectId),
      ]);
      setProject(projectData);
      setDocuments(docs);
      setGraph(graphData);
      setError(null);
    } catch (err) {
      setError('Failed to load project');
    } finally {
      setLoading(false);
    }
  }

  async function loadAvailableDocs() {
    try {
      const allDocs = await listDocuments();
      const projectDocIds = new Set(documents.map((d) => d.id));
      setAvailableDocs(allDocs.filter((d) => !projectDocIds.has(d.id)));
    } catch (err) {
      console.error('Failed to load available documents');
    }
  }

  async function handleAddDocument(docId: string) {
    try {
      await addDocumentToProject(projectId, docId);
      await loadProjectData();
      setShowAddDoc(false);
    } catch (err) {
      setError('Failed to add document');
    }
  }

  async function handleUploadAndAdd(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      setUploading(true);
      // Upload document
      const doc = await uploadDocument(file);
      // Wait a bit for processing
      await new Promise((r) => setTimeout(r, 2000));
      // Trigger extraction
      await extractDocument(doc.id);
      // Add to project
      await addDocumentToProject(projectId, doc.id);
      await loadProjectData();
    } catch (err) {
      setError('Failed to upload document');
    } finally {
      setUploading(false);
    }
  }

  async function handleRunPostProcessing() {
    try {
      setProcessing(true);
      const result = await runPostProcessing(projectId);
      console.log('Post-processing result:', result);
      await loadProjectData();
    } catch (err) {
      setError('Failed to run post-processing');
    } finally {
      setProcessing(false);
    }
  }

  // Filter nodes by type
  const entities = graph?.nodes.filter((n) =>
    ['person', 'organization', 'place', 'object'].includes(n.node_type)
  ) || [];
  const events = graph?.nodes.filter((n) => n.node_type === 'event') || [];
  const relationships = graph?.edges.filter((e) =>
    ['same_as', 'contradicts'].includes(e.edge_type)
  ) || [];

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="text-slate-500">Loading project...</div>
      </div>
    );
  }

  if (!project) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="text-red-500">Project not found</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-6xl mx-auto px-6 py-4">
          <button
            onClick={onNavigateBack}
            className="text-slate-500 hover:text-slate-700 text-sm mb-2 flex items-center gap-1"
          >
            ← Back to projects
          </button>
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-2xl font-semibold text-slate-900">{project.title}</h1>
              {project.description && (
                <p className="text-slate-500 mt-1">{project.description}</p>
              )}
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => {
                  loadAvailableDocs();
                  setShowAddDoc(true);
                }}
                className="px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 text-sm font-medium"
              >
                Add Document
              </button>
              <button
                onClick={handleRunPostProcessing}
                disabled={processing || documents.length < 2}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
              >
                {processing ? 'Processing...' : 'Run Analysis'}
              </button>
            </div>
          </div>

          {/* Stats */}
          {project.stats && (
            <div className="flex gap-6 mt-4 text-sm">
              <div className="text-center">
                <div className="font-semibold text-slate-900">{project.stats.doc_count}</div>
                <div className="text-slate-500">Documents</div>
              </div>
              <div className="text-center">
                <div className="font-semibold text-slate-900">{project.stats.node_count}</div>
                <div className="text-slate-500">Nodes</div>
              </div>
              <div className="text-center">
                <div className="font-semibold text-slate-900">{project.stats.edge_count}</div>
                <div className="text-slate-500">Edges</div>
              </div>
            </div>
          )}
        </div>
      </header>

      {/* Tabs */}
      <div className="bg-white border-b border-slate-200">
        <div className="max-w-6xl mx-auto px-6">
          <div className="flex gap-1">
            {[
              { id: 'swimlane' as Tab, label: 'Swimlane', count: 0 },
              { id: 'documents' as Tab, label: 'Documents', count: documents.length },
              { id: 'entities' as Tab, label: 'Entities', count: entities.length },
              { id: 'events' as Tab, label: 'Events', count: events.length },
              { id: 'relationships' as Tab, label: 'Analysis', count: relationships.length },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === tab.id
                    ? 'border-blue-600 text-blue-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700'
                }`}
              >
                {tab.label}
                {tab.count > 0 && (
                  <span className="ml-2 px-2 py-0.5 bg-slate-100 rounded-full text-xs">
                    {tab.count}
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Main content */}
      <main className="max-w-6xl mx-auto px-6 py-6">
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            {error}
          </div>
        )}

        {/* Add Document Modal */}
        {showAddDoc && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[80vh] overflow-hidden">
              <div className="p-6 border-b border-slate-200">
                <h2 className="text-lg font-semibold">Add Document</h2>
              </div>
              <div className="p-6">
                {/* Upload new */}
                <div className="mb-6">
                  <label className="block text-sm font-medium text-slate-700 mb-2">
                    Upload new document
                  </label>
                  <input
                    type="file"
                    accept=".txt,.pdf"
                    onChange={handleUploadAndAdd}
                    disabled={uploading}
                    className="block w-full text-sm text-slate-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                  />
                  {uploading && (
                    <p className="mt-2 text-sm text-slate-500">Uploading and extracting...</p>
                  )}
                </div>

                {/* Or select existing */}
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">
                    Or select existing document
                  </label>
                  {availableDocs.length === 0 ? (
                    <p className="text-slate-500 text-sm">No available documents</p>
                  ) : (
                    <div className="max-h-48 overflow-y-auto space-y-2">
                      {availableDocs.map((doc) => (
                        <button
                          key={doc.id}
                          onClick={() => handleAddDocument(doc.id)}
                          className="w-full text-left p-3 bg-slate-50 rounded-lg hover:bg-slate-100"
                        >
                          <div className="font-medium text-slate-900">{doc.title}</div>
                          <div className="text-sm text-slate-500">{doc.filename}</div>
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              </div>
              <div className="p-6 border-t border-slate-200 flex justify-end">
                <button
                  onClick={() => setShowAddDoc(false)}
                  className="px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 text-sm font-medium"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Tab content */}
        {activeTab === 'swimlane' && (
          <div className="relative" style={{ minHeight: 500 }}>
            <SwimlaneTimeline
              nodes={graph?.nodes || []}
              edges={graph?.edges || []}
              documents={documents}
              onEventClick={(event) => console.log('Event clicked:', event)}
            />
          </div>
        )}

        {activeTab === 'documents' && (
          <div className="space-y-3">
            {documents.length === 0 ? (
              <div className="text-center py-12 text-slate-500">
                No documents yet. Add a document to start.
              </div>
            ) : (
              documents.map((doc) => (
                <div
                  key={doc.id}
                  className="p-4 bg-white rounded-lg border border-slate-200 shadow-sm"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-medium text-slate-900">{doc.title}</h3>
                      <p className="text-sm text-slate-500">{doc.filename}</p>
                    </div>
                    <span
                      className={`px-2 py-1 text-xs rounded-full ${
                        doc.upload_status === 'ready'
                          ? 'bg-green-100 text-green-700'
                          : doc.upload_status === 'processing'
                          ? 'bg-yellow-100 text-yellow-700'
                          : 'bg-red-100 text-red-700'
                      }`}
                    >
                      {doc.upload_status}
                    </span>
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {activeTab === 'entities' && (
          <div className="space-y-3">
            {entities.length === 0 ? (
              <div className="text-center py-12 text-slate-500">
                No entities extracted yet. Add documents and run extraction.
              </div>
            ) : (
              entities.map((entity) => (
                <div
                  key={entity.id}
                  className="p-4 bg-white rounded-lg border border-slate-200 shadow-sm flex items-start gap-3"
                >
                  <div
                    className={`w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-medium ${
                      entity.node_type === 'person'
                        ? 'bg-blue-500'
                        : entity.node_type === 'organization'
                        ? 'bg-purple-500'
                        : 'bg-green-500'
                    }`}
                  >
                    {entity.node_type[0].toUpperCase()}
                  </div>
                  <div>
                    <h3 className="font-medium text-slate-900">{entity.label}</h3>
                    <p className="text-sm text-slate-500 capitalize">{entity.node_type}</p>
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {activeTab === 'events' && (
          <div className="space-y-3">
            {events.length === 0 ? (
              <div className="text-center py-12 text-slate-500">
                No events extracted yet. Add documents and run extraction.
              </div>
            ) : (
              events.map((event) => (
                <div
                  key={event.id}
                  className="p-4 bg-white rounded-lg border border-slate-200 shadow-sm"
                >
                  <h3 className="font-medium text-slate-900">{event.label}</h3>
                  {event.properties && 'description' in event.properties && (
                    <p className="text-sm text-slate-500 mt-1">
                      {event.properties.description as string}
                    </p>
                  )}
                </div>
              ))
            )}
          </div>
        )}

        {activeTab === 'relationships' && (
          <div className="space-y-3">
            {relationships.length === 0 ? (
              <div className="text-center py-12 text-slate-500">
                <p className="mb-2">No cross-document analysis yet.</p>
                <p className="text-sm">Add at least 2 documents and click "Run Analysis".</p>
              </div>
            ) : (
              relationships.map((rel) => (
                <div
                  key={rel.id}
                  className={`p-4 rounded-lg border shadow-sm ${
                    rel.edge_type === 'contradicts'
                      ? 'bg-red-50 border-red-200'
                      : 'bg-blue-50 border-blue-200'
                  }`}
                >
                  <div className="flex items-center gap-2 mb-2">
                    <span
                      className={`px-2 py-0.5 text-xs rounded-full font-medium ${
                        rel.edge_type === 'contradicts'
                          ? 'bg-red-100 text-red-700'
                          : 'bg-blue-100 text-blue-700'
                      }`}
                    >
                      {rel.edge_type}
                    </span>
                  </div>
                  {rel.properties && 'description' in rel.properties && (
                    <p className="text-sm text-slate-700">
                      {rel.properties.description as string}
                    </p>
                  )}
                  {rel.properties && 'canonical' in rel.properties && (
                    <p className="text-sm text-slate-700">
                      <strong>{rel.properties.canonical as string}</strong> ={' '}
                      {rel.properties.alias as string}
                    </p>
                  )}
                </div>
              ))
            )}
          </div>
        )}
      </main>
    </div>
  );
}
