import { useEffect, useState } from 'react'

type HealthStatus = {
  status: string
  timestamp: string
}

export default function App() {
  const [health, setHealth] = useState<HealthStatus | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch('/health')
      .then(res => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json() as Promise<HealthStatus>
      })
      .then(data => setHealth(data))
      .catch(() => setError('Backend not reachable'))
  }, [])

  return (
    <div className="min-h-screen bg-gray-950 text-white flex flex-col items-center justify-center gap-3">
      <h1 className="text-5xl font-bold tracking-tight">Sikta</h1>
      <p className="text-gray-400 text-lg">Document Timeline Intelligence</p>

      <div className="mt-6 text-sm font-mono">
        {error && <span className="text-red-400">{error}</span>}
        {health && (
          <span className="text-green-400">
            Backend: {health.status} &middot; {health.timestamp}
          </span>
        )}
        {!health && !error && (
          <span className="text-gray-500">Connecting to backend...</span>
        )}
      </div>
    </div>
  )
}
