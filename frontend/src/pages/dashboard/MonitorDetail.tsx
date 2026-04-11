import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { monitorsApi } from '../../api/monitors'
import { userApi } from '../../api/user'
import { DashboardLayout } from '../../components/layout/DashboardLayout'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { Card } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts'
import { ArrowLeft, Bell, CheckCircle, PauseCircle, PlayCircle, X } from 'lucide-react'
import { format } from '../../lib/format'
import { usePageTitle } from '../../lib/usePageTitle'

export function MonitorDetailPage() {
  usePageTitle('Monitor detail')
  const { id } = useParams<{ id: string }>()

  const queryClient = useQueryClient()

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

  const toggleMutation = useMutation({
    mutationFn: () => monitorsApi.update(detail!.monitor.id, { is_active: !detail!.monitor.is_active }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['monitors', id] }),
  })

  if (isLoading || !detail) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64 text-gray-400 text-sm">Loading…</div>
      </DashboardLayout>
    )
  }

  const { monitor, uptime, incidents } = detail

  const chartData = graph.map(p => ({
    time: format.time(p.timestamp),
    ms: p.is_up ? p.response_time_ms : null,
  }))

  const effectiveStatus = monitor.is_active ? monitor.status : 'paused'

  return (
    <DashboardLayout>
      <div className="max-w-3xl mx-auto">
        <Link to="/dashboard" className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 mb-6">
          <ArrowLeft size={14} /> Back to dashboard
        </Link>

        {/* Header */}
        <div className="flex items-start justify-between mb-6">
          <div>
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-2xl font-semibold text-gray-900">{monitor.name}</h1>
              <StatusBadge status={effectiveStatus} />
            </div>
            <a href={monitor.url} target="_blank" rel="noreferrer" className="text-sm text-indigo-500 hover:underline">
              {monitor.url} ↗
            </a>
            {monitor.last_checked_at && (
              <p className="text-xs text-gray-400 mt-1">
                Last checked {format.timeAgo(monitor.last_checked_at)}
              </p>
            )}
          </div>
          <button
            onClick={() => toggleMutation.mutate()}
            disabled={toggleMutation.isPending}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium border transition-colors disabled:opacity-50
              hover:bg-gray-50 border-gray-200 text-gray-600"
            title={monitor.is_active ? 'Pause monitoring' : 'Resume monitoring'}
          >
            {monitor.is_active
              ? <><PauseCircle size={15} className="text-amber-500" /> Pause</>
              : <><PlayCircle size={15} className="text-green-500" /> Resume</>
            }
          </button>
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
        <Card className="p-6 mb-6">
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
                      {i.resolved_at
                        ? ` · Resolved ${format.datetime(i.resolved_at)}`
                        : ` · ${format.timeAgo(i.started_at)}`}
                    </p>
                  </div>
                  <span className="text-xs text-gray-400">
                    {i.resolved_at
                      ? format.duration(i.started_at, i.resolved_at)
                      : `${format.ongoingDuration(i.started_at)} ongoing`}
                  </span>
                </div>
              ))}
            </div>
          )}
        </Card>

        {/* Alert subscriptions */}
        <SubscribeSection monitorId={monitor.id} />
      </div>
    </DashboardLayout>
  )
}

function SubscribeSection({ monitorId }: { monitorId: string }) {
  const queryClient = useQueryClient()
  const [selectedChannelId, setSelectedChannelId] = useState('')

  const { data: allChannels = [] } = useQuery({
    queryKey: ['alert-channels'],
    queryFn: () => userApi.listAlertChannels().then(r => r.data ?? []),
  })

  const { data: subscribed = [] } = useQuery({
    queryKey: ['monitor-subscriptions', monitorId],
    queryFn: () => userApi.listMonitorSubscriptions(monitorId).then(r => r.data ?? []),
  })

  const subscribedIds = new Set(subscribed.map(ch => ch.id))
  const available = allChannels.filter(ch => !subscribedIds.has(ch.id))

  const subscribeMutation = useMutation({
    mutationFn: () => userApi.subscribeMonitor(monitorId, selectedChannelId),
    onSuccess: () => {
      setSelectedChannelId('')
      queryClient.invalidateQueries({ queryKey: ['monitor-subscriptions', monitorId] })
    },
  })

  const unsubscribeMutation = useMutation({
    mutationFn: (channelId: string) => userApi.unsubscribeMonitor(monitorId, channelId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['monitor-subscriptions', monitorId] }),
  })

  return (
    <Card className="p-6">
      <div className="flex items-center gap-2 mb-4">
        <Bell size={16} className="text-indigo-500" />
        <h2 className="text-sm font-medium text-gray-700">Alert channels</h2>
      </div>

      {/* Subscribed channels */}
      {subscribed.length > 0 && (
        <div className="mb-4 space-y-2">
          {subscribed.map(ch => (
            <div key={ch.id} className="flex items-center gap-2 text-sm text-gray-700 bg-green-50 border border-green-100 rounded-lg px-3 py-2">
              <CheckCircle size={14} className="text-green-500 shrink-0" />
              <span>{ch.name || ch.config.email || ch.type}</span>
              <span className="text-xs text-gray-400 ml-auto mr-2 capitalize">{ch.type} · subscribed</span>
              <button
                onClick={() => unsubscribeMutation.mutate(ch.id)}
                disabled={unsubscribeMutation.isPending}
                className="text-gray-400 hover:text-red-500 transition-colors disabled:opacity-50"
                title="Remove subscription"
              >
                <X size={14} />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Add subscription */}
      {allChannels.length === 0 ? (
        <p className="text-sm text-gray-400">
          No alert channels yet.{' '}
          <Link to="/dashboard/alert-channels" className="text-indigo-500 hover:underline">Create one first.</Link>
        </p>
      ) : available.length === 0 ? (
        <p className="text-sm text-gray-400">All your alert channels are already subscribed.</p>
      ) : (
        <div className="flex gap-3 items-end">
          <div className="flex-1">
            <label className="text-xs font-medium text-gray-500 block mb-1">Add channel</label>
            <select
              value={selectedChannelId}
              onChange={e => setSelectedChannelId(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">Select a channel…</option>
              {available.map(ch => (
                <option key={ch.id} value={ch.id}>
                  {ch.name || ch.config.email || ch.type} ({ch.type})
                </option>
              ))}
            </select>
          </div>
          <Button
            onClick={() => subscribeMutation.mutate()}
            loading={subscribeMutation.isPending}
            disabled={!selectedChannelId}
          >
            Subscribe
          </Button>
        </div>
      )}
    </Card>
  )
}
