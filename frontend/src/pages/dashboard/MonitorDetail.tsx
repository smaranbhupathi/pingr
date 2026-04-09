import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { monitorsApi } from '../../api/monitors'
import { Navbar } from '../../components/ui/Navbar'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { Card } from '../../components/ui/Card'
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts'
import { ArrowLeft } from 'lucide-react'
import { format } from '../../lib/format'

export function MonitorDetailPage() {
  const { id } = useParams<{ id: string }>()

  const { data: detail, isLoading } = useQuery({
    queryKey: ['monitors', id],
    queryFn: () => monitorsApi.get(id!).then(r => r.data),
    refetchInterval: 30_000,
  })

  const { data: graph = [] } = useQuery({
    queryKey: ['monitors', id, 'graph'],
    queryFn: () => monitorsApi.graph(id!).then(r => r.data),
    refetchInterval: 60_000,
  })

  if (isLoading || !detail) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Navbar />
        <div className="flex items-center justify-center h-64 text-gray-400">Loading...</div>
      </div>
    )
  }

  const { monitor, uptime, incidents } = detail

  const chartData = graph.map(p => ({
    time: format.time(p.timestamp),
    ms: p.is_up ? p.response_time_ms : null,
  }))

  return (
    <div className="min-h-screen bg-gray-50">
      <Navbar />

      <main className="max-w-5xl mx-auto px-6 py-8">
        <Link to="/dashboard" className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 mb-6">
          <ArrowLeft size={14} /> Back to dashboard
        </Link>

        {/* Header */}
        <div className="flex items-start justify-between mb-6">
          <div>
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-2xl font-semibold text-gray-900">{monitor.name}</h1>
              <StatusBadge status={monitor.status} />
            </div>
            <a href={monitor.url} target="_blank" rel="noreferrer" className="text-sm text-indigo-500 hover:underline">
              {monitor.url} ↗
            </a>
          </div>
        </div>

        {/* Uptime stats */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          {[
            { label: '24h uptime', value: uptime.last_24h },
            { label: '7d uptime', value: uptime.last_7d },
            { label: '30d uptime', value: uptime.last_30d },
            { label: '90d uptime', value: uptime.last_90d },
          ].map(({ label, value }) => (
            <Card key={label} className="p-4 text-center">
              <p className="text-xs text-gray-500 mb-1">{label}</p>
              <p className={`text-2xl font-semibold ${value >= 99 ? 'text-green-600' : value >= 95 ? 'text-yellow-500' : 'text-red-500'}`}>
                {value.toFixed(2)}%
              </p>
            </Card>
          ))}
        </div>

        {/* Response time chart */}
        <Card className="p-6 mb-6">
          <h2 className="text-sm font-medium text-gray-700 mb-4">Response time (last 24h)</h2>
          {chartData.length === 0 ? (
            <div className="h-40 flex items-center justify-center text-gray-400 text-sm">
              No data yet — check back after the first monitor run
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={180}>
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient id="grad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#6366f1" stopOpacity={0.15} />
                    <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="time" tick={{ fontSize: 11 }} tickLine={false} />
                <YAxis tick={{ fontSize: 11 }} tickLine={false} unit="ms" />
                <Tooltip formatter={(v) => [`${v}ms`, 'Response time']} />
                <Area type="monotone" dataKey="ms" stroke="#6366f1" fill="url(#grad)" strokeWidth={2} dot={false} connectNulls={false} />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </Card>

        {/* Incidents */}
        <Card className="p-6">
          <h2 className="text-sm font-medium text-gray-700 mb-4">Incident history</h2>
          {incidents.length === 0 ? (
            <p className="text-sm text-gray-400 text-center py-4">No incidents — looking good! 🎉</p>
          ) : (
            <div className="divide-y divide-gray-100">
              {incidents.map(i => (
                <div key={i.id} className="py-3 flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-900">
                      {i.resolved_at ? '🟢 Resolved' : '🔴 Ongoing'}
                    </p>
                    <p className="text-xs text-gray-400 mt-0.5">
                      Started {format.datetime(i.started_at)}
                      {i.resolved_at && ` · Resolved ${format.datetime(i.resolved_at)}`}
                    </p>
                  </div>
                  {i.resolved_at && (
                    <span className="text-xs text-gray-400">
                      {format.duration(i.started_at, i.resolved_at)}
                    </span>
                  )}
                </div>
              ))}
            </div>
          )}
        </Card>
      </main>
    </div>
  )
}
