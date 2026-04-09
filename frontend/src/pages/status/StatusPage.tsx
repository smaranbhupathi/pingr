import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { monitorsApi, type MonitorDetail } from '../../api/monitors'
import { format } from '../../lib/format'

export function StatusPage() {
  const { username } = useParams<{ username: string }>()

  const { data: monitors = [], isLoading, isError } = useQuery({
    queryKey: ['status', username],
    queryFn: () => monitorsApi.statusPage(username!).then(r => r.data),
    refetchInterval: 60_000,
  })

  const allUp = monitors.every(m => m.monitor.status !== 'down')

  if (isLoading) {
    return <StatusShell username={username!}><p className="text-gray-400 text-center py-12">Loading...</p></StatusShell>
  }

  if (isError) {
    return <StatusShell username={username!}><p className="text-red-400 text-center py-12">Failed to load status page.</p></StatusShell>
  }

  return (
    <StatusShell username={username!}>
      {/* Overall status banner */}
      <div className={`rounded-xl p-5 mb-8 flex items-center gap-3 ${allUp ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
        <span className="text-2xl">{allUp ? '✅' : '🔴'}</span>
        <div>
          <p className={`font-semibold ${allUp ? 'text-green-800' : 'text-red-800'}`}>
            {allUp ? 'All systems operational' : 'Some systems are down'}
          </p>
          <p className="text-sm text-gray-500 mt-0.5">
            Last updated {format.time(new Date().toISOString())}
          </p>
        </div>
      </div>

      {/* Monitor list */}
      {monitors.length === 0 ? (
        <p className="text-gray-400 text-center py-12">No monitors configured yet.</p>
      ) : (
        <div className="space-y-3">
          {monitors.map(detail => (
            <MonitorStatusRow key={detail.monitor.id} detail={detail} />
          ))}
        </div>
      )}
    </StatusShell>
  )
}

function MonitorStatusRow({ detail }: { detail: MonitorDetail }) {
  const { monitor, uptime, incidents } = detail
  const openIncident = incidents.find(i => !i.resolved_at)
  const isDown = monitor.status === 'down'

  return (
    <div className={`bg-white border rounded-xl p-5 ${isDown ? 'border-red-200' : 'border-gray-200'}`}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className={`w-2.5 h-2.5 rounded-full ${isDown ? 'bg-red-500' : 'bg-green-500'}`} />
          <div>
            <p className="font-medium text-gray-900">{monitor.name}</p>
            <p className="text-xs text-gray-400">{monitor.url}</p>
          </div>
        </div>
        <div className="text-right">
          <p className={`text-sm font-semibold ${isDown ? 'text-red-600' : 'text-green-600'}`}>
            {isDown ? 'Down' : 'Operational'}
          </p>
          <p className="text-xs text-gray-400">{uptime.last_30d.toFixed(2)}% uptime (30d)</p>
        </div>
      </div>

      {openIncident && (
        <div className="mt-3 pt-3 border-t border-red-100 text-sm text-red-600">
          Incident started {format.datetime(openIncident.started_at)}
        </div>
      )}
    </div>
  )
}

function StatusShell({ username, children }: { username: string; children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-2xl mx-auto px-6 py-12">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900">{username}</h1>
          <p className="text-sm text-gray-400 mt-1">Service Status</p>
        </div>
        {children}
        <p className="text-center text-xs text-gray-300 mt-12">
          Powered by{' '}
          <Link to="/" className="text-indigo-400 hover:underline">Pingr</Link>
        </p>
      </div>
    </div>
  )
}
