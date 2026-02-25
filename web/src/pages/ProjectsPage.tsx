import { useState, useEffect } from 'react';
import { listProjects, createProject, Project } from '../api/projects';

interface Props {
  onSelectProject: (projectId: string) => void;
}

export default function ProjectsPage({ onSelectProject }: Props) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showNewProject, setShowNewProject] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [newDescription, setNewDescription] = useState('');
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    loadProjects();
  }, []);

  async function loadProjects() {
    try {
      setLoading(true);
      const data = await listProjects();
      setProjects(data);
      setError(null);
    } catch (err) {
      setError('Failed to load projects');
    } finally {
      setLoading(false);
    }
  }

  async function handleCreateProject(e: React.FormEvent) {
    e.preventDefault();
    if (!newTitle.trim()) return;

    try {
      setCreating(true);
      const project = await createProject(newTitle.trim(), newDescription.trim());
      setProjects([project, ...projects]);
      setShowNewProject(false);
      setNewTitle('');
      setNewDescription('');
    } catch (err) {
      setError('Failed to create project');
    } finally {
      setCreating(false);
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="text-slate-500">Loading projects...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🔍</span>
            <h1 className="text-xl font-semibold text-slate-900">Sikta</h1>
          </div>
          <button
            onClick={() => setShowNewProject(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium"
          >
            New Project
          </button>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-6xl mx-auto px-6 py-8">
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            {error}
          </div>
        )}

        {/* New Project Form */}
        {showNewProject && (
          <div className="mb-6 p-6 bg-white rounded-lg border border-slate-200 shadow-sm">
            <h2 className="text-lg font-semibold mb-4">Create New Project</h2>
            <form onSubmit={handleCreateProject}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Title
                </label>
                <input
                  type="text"
                  value={newTitle}
                  onChange={(e) => setNewTitle(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Project title"
                  required
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Description (optional)
                </label>
                <textarea
                  value={newDescription}
                  onChange={(e) => setNewDescription(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Brief description"
                  rows={3}
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="submit"
                  disabled={creating}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
                >
                  {creating ? 'Creating...' : 'Create Project'}
                </button>
                <button
                  type="button"
                  onClick={() => setShowNewProject(false)}
                  className="px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 text-sm font-medium"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Projects list */}
        {projects.length === 0 ? (
          <div className="text-center py-16">
            <div className="text-slate-400 text-5xl mb-4">📁</div>
            <h2 className="text-xl font-semibold text-slate-700 mb-2">No projects yet</h2>
            <p className="text-slate-500 mb-6">Create your first project to start extracting documents.</p>
            <button
              onClick={() => setShowNewProject(true)}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium"
            >
              Create Project
            </button>
          </div>
        ) : (
          <div className="grid gap-4">
            {projects.map((project) => (
              <div
                key={project.id}
                onClick={() => onSelectProject(project.id)}
                className="p-6 bg-white rounded-lg border border-slate-200 shadow-sm hover:shadow-md hover:border-slate-300 cursor-pointer transition-all"
              >
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="text-lg font-semibold text-slate-900">{project.title}</h3>
                    {project.description && (
                      <p className="text-slate-500 mt-1">{project.description}</p>
                    )}
                  </div>
                  {project.stats && (
                    <div className="flex gap-4 text-sm">
                      <div className="text-center">
                        <div className="font-semibold text-slate-900">{project.stats.doc_count}</div>
                        <div className="text-slate-500">Docs</div>
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
                <div className="mt-4 text-xs text-slate-400">
                  Created {new Date(project.created_at).toLocaleDateString()}
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
