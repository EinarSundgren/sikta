import { useState } from 'react';
import LandingPage from './pages/LandingPage';
import TimelineView from './pages/TimelineView';

type AppView = { screen: 'landing' } | { screen: 'timeline'; docId: string };

export default function App() {
  const [view, setView] = useState<AppView>({ screen: 'landing' });

  if (view.screen === 'timeline') {
    return (
      <TimelineView
        docId={view.docId}
        onNavigateHome={() => setView({ screen: 'landing' })}
      />
    );
  }

  return (
    <LandingPage
      onNavigate={docId => setView({ screen: 'timeline', docId })}
    />
  );
}
