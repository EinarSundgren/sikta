import { useState } from 'react';
import LandingPage from './pages/LandingPage';
import TimelineView from './pages/TimelineView';
import ProjectsPage from './pages/ProjectsPage';
import ProjectView from './pages/ProjectView';

type AppView =
  | { screen: 'landing' }
  | { screen: 'timeline'; docId: string }
  | { screen: 'projects' }
  | { screen: 'project'; projectId: string };

export default function App() {
  const [view, setView] = useState<AppView>({ screen: 'landing' });

  // Project view
  if (view.screen === 'project') {
    return (
      <ProjectView
        projectId={view.projectId}
        onNavigateBack={() => setView({ screen: 'projects' })}
      />
    );
  }

  // Projects list
  if (view.screen === 'projects') {
    return (
      <ProjectsPage
        onSelectProject={(projectId) => setView({ screen: 'project', projectId })}
      />
    );
  }

  // Timeline view (document detail)
  if (view.screen === 'timeline') {
    return (
      <TimelineView
        docId={view.docId}
        onNavigateHome={() => setView({ screen: 'landing' })}
      />
    );
  }

  // Landing page (project-centric)
  return (
    <LandingPage
      onNavigateToProject={(projectId) => setView({ screen: 'project', projectId })}
    />
  );
}
